package strategy

import (
	"go-stock-analyzer/backend/storage"
	"strconv"
	"strings"
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
	k := klines[len(klines)-1]
	expr := strings.ToLower(s.Expr)

	vars := map[string]float64{
		"close": k.Close,
		"ma5":   k.MA5, "ma10": k.MA10, "ma20": k.MA20, "ma30": k.MA30,
		"macd_dif": k.DIF, "macd_dea": k.DEA, "macd": k.MACD,
	}

	// 简单处理：仅支持 "var > var" "var < var" 组合 AND
	conds := strings.Split(expr, "and")
	for _, cond := range conds {
		cond = strings.TrimSpace(cond)
		var op string
		if strings.Contains(cond, ">=") {
			op = ">="
		}
		if strings.Contains(cond, "<=") {
			op = "<="
		}
		if strings.Contains(cond, ">") {
			op = ">"
		}
		if strings.Contains(cond, "<") {
			op = "<"
		}
		if op == "" {
			continue
		}
		parts := strings.Split(cond, op)
		left := strings.TrimSpace(parts[0])
		right := strings.TrimSpace(parts[1])
		lv := vars[left]
		rv, ok := vars[right]
		if !ok {
			rv, _ = strconv.ParseFloat(right, 64)
		}
		okCond := false
		switch op {
		case ">":
			okCond = lv > rv
		case "<":
			okCond = lv < rv
		case ">=":
			okCond = lv >= rv
		case "<=":
			okCond = lv <= rv
		}
		if !okCond {
			return false
		}
	}
	return true
}
