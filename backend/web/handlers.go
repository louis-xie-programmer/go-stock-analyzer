package web

import (
	"github.com/gin-gonic/gin"
	"go-stock-analyzer/backend/storage"
	"net/http"
)

func GetResults(c *gin.Context) {
	rows, _ := storage.DB.Query("SELECT code,date,strategy FROM results ORDER BY date DESC LIMIT 100")
	var results []map[string]string
	for rows.Next() {
		var code, date, strategy string
		rows.Scan(&code, &date, &strategy)
		results = append(results, map[string]string{
			"code": code, "date": date, "strategy": strategy})
	}
	c.JSON(http.StatusOK, results)
}

func TestDSL(c *gin.Context) {
	var body struct {
		Expr string `json:"expr"`
	}
	c.BindJSON(&body)
	c.JSON(200, gin.H{"expr": body.Expr, "ok": true})
}
