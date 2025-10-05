package fetcher

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"go-stock-analyzer/backend/storage"
)

// 获取指定股票的日 K 线数据并计算常用指标（MA, MACD）
func FetchKLine(symbol string, days int) ([]storage.KLine, error) {
	url := fmt.Sprintf("http://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData?symbol=%s&scale=240&ma=no&datalen=%d", symbol, days)
	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()
	text := strings.TrimSpace(string(body))
	if text == "" || text == "null" {
		return nil, fmt.Errorf("empty kline for %s", symbol)
	}
	var raw []struct {
		Day    string `json:"day"`
		Open   string `json:"open"`
		High   string `json:"high"`
		Low    string `json:"low"`
		Close  string `json:"close"`
		Volume string `json:"volume"`
	}
	if err := json.Unmarshal([]byte(text), &raw); err != nil {
		return nil, err
	}
	klines := make([]storage.KLine, 0, len(raw))
	closes := make([]float64, 0, len(raw))
	for _, r := range raw {
		open, _ := strconv.ParseFloat(r.Open, 64)
		high, _ := strconv.ParseFloat(r.High, 64)
		low, _ := strconv.ParseFloat(r.Low, 64)
		closep, _ := strconv.ParseFloat(r.Close, 64)
		vol, _ := strconv.ParseFloat(r.Volume, 64)
		k := storage.KLine{
			Code:   symbol,
			Date:   r.Day,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closep,
			Volume: vol,
		}
		klines = append(klines, k)
		closes = append(closes, closep)
	}
	// 计算指标
	for i := range klines {
		sub := closes[:i+1]
		klines[i].MA5 = CalcMA(sub, 5)
		klines[i].MA10 = CalcMA(sub, 10)
		klines[i].MA20 = CalcMA(sub, 20)
		klines[i].MA30 = CalcMA(sub, 30)
		dif, dea, macd := CalcMACD(sub)
		klines[i].DIF = dif
		klines[i].DEA = dea
		klines[i].MACD = macd
	}
	return klines, nil
}
