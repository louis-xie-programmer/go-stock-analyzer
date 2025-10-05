package web

import (
	"go-stock-analyzer/backend/realtime"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RunServer() {
	r := gin.Default()

	// API endpoints
	r.GET("/api/stocks", GetStocksHandler)
	r.GET("/api/watchlist", GetWatchlistHandler)
	r.POST("/api/watchlist/add", AddWatchlistHandler)
	r.DELETE("/api/watchlist/remove", RemoveWatchlistHandler)
	r.GET("/api/kline", GetKLineHandler)

	// websocket endpoint for realtime
	r.GET("/ws/realtime", func(c *gin.Context) {
		realtime.HandleWebSocket(c.Writer, c.Request)
	})

	// static frontend (optional)
	r.Static("/static", "../frontend/dist")
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Go-stock-analyzer backend")
	})

	go realtime.RunHub()
	r.Run(":8080")
}
