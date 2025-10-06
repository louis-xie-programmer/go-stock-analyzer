package web

import (
	"encoding/json"
	"fmt"
	"go-stock-analyzer/backend/realtime"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

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
	r.GET("/api/timeline", GetTimelineHandler)

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

// 新浪分时接口返回结构
// [{"day":"2024-06-14","time":"09:30","price":"10.23","volume":"1234"}, ...]
type SinaTimeline struct {
	Time   string `json:"time"`
	Price  string `json:"price"`
	Volume string `json:"volume"`
}
type Timeline struct {
	Time   string  `json:"time"`
	Price  float64 `json:"price"`
	Volume float64 `json:"volume"`
}

func GetTimelineHandler(c *gin.Context) {
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol required"})
		return
	}
	url := fmt.Sprintf("http://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData?symbol=%s&scale=5&datalen=48", symbol)
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	str := string(body)
	str = strings.ReplaceAll(str, "'", "\"")
	if !strings.HasPrefix(str, "[") {
		c.JSON(http.StatusOK, []Timeline{})
		return
	}
	// 新浪5分钟K线结构
	type SinaKLine struct {
		Day    string `json:"day"`
		Open   string `json:"open"`
		High   string `json:"high"`
		Low    string `json:"low"`
		Close  string `json:"close"`
		Volume string `json:"volume"`
	}
	var klines []SinaKLine
	if err := json.Unmarshal([]byte(str), &klines); err != nil {
		c.JSON(http.StatusOK, []Timeline{})
		return
	}
	out := make([]Timeline, 0, len(klines))
	for _, k := range klines {
		price, _ := strconv.ParseFloat(k.Close, 64)
		volume, _ := strconv.ParseFloat(k.Volume, 64)
		// 只取时间部分
		t := k.Day
		if len(t) >= 8 {
			// 兼容 "2025-09-30 09:35:00" 或 "09:35:00"
			parts := strings.Split(t, " ")
			if len(parts) == 2 {
				t = parts[1]
			}
		}
		out = append(out, Timeline{Time: t, Price: price, Volume: volume})
	}
	c.JSON(http.StatusOK, out)
}
