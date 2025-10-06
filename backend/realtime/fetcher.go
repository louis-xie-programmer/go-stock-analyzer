package realtime

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var IsMarketOpen = func(code string) bool {
	// 简单判断是否在交易时间内（9:30-11:30, 13:00-15:00）
	now := time.Now()
	weekday := now.Weekday()
	if weekday == time.Saturday || weekday == time.Sunday {
		return false
	}
	hour := now.Hour()
	minute := now.Minute()
	if ((hour == 9 && minute < 30) || (hour < 9) || (hour == 11 && minute > 30)) {
		return false
	}
	if ((hour == 12) || (hour > 15)) {
		return false
	}

	quotes, err := FetchSinaQuotes([]string{code}) // 试探性请求，确保行情接口可用
	if err != nil || len(quotes) == 0 {
		return false
	}
	currDate := now.Format("2006-01-02")
	has := strings.HasPrefix(quotes[0].Time, currDate+" ")
	if (has && quotes[0].Price != 0) {
		return true
	}
	
	return false
}

type Quote struct {
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	PrevClose float64 `json:"prev_close"`
	Open      float64 `json:"open"`
	High      float64 `json:"high"`
	Low       float64 `json:"low"`
	Volume    int64   `json:"volume"`
	Time      string  `json:"time"`
}

var Snapshot = struct {
	m  map[string]Quote
	mu sync.RWMutex
}{m: make(map[string]Quote)}

func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	upgrader := websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	client := &Client{conn: conn, send: make(chan []byte, 256), subs: make(map[string]bool)}
	h.register <- client
	go client.writePump()
	go client.readPump()
}

func FetchSinaQuotes(codes []string) ([]Quote, error) {
	url := "http://hq.sinajs.cn/list=" + strings.Join(codes, ",")
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Add("Referer", "http://finance.sina.com.cn/")
	c := &http.Client{}
	resp, err := c.Do(req)
	if err != nil {
		log.Println("realtime poll error:", err)
		return nil, err
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	var quotes []Quote

	lines := strings.Split(string(body), ";")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		eq := strings.Index(line, "=")
		if eq < 0 {
			continue
		}
		left := line[:eq]
		start := strings.LastIndex(left, "hq_str_")
		if start < 0 {
			continue
		}
		code := left[start+7:]
		qstart := strings.Index(line, "\"")
		qend := strings.LastIndex(line, "\"")
		if qstart < 0 || qend <= qstart {
			continue
		}
		body := line[qstart+1 : qend]
		fields := strings.Split(body, ",")
		if len(fields) < 6 {
			continue
		}
		name := fields[0]
		open := parseFloat(fields[1])
		prev := parseFloat(fields[2])
		price := parseFloat(fields[3])
		high := parseFloat(fields[4])
		low := parseFloat(fields[5])
		var vol int64 = 0
		if len(fields) > 8 {
			vol = parseInt64(fields[8])
		}
		tstr := ""
		if len(fields) >= 32 {
			tstr = fields[30] + " " + fields[31]
		}
		q := Quote{Code: code, Name: name, Price: price, PrevClose: prev, Open: open, High: high, Low: low, Volume: vol, Time: tstr}
		quotes = append(quotes, q)
	}
	return quotes, nil
}
// 广播消息给所有客户端
func StartPolling(symbols []string, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			for i := 0; i < len(symbols); i += 60 {
				j := i + 60
				if j > len(symbols) {
					j = len(symbols)
				}
				batch := symbols[i:j]
				quotes, err := FetchSinaQuotes(batch)
				if err != nil {
					log.Println("realtime poll error:", err)
					continue
				}
				parseAndBroadcast(quotes)
			}
			<-ticker.C
		}
	}()
}

// 广播消息给所有客户端
func parseAndBroadcast(quotes []Quote) {
	for _, quote := range quotes {
		code := quote.Code
		Snapshot.mu.Lock()
		Snapshot.m[code] = quote
		Snapshot.mu.Unlock()
		b, _ := json.Marshal(quote)
		Broadcast(b)
	}
}

func parseFloat(s string) float64 {
	v, _ := strconv.ParseFloat(strings.TrimSpace(s), 64)
	return v
}
func parseInt64(s string) int64 {
	v, _ := strconv.ParseInt(strings.TrimSpace(s), 10, 64)
	return v
}
