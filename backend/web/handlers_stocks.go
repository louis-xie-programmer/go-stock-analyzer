package web

import (
	"go-stock-analyzer/backend/fetcher"
	"go-stock-analyzer/backend/realtime"
	"go-stock-analyzer/backend/storage"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// 查询是否开始交易
// GET /api/is_market_open?code=sz000001
func IsMarketOpenHandler(c *gin.Context) {
	code := strings.TrimSpace(c.Query("code"))
	open := realtime.IsMarketOpen(code)
	c.JSON(http.StatusOK, gin.H{"is_open": open})
}

// GET /api/stocks?q=&board=&page=&size=
func GetStocksHandler(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	board := strings.TrimSpace(c.Query("board"))
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "50")
	page, _ := strconv.Atoi(pageStr)
	size, _ := strconv.Atoi(sizeStr)
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 500 {
		size = 50
	}
	offset := (page - 1) * size
	list, total, err := storage.QueryStocks(q, board, offset, size)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"total": total, "list": list})
}

// Watchlist handlers
func GetWatchlistHandler(c *gin.Context) {
	list, err := storage.GetWatchlist()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}

func AddWatchlistHandler(c *gin.Context) {
	var body struct {
		Symbol string `json:"symbol"`
		Name   string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if body.Symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol required"})
		return
	}
	if err := storage.AddToWatchlist(body.Symbol, body.Name); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "ok"})
}

func RemoveWatchlistHandler(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol required"})
		return
	}
	if err := storage.RemoveFromWatchlist(symbol); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"msg": "removed"})
}

// GET /api/kline?symbol=sz000001&datalen=120
func GetKLineHandler(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol required"})
		return
	}
	datalenStr := c.DefaultQuery("datalen", "120")
	datalen, _ := strconv.Atoi(datalenStr)
	if datalen <= 0 {
		datalen = 120
	}
	// If symbol is in watchlist (自选股) -> try to load from DB (these are persisted at startup with 300 days)
	watch, _ := storage.GetWatchlist()
	isWatch := false
	for _, w := range watch {
		if w.Symbol == symbol {
			isWatch = true
			break
		}
	}
	if isWatch {
		// load from DB; datalen may be <= stored days (we store 300 days at startup)
		klines, err := storage.LoadKLines(symbol, datalen)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		// if DB returns fewer than requested (e.g., first time), try fetch and save
		if len(klines) < datalen {
			fetched, err := fetcher.FetchKLine(symbol, datalen)
			if err == nil && len(fetched) > 0 {
				_ = storage.SaveKLines(symbol, fetched)
				// return fetched (most up-to-date)
				c.JSON(http.StatusOK, fetched)
				return
			}
		}
		c.JSON(http.StatusOK, klines)
		return
	}

	// Non-watch symbols: fetch on-the-fly from remote and do NOT persist (only return datalen, default 120)
	klines, err := fetcher.FetchKLine(symbol, datalen)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, klines)
}
