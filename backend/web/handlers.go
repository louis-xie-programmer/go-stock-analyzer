package web

import (
	"net/http"
	"strconv"
	"strings"

	"go-stock-analyzer/backend/storage"
	"go-stock-analyzer/backend/strategy"

	"github.com/gin-gonic/gin"
)

// StockResponse 用于前端展示
type StockResponse struct {
	Symbol string  `json:"symbol"`
	Code   string  `json:"code"`
	Name   string  `json:"name"`
	Trade  float64 `json:"trade"`
}

// GetStocksHandler 基于 Gin 的股票列表接口
func GetStocksHandler(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "20")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}
	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 {
		size = 20
	}
	offset := (page - 1) * size

	stocks, total, err := storage.QueryStocks(q, offset, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total": total,
		"list":  stocks,
	})
}

// 获取自选股列表
func GetWatchlistHandler(c *gin.Context) {
	list, err := storage.GetWatchlist()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

// 添加股票
func AddWatchlistHandler(c *gin.Context) {
	var req struct {
		Symbol string `json:"symbol"`
		Name   string `json:"name"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "参数错误"})
		return
	}

	if err := storage.AddToWatchlist(req.Symbol, req.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "已加入自选"})
}

// 删除股票
func RemoveWatchlistHandler(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少symbol"})
		return
	}
	if err := storage.RemoveFromWatchlist(symbol); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "已移除"})
}

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
