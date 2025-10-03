package strategy

import (
	"go-stock-analyzer/backend/storage"
)

type CompositeStrategy struct {
	HoldDays int
}

func NewCompositeStrategy(holdDays int) *CompositeStrategy {
	return &CompositeStrategy{HoldDays: holdDays}
}

func (s *CompositeStrategy) Name() string { return "Composite" }

func (s *CompositeStrategy) Match(code string, klines []storage.KLine) bool {
	if len(klines) < s.HoldDays+1 {
		return false
	}

	maStrat := NewMAStrategy(20, s.HoldDays)
	macdStrat := NewMACDStrategy()

	return maStrat.Match(code, klines) && macdStrat.Match(code, klines)
}
