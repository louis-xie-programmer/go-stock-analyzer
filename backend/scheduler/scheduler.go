package scheduler

import (
	"log"
	"time"

	"go-stock-analyzer/backend/config"
	"go-stock-analyzer/backend/fetcher"
	"go-stock-analyzer/backend/storage"
	"go-stock-analyzer/backend/strategy"
)

func DailyUpdate(stocks []string) {
	log.Println("DailyUpdate: start fetching and computing...")
	for _, code := range stocks {
		klines, err := fetcher.FetchAndCompute(code, config.Cfg.KLineDays)
		if err != nil {
			log.Printf("fetch error %s: %v\n", code, err)
			continue
		}
		if err := storage.SaveKLines(code, klines); err != nil {
			log.Printf("save klines error %s: %v\n", code, err)
		}
	}
	log.Println("Running strategies...")
	strategy.RunAll(stocks)
	log.Println("DailyUpdate: done")
}

func ScheduleDailyUpdate(stocks []string) {
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
			DailyUpdate(stocks)
		}
	}()
}
