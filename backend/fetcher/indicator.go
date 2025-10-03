package fetcher

func CalcMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}
	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}
	return sum / float64(period)
}

func CalcEMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}
	multiplier := 2.0 / float64(period+1)
	ema := prices[0]
	for i := 1; i < len(prices); i++ {
		ema = (prices[i]-ema)*multiplier + ema
	}
	return ema
}

func CalcMACD(prices []float64) (dif, dea, macd float64) {
	ema12 := CalcEMA(prices, 12)
	ema26 := CalcEMA(prices, 26)
	dif = ema12 - ema26
	dea = CalcEMA([]float64{dif}, 9)
	macd = 2 * (dif - dea)
	return
}
