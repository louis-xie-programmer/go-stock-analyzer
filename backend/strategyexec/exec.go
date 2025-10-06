package strategyexec

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"go-stock-analyzer/backend/storage"

	"github.com/traefik/yaegi/interp"
	"github.com/traefik/yaegi/stdlib"
)

// 执行器配置
type ExecConfig struct {
	TotalTimeout     time.Duration // 总超时
	PerSymbolTimeout time.Duration // 每只股票执行超时，可二次控制
}

// 默认配置（可修改）
var DefaultExecConfig = ExecConfig{
	TotalTimeout:     30 * time.Second,
	PerSymbolTimeout: 800 * time.Millisecond,
}

// ExecuteStrategy 用用户 code 在 symbols 列表上执行 Match 函数。
// - code: 用户提供的源码字符串，必须定义 `func Match(symbol string, klines []map[string]interface{}) bool`
// - symbols: 如 ["sz000001", "sh600000"]
// - loadKlines: 由调用方提供加载函数 (symbol, days) -> []storage.KLine
func ExecuteStrategy(code string, symbols []string, klineDays int, loadKlines func(string, int) ([]storage.KLine, error), cfg ExecConfig) ([]string, error) {
	if cfg.TotalTimeout == 0 {
		cfg = DefaultExecConfig
	}
	// 创建解释器
	i := interp.New(interp.Options{})
	// 允许基础 stdlib（你可以选择禁用某些包，但为方便演示使用 stdlib）
	i.Use(stdlib.Symbols)

	// Prepare a wrapper source that will expose Match as global symbol.
	// We will evaluate the user's code directly.
	if _, err := i.Eval(code); err != nil {
		return nil, fmt.Errorf("compile error: %w", err)
	}

	// 获取 Match 符号
	v, err := i.Eval("Match")
	if err != nil {
		return nil, fmt.Errorf("Match function not found in code: %w", err)
	}
	matchVal := v.Interface()
	matchFunc := reflect.ValueOf(matchVal)
	if matchFunc.Kind() != reflect.Func {
		return nil, fmt.Errorf("Match is not a function")
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.TotalTimeout)
	defer cancel()

	matched := []string{}
	start := time.Now()

	for _, sym := range symbols {
		select {
		case <-ctx.Done():
			return matched, fmt.Errorf("total timeout reached")
		default:
		}

		// load klines from caller-provided loader
		klines, err := loadKlines(sym, klineDays)
		if err != nil {
			// skip on load error
			continue
		}
		// convert []storage.KLine -> []map[string]interface{}
		arg := make([]map[string]interface{}, 0, len(klines))
		for _, k := range klines {
			m := map[string]interface{}{
				"Date":  k.Date,
				"Open":  k.Open,
				"High":  k.High,
				"Low":   k.Low,
				"Close": k.Close,
				"Volume": k.Volume,
			}
			arg = append(arg, m)
		}

		// prepare call with a per-symbol timeout
		ch := make(chan bool, 1)
		errCh := make(chan error, 1)
		go func(sym string, arg []map[string]interface{}) {
			defer func() {
				// recover from panics in user code
				if r := recover(); r != nil {
					errCh <- fmt.Errorf("panic in user code: %v", r)
				}
			}()
			// Call matchFunc(symbol, arg)
			in := []reflect.Value{reflect.ValueOf(sym), reflect.ValueOf(arg)}
			out := matchFunc.Call(in)
			if len(out) == 0 {
				errCh <- fmt.Errorf("Match must return bool")
				return
			}
			ok, okCast := out[0].Interface().(bool)
			if !okCast {
				errCh <- fmt.Errorf("Match return is not bool")
				return
			}
			ch <- ok
		}(sym, arg)

		select {
		case ok := <-ch:
			if ok {
				matched = append(matched, sym)
			}
		case e := <-errCh:
			// log and skip symbol
			_ = e // can log
			continue
		case <-time.After(cfg.PerSymbolTimeout):
			// per-symbol timeout, skip
			continue
		case <-ctx.Done():
			return matched, fmt.Errorf("total timeout reached")
		}
	}

	_ = start // could measure duration
	return matched, nil
}
