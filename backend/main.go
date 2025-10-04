package main

import (
	"encoding/csv"
	"log"
	"os"
	"time"

	"go-stock-analyzer/backend/config"
	"go-stock-analyzer/backend/realtime"
	"go-stock-analyzer/backend/scheduler"
	"go-stock-analyzer/backend/storage"
	"go-stock-analyzer/backend/web"
)

func loadStockList(path string) ([]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	records, err := r.ReadAll()
	if err != nil {
		return nil, err
	}
	out := []string{}
	for _, rec := range records {
		if len(rec) > 0 {
			out = append(out, rec[0])
		}
	}
	return out, nil
}

func main() {
	config.LoadConfig("backend/config/config.yaml")
	storage.InitDB(config.Cfg.DBPath)

	stocks, err := loadStockList(config.Cfg.StockListPath)
	if err != nil {
		log.Printf("load stock list error: %v\n", err)
		stocks = []string{"sz000001"}
	}

	// 抓取所有股票列表
	scheduler.Start()

	// initial update
	// scheduler.DailyUpdate(stocks)

	// start realtime polling
	realtime.StartPolling(stocks, 2*time.Second)

	// schedule daily updates
	//scheduler.ScheduleDailyUpdate(stocks)

	log.Println("Starting web server on :8080")
	web.RunServer()
}
