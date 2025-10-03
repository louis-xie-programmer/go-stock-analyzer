package scheduler

import (
	"go-stock-analyzer/backend/config"
	"go-stock-analyzer/backend/fetcher"
	"go-stock-analyzer/backend/storage"
	"go-stock-analyzer/backend/strategy"
	"time"
)

func DailyUpdate(stocks []string) {
	for _, code := range stocks {
		klines, err := fetcher.FetchAndCompute(code, config.Cfg.KLineDays)
		if err == nil {
			storage.SaveKLines(code, klines)
		}
	}
	strategy.RunAll(stocks)
}

func ScheduleDailyUpdate(stocks []string) {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), config.Cfg.UpdateHour, config.Cfg.UpdateMinute, 0, 0, now.Location())
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}
			time.Sleep(time.Until(next))
			DailyUpdate(stocks)
		}
	}()
}
