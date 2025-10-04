package strategy

import "go-stock-analyzer/backend/storage"

type MAStrategy struct {
	MA       int
	HoldDays int
}

func NewMAStrategy(ma, holdDays int) *MAStrategy {
	return &MAStrategy{MA: ma, HoldDays: holdDays}
}

func (s *MAStrategy) Name() string { return "MA" }

func (s *MAStrategy) Match(code string, klines []storage.KLine) bool {
	if len(klines) < s.HoldDays {
		return false
	}
	for i := len(klines) - s.HoldDays; i < len(klines); i++ {
		k := klines[i]
		var maValue float64
		switch s.MA {
		case 5:
			maValue = k.MA5
		case 10:
			maValue = k.MA10
		case 20:
			maValue = k.MA20
		case 30:
			maValue = k.MA30
		default:
			maValue = k.MA20
		}
		if k.Close < maValue {
			return false
		}
	}
	return true
}
