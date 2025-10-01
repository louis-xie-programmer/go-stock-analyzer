package main

import (
	"go-stock-analyzer/scheduler"
	"go-stock-analyzer/web"
	"os"
)

func main() {
	stockFile := "stock_list.csv"
	resultDir := "results"
	hour, minute := 15, 0
	workerNum := 20

	if _, err := os.Stat(resultDir); os.IsNotExist(err) {
		os.Mkdir(resultDir, 0755)
	}

	// 启动Web服务器
	go web.StartWebServer(":8080")

	// 每日任务
	scheduler.ScheduleDailyTaskWithWeb(stockFile, resultDir, hour, minute, workerNum, web.SetStockData)
}
