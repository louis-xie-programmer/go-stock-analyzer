package fetcher

import (
	"encoding/json"
	"fmt"
	"go-stock-analyzer/backend/storage"
	"io"
	"net/http"
	"strings"
	"time"
)

// 新浪财经返回的股票条目
type sinaItem struct {
	Symbol string `json:"symbol"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Trade  string `json:"trade"`
}

// 板块分类
func classifyBoard(symbol, code string) string {
	// symbol like "sh600000" or "sz000001"
	if strings.HasPrefix(symbol, "sh") {
		if strings.HasPrefix(code, "688") {
			return "科创板"
		}
		// treat other sh as 上证主板 if starts with 6
		if strings.HasPrefix(code, "6") {
			return "上证主板"
		}
	}
	if strings.HasPrefix(symbol, "sz") {
		if strings.HasPrefix(code, "300") || strings.HasPrefix(code, "301") {
			return "创业板"
		}
		if strings.HasPrefix(code, "000") {
			return "深证主板"
		}
	}
	return ""
}

// FetchAllStocks 从新浪财经抓取沪深A股列表并分类为四大板块
func FetchAllStocks() ([]storage.StockInfo, error) {
	var all []storage.StockInfo
	page := 1
	pageSize := 200

	client := &http.Client{Timeout: 15 * time.Second}
	for {
		url := fmt.Sprintf("http://vip.stock.finance.sina.com.cn/quotes_service/api/json_v2.php/Market_Center.getHQNodeData?page=%d&num=%d&sort=symbol&asc=1&node=hs_a", page, pageSize)
		resp, err := client.Get(url)
		if err != nil {
			// network error - return so caller can decide
			return nil, err
		}
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		text := strings.TrimSpace(string(body))
		text = strings.ReplaceAll(text, "'", "\"")
		if text == "[]" || text == "" {
			break
		}
		var items []sinaItem
		if err := json.Unmarshal([]byte(text), &items); err != nil {
			// if unmarshal fails, stop fetching
			break
		}
		if len(items) == 0 {
			break
		}
		for _, it := range items {
			board := classifyBoard(it.Symbol, it.Code)
			if board == "" {
				continue
			}
			trade := 0.0
			// parse trade float safely
			fmt.Sscanf(it.Trade, "%f", &trade)
			s := storage.StockInfo{
				Symbol: it.Symbol,
				Code:   it.Code,
				Name:   it.Name,
				Market: strings.ToUpper(it.Symbol[:2]),
				Board:  board,
				Trade:  trade,
			}
			all = append(all, s)
		}
		// next page
		page++
		// sleep a bit to be polite
		time.Sleep(200 * time.Millisecond)
		// safety cap
		if page > 50 {
			break
		}
	}
	return all, nil
}
