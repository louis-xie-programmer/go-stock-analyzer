package web

import (
	"net/http"

	"go-stock-analyzer/backend/storage"
	"go-stock-analyzer/backend/strategy"

	"github.com/gin-gonic/gin"
)

func GetResults(c *gin.Context) {
	results, err := storage.LoadResults(200)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, results)
}

func TestDSL(c *gin.Context) {
	var body struct {
		Code string `json:"code"`
		Expr string `json:"expr"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	klines, err := storage.LoadKLines(body.Code, 60)
	if err != nil || len(klines) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "no klines"})
		return
	}
	dsl := strategy.NewDSLStrategy(body.Expr)
	ok := dsl.Match(body.Code, klines)
	c.JSON(http.StatusOK, gin.H{"code": body.Code, "expr": body.Expr, "match": ok, "params": klines[len(klines)-1]})
}
