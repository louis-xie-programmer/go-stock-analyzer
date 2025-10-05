package fetcher

func CalcMA(values []float64, n int) float64 {
	if len(values) < n {
		return 0
	}
	sum := 0.0
	for i := len(values) - n; i < len(values); i++ {
		sum += values[i]
	}

	return sum / float64(n)
}

func CalcEMASequence(values []float64, period int) []float64 {
	n := len(values)
	out := make([]float64, n)
	if n == 0 {
		return out
	}
	mult := 2.0 / float64(period+1)
	out[0] = values[0]
	for i := 1; i < n; i++ {
		out[i] = (values[i]-out[i-1])*mult + out[i-1]
	}
	return out
}

func CalcMACD(values []float64) (dif, dea, macd float64) {
	if len(values) == 0 {
		return 0, 0, 0
	}
	ema12 := CalcEMASequence(values, 12)
	ema26 := CalcEMASequence(values, 26)
	difSeries := make([]float64, len(values))
	for i := range values {
		difSeries[i] = ema12[i] - ema26[i]
	}
	deaSeries := CalcEMASequence(difSeries, 9)
	last := len(values) - 1
	dif = difSeries[last]
	dea = deaSeries[last]
	macd = 2 * (dif - dea)
	return
}
