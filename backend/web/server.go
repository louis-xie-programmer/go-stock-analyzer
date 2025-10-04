package web

import (
	"go-stock-analyzer/backend/realtime"
	"net/http"

	"github.com/gin-gonic/gin"
)

func RunServer() {
	r := gin.Default()
	r.GET("/api/results", GetResults)
	r.POST("/api/dsl/test", TestDSL)

	r.GET("/ws/realtime", func(c *gin.Context) {
		realtime.HandleWebSocket(c.Writer, c.Request)
	})

	// start hub
	go realtime.RunHub()

	// serve static frontend if built
	r.Static("/static", "./frontend/dist")
	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "go-stock-analyzer backend")
	})

	r.Run(":8080")
}
