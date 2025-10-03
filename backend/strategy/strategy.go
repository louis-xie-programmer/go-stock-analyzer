package strategy

import (
	"go-stock-analyzer/backend/config"
	"go-stock-analyzer/backend/storage"
)

// Strategy 策略接口
type Strategy interface {
	Name() string
	Match(code string, klines []storage.KLine) bool
}

func RunAll(stocks []string) {
	for _, s := range config.Cfg.Strategies {
		if !s.Enabled {
			continue
		}
		var strat Strategy
		switch s.Name {
		case "MA":
			strat = NewMAStrategy(int(s.Params["ma"].(int)), int(s.Params["hold_days"].(int)))
		case "MACD":
			strat = NewMACDStrategy()
		case "Composite":
			strat = NewCompositeStrategy(int(s.Params["hold_days"].(int)))
		case "DSL":
			strat = NewDSLStrategy(s.Params["expr"].(string))
		}
		if strat == nil {
			continue
		}

		for _, code := range stocks {
			klines, _ := storage.LoadKLines(code, config.Cfg.KLineDays)
			if strat.Match(code, klines) {
				storage.SaveResult(code, klines[len(klines)-1].Date, strat.Name())
			}
		}
	}
}
