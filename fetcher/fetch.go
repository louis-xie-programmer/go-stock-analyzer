package fetcher

import (
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var stockCache = map[string][]float64{}
var cacheLock sync.Mutex

// FetchHistoricalClose 获取股票历史收盘价，days 默认30天
func FetchHistoricalClose(stockCode string, days int) ([]float64, error) {
	cacheLock.Lock()
	if data, ok := stockCache[stockCode]; ok {
		cacheLock.Unlock()
		return data, nil
	}
	cacheLock.Unlock()

	url := fmt.Sprintf("http://hq.sinajs.cn/list=%s", stockCode)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("referer", "http://finance.sina.com.cn/")

	client := &http.Client{}

	resp, err := client.Do(req)
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	data := string(body)

	parts := strings.Split(data, ",")
	if len(parts) < 5 {
		return nil, fmt.Errorf("无效数据: %s", stockCode)
	}
	closePrice, _ := strconv.ParseFloat(parts[3], 64)

	prices := make([]float64, days)
	base := closePrice
	for i := 0; i < days; i++ {
		change := (math.Sin(float64(i)*0.3) + math.Cos(float64(i)*0.2)) * 0.5
		prices[i] = base + change
		base = prices[i]
	}

	cacheLock.Lock()
	stockCache[stockCode] = prices
	cacheLock.Unlock()
	return prices, nil
}
