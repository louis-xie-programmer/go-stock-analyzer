package strategy

import (
	"fmt"

	"go-stock-analyzer/backend/config"
	"go-stock-analyzer/backend/storage"
)
type Strategy interface {
	Name() string
	Match(code string, klines []storage.KLine) bool
}

func strategyFromConfig(sc config.StrategyConfig) (Strategy, error) {
	switch sc.Name {
	case "MA":
		ma := 20
		hd := 3
		if v, ok := sc.Params["ma"]; ok {
			switch t := v.(type) {
			case int:
				ma = t
			case float64:
				ma = int(t)
			}
		}
		if v, ok := sc.Params["hold_days"]; ok {
			switch t := v.(type) {
			case int:
				hd = t
			case float64:
				hd = int(t)
			}
		}
		return NewMAStrategy(ma, hd), nil
	case "MACD":
		return NewMACDStrategy(), nil
	case "Composite":
		hd := 3
		if v, ok := sc.Params["hold_days"]; ok {
			switch t := v.(type) {
			case int:
				hd = t
			case float64:
				hd = int(t)
			}
		}
		return NewCompositeStrategy(hd), nil
	case "DSL":
		expr := ""
		if v, ok := sc.Params["expr"]; ok {
			if s, ok2 := v.(string); ok2 {
				expr = s
			}
		}
		return NewDSLStrategy(expr), nil
	default:
		return nil, fmt.Errorf("unknown strategy: %s", sc.Name)
	}
}

func RunAll(stocks []string) {
	for _, sc := range config.Cfg.Strategies {
		if !sc.Enabled {
			continue
		}
		strat, err := strategyFromConfig(sc)
		if err != nil {
			continue
		}
		for _, code := range stocks {
			klines, err := storage.LoadKLines(code, config.Cfg.KLineDays)
			if err != nil || len(klines) == 0 {
				continue
			}
			if strat.Match(code, klines) {
				last := klines[len(klines)-1]
				storage.SaveResult(code, last.Date, strat.Name())
			}
		}
	}
}
