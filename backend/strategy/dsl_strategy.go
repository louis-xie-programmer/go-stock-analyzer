package strategy

import (
	"fmt"

	"go-stock-analyzer/backend/storage"

	"github.com/Knetic/govaluate"
)

type DSLStrategy struct {
	Expr string
}

func NewDSLStrategy(expr string) *DSLStrategy {
	return &DSLStrategy{Expr: expr}
}

func (s *DSLStrategy) Name() string { return "DSL" }

func (s *DSLStrategy) Match(code string, klines []storage.KLine) bool {
	if len(klines) == 0 {
		return false
	}
	last := klines[len(klines)-1]

	parameters := map[string]interface{}{
		"close":     last.Close,
		"open":      last.Open,
		"high":      last.High,
		"low":       last.Low,
		"volume":    last.Volume,
		"ma5":       last.MA5,
		"ma10":      last.MA10,
		"ma20":      last.MA20,
		"ma30":      last.MA30,
		"macd_dif":  last.DIF,
		"macd_dea":  last.DEA,
		"macd_hist": last.MACD,
	}

	expr, err := govaluate.NewEvaluableExpression(s.Expr)
	if err != nil {
		fmt.Println("dsl parse error:", err)
		return false
	}
	res, err := expr.Evaluate(parameters)
	if err != nil {
		fmt.Println("dsl eval error:", err)
		return false
	}
	pass, ok := res.(bool)
	if !ok {
		return false
	}
	return pass
}
