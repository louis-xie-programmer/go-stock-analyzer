package strategy

import "go-stock-analyzer/backend/storage"

type MACDStrategy struct{}

func NewMACDStrategy() *MACDStrategy { return &MACDStrategy{} }

func (s *MACDStrategy) Name() string { return "MACD" }

func (s *MACDStrategy) Match(code string, klines []storage.KLine) bool {
	if len(klines) < 2 {
		return false
	}
	prev := klines[len(klines)-2]
	last := klines[len(klines)-1]
	return prev.DIF < prev.DEA && last.DIF > last.DEA
}
