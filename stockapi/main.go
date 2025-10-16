package main

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"github.com/gin-gonic/gin"
)

// GetStocksHandler 获取股票数据
func GetStocksHandler(c *gin.Context) {
	// 获取请求中的股票代码
	symbol := c.Query("symbol")
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}

	// 拼接请求URL
	url := fmt.Sprintf("http://127.0.0.1:18080/api/public/stock_zh_a_hist?symbol=%s", symbol)

	// 发起HTTP请求
	resp, err := http.Get(url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法连接到股票数据源：" + err.Error()})
		return
	}
	defer resp.Body.Close()

	// 如果AKTools返回的状态码不是200，则返回错误
	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": fmt.Sprintf("无法获取数据，状态码：%d", resp.StatusCode)})
		return
	}

	// 读取返回的Body内容
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "无法读取股票数据：" + err.Error()})
		return
	}

	// 将获取到的数据返回
	c.Data(http.StatusOK, "application/json", body)
}

func main() {
	// 创建Gin引擎
	r := gin.Default()

	// 注册路由
	r.GET("/api/stocks", GetStocksHandler)

	// 启动服务器
	r.Run(":8088")
}
