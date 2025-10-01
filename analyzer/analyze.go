package analyzer

// 计算移动平均线
func CalculateMA(prices []float64, period int) float64 {
	if len(prices) < period {
		return 0
	}
	sum := 0.0
	for i := len(prices) - period; i < len(prices); i++ {
		sum += prices[i]
	}
	return sum / float64(period)
}

// 连续3天站上20日均线
func CheckContinuousAboveMA20(closes []float64) bool {
	if len(closes) < 22 {
		return false
	}
	for i := len(closes) - 3; i < len(closes); i++ {
		ma20 := CalculateMA(closes[:i], 20)
		if closes[i] <= ma20 {
			return false
		}
	}
	return true
}
