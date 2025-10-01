package web

import (
	"encoding/json"
	"go-stock-analyzer/scheduler"
	"net/http"
	"sync"
)

var stockData []scheduler.Stock
var dataLock sync.Mutex

func SetStockData(stocks []scheduler.Stock) {
	dataLock.Lock()
	defer dataLock.Unlock()
	stockData = stocks
}

func StartWebServer(addr string) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/templates/index.html")
	})

	http.HandleFunc("/api/stocks", func(w http.ResponseWriter, r *http.Request) {
		dataLock.Lock()
		defer dataLock.Unlock()
		json.NewEncoder(w).Encode(stockData)
	})

	http.ListenAndServe(addr, nil)
}
