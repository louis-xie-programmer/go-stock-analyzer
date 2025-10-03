package fetcher

import (
	"encoding/json"
	"fmt"
	"go-stock-analyzer/backend/storage"
	"io/ioutil"
	"net/http"
	"strconv"
)

// FetchAndCompute 获取市场数据（新浪）
func FetchAndCompute(stockCode string, days int) ([]storage.KLine, error) {
	url := fmt.Sprintf(
		"http://money.finance.sina.com.cn/quotes_service/api/json_v2.php/CN_MarketData.getKLineData?symbol=%s&scale=240&ma=no&datalen=%d",
		stockCode, days)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)

	var data []struct {
		Day    string `json:"day"`
		Open   string `json:"open"`
		High   string `json:"high"`
		Low    string `json:"low"`
		Close  string `json:"close"`
		Volume string `json:"volume"`
	}
	json.Unmarshal(body, &data)

	klines := []storage.KLine{}
	closes := []float64{}
	for _, d := range data {
		open, _ := strconv.ParseFloat(d.Open, 64)
		high, _ := strconv.ParseFloat(d.High, 64)
		low, _ := strconv.ParseFloat(d.Low, 64)
		closePrice, _ := strconv.ParseFloat(d.Close, 64)
		volume, _ := strconv.ParseFloat(d.Volume, 64)
		closes = append(closes, closePrice)

		klines = append(klines, storage.KLine{
			Code:   stockCode,
			Date:   d.Day,
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closePrice,
			Volume: volume,
		})
	}

	for i := range klines {
		subCloses := closes[:i+1]
		klines[i].MA5 = CalcMA(subCloses, 5)
		klines[i].MA10 = CalcMA(subCloses, 10)
		klines[i].MA20 = CalcMA(subCloses, 20)
		klines[i].MA30 = CalcMA(subCloses, 30)
		dif, dea, macd := CalcMACD(subCloses)
		klines[i].DIF, klines[i].DEA, klines[i].MACD = dif, dea, macd
	}
	return klines, nil
}
