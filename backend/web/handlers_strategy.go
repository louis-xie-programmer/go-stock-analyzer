package web

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"go-stock-analyzer/backend/storage"
	"go-stock-analyzer/backend/strategyexec"
)

// POST /api/strategy 保存策略
func SaveStrategyHandler(c *gin.Context) {
	var body struct {
		Name string `json:"name"`
		Desc string `json:"description"`
		Code string `json:"code"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	// SaveStrategyDB 需要 storage 层函数访问 db
	id, err := storage.SaveStrategyDB(&storage.Strategy{
		Name: body.Name, Desc: body.Desc, Code: body.Code,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

// POST /api/strategy/run 运行策略
// body: { "id": optional, "code": optional, "target": "watchlist"|"board:上证主板"|"all" }
func RunStrategyHandler(c *gin.Context) {
	var body struct {
		ID     int64  `json:"id"`
		Code   string `json:"code"`
		Target string `json:"target"`
		Days   int    `json:"days"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	code := body.Code
	if body.ID != 0 && code == "" {
		// load from DB
		s, err := storage.GetStrategyDB(body.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		code = s.Code
	}
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no code provided"})
		return
	}
	// determine symbols based on target
	symbols := []string{}
	if body.Target == "" || body.Target == "watchlist" {
		wl, _ := storage.GetWatchlist() // implement GetWatchlistDB that uses storage db
		for _, w := range wl {
			symbols = append(symbols, w.Symbol)
		}
	} else if strings.HasPrefix(body.Target, "board:") {
		board := strings.TrimPrefix(body.Target, "board:")
		list, _, _ := storage.QueryStocks("", board, 0, 10000)
		for _, s := range list {
			symbols = append(symbols, s.Symbol)
		}
	} else if body.Target == "all" {
		list, _, _ := storage.QueryStocks("", "", 0, 1000000)
		for _, s := range list {
			symbols = append(symbols, s.Symbol)
		}
	} else {
		// fallback: treat target as comma-separated symbols
		symbols = strings.Split(body.Target, ",")
	}

	if len(symbols) == 0 {
		c.JSON(http.StatusOK, gin.H{"matches": []string{}})
		return
	}

	// loader function: use storage.LoadKLines (package-level) - be careful with days default
	days := body.Days
	if days <= 0 {
		days = 120
	}
	loader := func(sym string, d int) ([]storage.KLine, error) {
		return storage.LoadKLines(sym, d)
	}

	start := time.Now()
	matches, err := strategyexec.ExecuteStrategy(code, symbols, days, loader, strategyexec.DefaultExecConfig)
	duration := time.Since(start)
	// Save run log optionally (if ID provided)
	if body.ID != 0 {
		_ = storage.SaveStrategyRunLog(body.ID, body.Target, len(matches), duration.Milliseconds(), "")
	}
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"matches": matches, "error": err.Error(), "duration_ms": duration.Milliseconds()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"matches": matches, "duration_ms": duration.Milliseconds()})
}

// GET /api/strategy 查询所有策略
func ListStrategiesHandler(c *gin.Context) {
	list, err := storage.ListStrategiesDB()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"list": list})
}

// PUT /api/strategy/:id 修改策略
func UpdateStrategyHandler(c *gin.Context) {
	var body struct {
		Name   string `json:"name"`
		Desc   string `json:"description"`
		Code   string `json:"code"`
		Author string `json:"author"`
	}
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	s := &storage.Strategy{
		ID:     id,
		Name:   body.Name,
		Desc:   body.Desc,
		Code:   body.Code,
		Author: body.Author,
	}
	if err := storage.UpdateStrategyDB(s); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}

// DELETE /api/strategy/:id 删除策略
func DeleteStrategyHandler(c *gin.Context) {
	idStr := c.Param("id")
	if idStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing id"})
		return
	}
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	if err := storage.DeleteStrategyDB(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"id": id})
}
