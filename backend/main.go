package main

import (
	"encoding/csv"
	"go-stock-analyzer/backend/config"
	"go-stock-analyzer/backend/scheduler"
	"go-stock-analyzer/backend/storage"
	"go-stock-analyzer/backend/web"
	"log"
	"os"
)

func main() {
	config.LoadConfig("backend/config/config.yaml")
	storage.InitDB(config.Cfg.DBPath)

	file, _ := os.Open(config.Cfg.StockListPath)
	defer file.Close()
	reader := csv.NewReader(file)
	records, _ := reader.ReadAll()
	stocks := []string{}
	for _, r := range records {
		stocks = append(stocks, r[0])
	}

	scheduler.DailyUpdate(stocks)
	scheduler.ScheduleDailyUpdate(stocks)

	log.Println("系统启动完成，Web端口:8080")
	web.RunServer()
}
