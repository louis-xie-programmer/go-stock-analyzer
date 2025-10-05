package main

import (
	"log"
	"sync"
	"time"

	"go-stock-analyzer/backend/config"
	"go-stock-analyzer/backend/fetcher"
	"go-stock-analyzer/backend/realtime"
	"go-stock-analyzer/backend/scheduler"
	"go-stock-analyzer/backend/storage"
	"go-stock-analyzer/backend/web"
)

func main() {
	// 加载配置
	config.LoadConfig("backend/config/config.yaml")

	// init db
	if err := storage.InitDB(config.Cfg.DBPath); err != nil {
		log.Fatalf("init db failed: %v", err)
	}
	// 每次启动校验板块股票列表是否有更新（首次启动会初始化）
	log.Println("checking stock list updates from Sina...")
	if list, err := fetcher.FetchAllStocks(); err != nil {
		log.Printf("fetch all stocks failed: %v", err)
	} else {
		// compare counts; if different或更新则 upsert 到 DB
		n, err := storage.CountStocksByBoards()
		if err != nil {
			log.Printf("count stocks error: %v", err)
		}
		log.Printf("existing stocks in DB (boards of interest): %d; fetched: %d", n, len(list))
		// 如果数量不一致，尝试保存（INSERT OR REPLACE 会做 upsert）
		if len(list) == 0 {
			log.Println("fetched stock list empty, skipping save")
		} else if n != len(list) {
			log.Println("stock list changed or first-run — saving fetched list...")
			if err := storage.SaveStocks(list); err != nil {
				log.Printf("save stocks failed: %v", err)
			} else {
				log.Printf("saved %d stocks", len(list))
			}
		} else {
			log.Println("stock list appears unchanged; no save needed")
		}
	}
	// 启动时加载自选股的 K 线数据并保存
	// load watchlist (自选股)
	watch, _ := storage.GetWatchlist()
	symbols := []string{}
	for _, w := range watch {
		symbols = append(symbols, w.Symbol)
	}
	log.Printf("symbols: %d\n", len(symbols))
	// 为自选股抓取 300 天日 K 线并保存（后台异步执行以不阻塞 web 启动）
	if len(symbols) > 0 {
		go func(syms []string) {
			conc := config.Cfg.WorkerConcurrency
			if conc <= 0 {
				conc = 5
			}
			retries := config.Cfg.WorkerRetries
			if retries <= 0 {
				retries = 3
			}
			delayMs := config.Cfg.WorkerDelayMs
			if delayMs <= 0 {
				delayMs = 200
			}
			backoffMs := config.Cfg.WorkerBackoffMs
			if backoffMs <= 0 {
				backoffMs = 500
			}
			days := config.Cfg.WatchlistKlineDays
			if days <= 0 {
				days = 300
			}

			jobs := make(chan string, len(syms))
			var wg sync.WaitGroup

			for i := 0; i < conc; i++ {
				wg.Add(1)
				go func(id int) {
					defer wg.Done()
					clientDelay := time.Duration(delayMs) * time.Millisecond
					for sym := range jobs {
						var lastErr error
						for attempt := 1; attempt <= retries; attempt++ {
							klines, err := fetcher.FetchKLine(sym, days)
							log.Printf("worker %d: fetched kline for %s (attempt %d/%d): %d days, err=%v", id, sym, attempt, retries, len(klines), err)
							if err == nil {
								if err := storage.SaveKLines(sym, klines); err != nil {
									lastErr = err
								} else {
									// saved successfully
									lastErr = nil
									break
								}
							} else {
								lastErr = err
								// backoff
								time.Sleep(time.Duration(attempt*backoffMs) * time.Millisecond)
							}
						}
						if lastErr != nil {
							log.Printf("startup: failed to fetch/save kline for %s: %v", sym, lastErr)
						}
						time.Sleep(clientDelay)
					}
				}(i)
			}

			for _, s := range syms {
				jobs <- s
			}
			close(jobs)
			wg.Wait()
		}(symbols)
	}
	// if no watchlist, poll a few top symbols from DB
	if len(symbols) == 0 {
		// get some example symbols
		stocks, _, _ := storage.QueryStocks("", "", 0, 50)
		for _, s := range stocks {
			symbols = append(symbols, s.Symbol)
		}
	}

	realtime.StartPolling(symbols, 2*time.Second)

	// start daily scheduler (15:00)
	scheduler.StartDailyTask()

	// run web server (blocks)
	log.Println("starting web server :8080")
	web.RunServer()
}
