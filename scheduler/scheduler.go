package scheduler

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"time"

	"go-stock-analyzer/analyzer"
	"go-stock-analyzer/fetcher"
)

type Stock struct {
	Code  string
	Close []float64
}

// 并发筛选股票
func FilterStocksConcurrent(stockCodes []string, workerNum int) []Stock {
	selected := []Stock{}
	stockChan := make(chan string, len(stockCodes))
	resultChan := make(chan Stock, len(stockCodes))
	done := make(chan struct{})

	for i := 0; i < workerNum; i++ {
		go func() {
			for code := range stockChan {
				closes, err := fetcher.FetchHistoricalClose(code, 30)
				if err != nil {
					continue
				}
				if analyzer.CheckContinuousAboveMA20(closes) {
					resultChan <- Stock{Code: code, Close: closes}
				}
			}
			done <- struct{}{}
		}()
	}

	// 发送任务
	go func() {
		for _, code := range stockCodes {
			stockChan <- code
		}
		close(stockChan)
	}()

	// 等待所有 worker 完成
	go func() {
		for i := 0; i < workerNum; i++ {
			<-done
		}
		close(resultChan)
	}()

	for s := range resultChan {
		selected = append(selected, s)
	}
	return selected
}

// 保存CSV
func SaveToCSV(stocks []Stock, filename string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Println("创建 CSV 文件失败:", err)
		return
	}
	defer file.Close()
	writer := csv.NewWriter(file)
	defer writer.Flush()
	writer.Write([]string{"股票代码"})
	for _, s := range stocks {
		writer.Write([]string{s.Code})
	}
}

// 每日定时任务
func ScheduleDailyTask(stockFile, resultDir string, hour, minute, workerNum int) {
	for {
		now := time.Now()
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if now.After(nextRun) {
			nextRun = nextRun.Add(24 * time.Hour)
		}
		time.Sleep(nextRun.Sub(now))

		fmt.Println("📊 开始每日股票筛选任务...")

		file, err := os.Open(stockFile)
		if err != nil {
			log.Println("打开股票列表失败:", err)
			continue
		}
		reader := csv.NewReader(file)
		lines, _ := reader.ReadAll()
		file.Close()

		codes := []string{}
		for _, line := range lines {
			if len(line) > 0 {
				codes = append(codes, line[0])
			}
		}

		selected := FilterStocksConcurrent(codes, workerNum)
		filename := fmt.Sprintf("%s/selected_%s.csv", resultDir, time.Now().Format("20060102"))
		SaveToCSV(selected, filename)

		fmt.Printf("✅ 今日筛选完成，共 %d 只股票，结果已保存到 %s\n", len(selected), filename)
	}
}

func ScheduleDailyTaskWithWeb(stockFile, resultDir string, hour, minute, workerNum int, updateWeb func([]Stock)) {
	for {
		now := time.Now()
		nextRun := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())
		if now.After(nextRun) {
			nextRun = nextRun.Add(24 * time.Hour)
		}
		time.Sleep(nextRun.Sub(now))

		file, _ := os.Open(stockFile)
		reader := csv.NewReader(file)
		lines, _ := reader.ReadAll()
		file.Close()

		codes := []string{}
		for _, line := range lines {
			if len(line) > 0 {
				codes = append(codes, line[0])
			}
		}

		selected := FilterStocksConcurrent(codes, workerNum)
		filename := fmt.Sprintf("%s/selected_%s.csv", resultDir, time.Now().Format("20060102"))
		SaveToCSV(selected, filename)

		// 更新Web展示数据
		updateWeb(selected)
	}
}
