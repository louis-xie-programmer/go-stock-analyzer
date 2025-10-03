package web

import (
	"github.com/gin-gonic/gin"
)

func RunServer() {
	r := gin.Default()
	r.GET("/api/results", GetResults)
	r.POST("/api/dsl/test", TestDSL)
	r.Run(":8080")
}
