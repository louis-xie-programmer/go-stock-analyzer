package scheduler

import (
	"log"
	"time"

	"go-stock-analyzer/backend/config"
	"go-stock-analyzer/backend/fetcher"
	"go-stock-analyzer/backend/storage"
	"go-stock-analyzer/backend/strategy"
)

func Start() {
	// 每天 09:00 更新股票池
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 9, 0, 0, 0, now.Location())
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}
			time.Sleep(next.Sub(now))

			log.Println("📈 开始同步股票池数据...")
			stocks, err := fetcher.FetchAllStocks()
			if err != nil {
				log.Printf("同步股票池失败: %v", err)
				continue
			}
			err = storage.SaveStocks(stocks)
			if err != nil {
				log.Printf("保存股票池失败: %v", err)
				continue
			}
			log.Printf("✅ 同步完成，共 %d 只股票", len(stocks))
		}
	}()
}

// 每天15点执行分析任务
func StartDailyTask() {
	go func() {
		for {
			now := time.Now()
			next := time.Date(now.Year(), now.Month(), now.Day(), 15, 0, 0, 0, now.Location())
			if now.After(next) {
				next = next.Add(24 * time.Hour)
			}
			d := next.Sub(now)
			log.Println("下次任务执行时间:", next)
			time.Sleep(d)

			runAnalysis()
		}
	}()
}

func runAnalysis() {
	log.Println("开始执行每日分析任务...")

	stocks, err := storage.GetWatchlist()
	if err != nil {
		log.Println("读取自选股失败:", err)
		return
	}

	var symbols []string

	for _, s := range stocks {
		klines, err := fetcher.FetchAndCompute(s.Symbol, config.Cfg.KLineDays)
		if err != nil {
			log.Printf("fetch error %s: %v\n", s.Symbol, err)
			continue
		}
		if err := storage.SaveKLines(s.Symbol, klines); err != nil {
			log.Printf("save klines error %s: %v\n", s.Symbol, err)
		}
		symbols = append(symbols, s.Symbol)
	}
	log.Println("Running strategies...")
	strategy.RunAll(symbols)

	log.Println("分析任务完成。")
}

//func DailyUpdate(stocks []string) {
//	log.Println("DailyUpdate: start fetching and computing...")
//	for _, code := range stocks {
//		klines, err := fetcher.FetchAndCompute(code, config.Cfg.KLineDays)
//		if err != nil {
//			log.Printf("fetch error %s: %v\n", code, err)
//			continue
//		}
//		if err := storage.SaveKLines(code, klines); err != nil {
//			log.Printf("save klines error %s: %v\n", code, err)
//		}
//	}
//	log.Println("Running strategies...")
//	strategy.RunAll(stocks)
//	log.Println("DailyUpdate: done")
//}

//func ScheduleDailyUpdate(stocks []string) {
//	go func() {
//		for {
//			now := time.Now()
//			next := time.Date(now.Year(), now.Month(), now.Day(), config.Cfg.UpdateHour, config.Cfg.UpdateMinute, 0, 0, now.Location())
//			if now.After(next) {
//				next = next.Add(24 * time.Hour)
//			}
//			sleep := time.Until(next)
//			log.Printf("Scheduler sleeping until %v\n", next)
//			time.Sleep(sleep)
//			DailyUpdate(stocks)
//		}
//	}()
//}
