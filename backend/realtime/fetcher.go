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
				url := "http://hq.sinajs.cn/list=" + strings.Join(batch, ",")
				req, _ := http.NewRequest("GET", url, nil)
				req.Header.Add("Referer", "http://finance.sina.com.cn/")
				c := &http.Client{}
				resp, err := c.Do(req)
				if err != nil {
					log.Println("realtime poll error:", err)
					continue
				}
				body, _ := io.ReadAll(resp.Body)
				resp.Body.Close()
				parseAndBroadcast(string(body))
			}
			<-ticker.C
		}
	}()
}

func parseAndBroadcast(raw string) {
	lines := strings.Split(raw, ";")
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
			tstr = fields[len(fields)-3] + " " + fields[len(fields)-2]
		}
		q := Quote{Code: code, Name: name, Price: price, PrevClose: prev, Open: open, High: high, Low: low, Volume: vol, Time: tstr}
		Snapshot.mu.Lock()
		Snapshot.m[code] = q
		Snapshot.mu.Unlock()
		b, _ := json.Marshal(q)
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
