package scheduler

import (
	"log"
	"time"

	"go-stock-analyzer/backend/config"
	"go-stock-analyzer/backend/fetcher"
	"go-stock-analyzer/backend/storage"
	"go-stock-analyzer/backend/strategy"
)

// StartDailyTask 启动每日定时任务协程
func StartDailyTask() {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), config.Cfg.UpdateHour, config.Cfg.UpdateMinute, 0, 0, now.Location())
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}
			sleep := time.Until(next)
			log.Printf("Scheduler sleeping until %v\n", next)
			time.Sleep(sleep)
			runAnalysis()
		}
	}()
}
// 执行每日分析任务
func runAnalysis() {
	log.Println("Daily analysis start: reading watchlist")
	watch, err := storage.GetWatchlist()
	if err != nil {
		log.Println("get watchlist error:", err)
		return
	}
	if len(watch) == 0 {
		log.Println("watchlist empty, nothing to analyze")
		return
	}
	
	var symbols []string
	for _, w := range watch {
		symbols = append(symbols, w.Symbol)
	}
	log.Printf("Analyzing %d watched stocks\n", len(symbols))

	for _, sym := range symbols {
		klines, err := fetcher.FetchKLine(sym, config.Cfg.KLineDays)
		if err != nil {
			log.Printf("fetch kline %s error: %v", sym, err)
			continue
		}
		
		if err := storage.SaveKLines(sym, klines); err != nil {
			log.Printf("save kline %s error: %v", sym, err)
		}
	}
	// 运行所有策略
	strategy.RunAll(symbols)
	log.Println("Daily analysis finished")
}
