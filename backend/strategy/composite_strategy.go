package strategy

import "go-stock-analyzer/backend/storage"

type CompositeStrategy struct {
	HoldDays int
}

func NewCompositeStrategy(holdDays int) *CompositeStrategy {
	return &CompositeStrategy{HoldDays: holdDays}
}

func (s *CompositeStrategy) Name() string { return "Composite" }

func (s *CompositeStrategy) Match(code string, klines []storage.KLine) bool {
	if len(klines) < s.HoldDays+2 {
		return false
	}
	ma := NewMAStrategy(20, s.HoldDays)
	macd := NewMACDStrategy()
	return ma.Match(code, klines) && macd.Match(code, klines)
}
