package storage

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

// 股票基本信息
type StockInfo struct {
	Symbol string  `json:"symbol"`
	Code   string  `json:"code"`
	Name   string  `json:"name"`
	Market string  `json:"market"`
	Board  string  `json:"board"`
	Trade  float64 `json:"trade"`
}
// KLine 日 K 线数据及指标
type KLine struct {
	Code   string  `json:"code"`
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume float64 `json:"volume"`
	MA5    float64 `json:"ma5"`
	MA10   float64 `json:"ma10"`
	MA20   float64 `json:"ma20"`
	MA30   float64 `json:"ma30"`
	DIF    float64 `json:"dif"`
	DEA    float64 `json:"dea"`
	MACD   float64 `json:"macd"`
}

// 初始化数据库连接和表
func InitDB(path string) error {
	var err error
	db, err = sql.Open("sqlite3", path)
	if err != nil {
		return err
	}
	
	_, _ = db.Exec("PRAGMA journal_mode = WAL;")
	_, _ = db.Exec("PRAGMA synchronous = NORMAL;")

	// 股票基本信息表
	stocksSQL := `CREATE TABLE IF NOT EXISTS stocks (
		symbol TEXT PRIMARY KEY,
		code TEXT,
		name TEXT,
		market TEXT,
		board TEXT,
		trade REAL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err = db.Exec(stocksSQL); err != nil {
		return err
	}
	// 自选股表
	watchSQL := `CREATE TABLE IF NOT EXISTS watchlist (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT UNIQUE,
		name TEXT,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`
	if _, err = db.Exec(watchSQL); err != nil {
		return err
	}
	// K线数据表
	klineSQL := `CREATE TABLE IF NOT EXISTS kline (
		code TEXT, date TEXT, open REAL, close REAL, high REAL, low REAL, volume REAL,
		ma5 REAL, ma10 REAL, ma20 REAL, ma30 REAL,
		dif REAL, dea REAL, macd REAL,
		PRIMARY KEY(code,date)
	);`
	if _, err = db.Exec(klineSQL); err != nil {
		return err
	}
  // 策略结果表
	resultsSQL := `CREATE TABLE IF NOT EXISTS results (
		code TEXT, date TEXT, strategy TEXT,
		PRIMARY KEY(code,date,strategy)
	);`
	if _, err = db.Exec(resultsSQL); err != nil {
		return err
	}

	return nil
}

// SaveStocks 批量保存股票基本信息
func SaveStocks(list []StockInfo) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT OR REPLACE INTO stocks(symbol,code,name,market,board,trade,updated_at) VALUES(?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	now := time.Now().Format("2006-01-02 15:04:05")
	for _, s := range list {
		if _, err := stmt.Exec(s.Symbol, s.Code, s.Name, s.Market, s.Board, s.Trade, now); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// CountStocksByBoards 返回指定板块的股票数量
func CountStocksByBoards() (int, error) {
	row := db.QueryRow(`SELECT COUNT(*) FROM stocks WHERE board IN ('上证主板','深证主板','创业板','科创板')`)
	var n int
	if err := row.Scan(&n); err != nil {
		return 0, err
	}
	return n, nil
}

// QueryStocks 按条件分页查询股票列表
func QueryStocks(keyword, board string, offset, limit int) ([]StockInfo, int, error) {
	args := []interface{}{}
	where := " WHERE 1=1 "
	if board != "" {
		where += " AND board = ? "
		args = append(args, board)
	} else {
		where += " AND board IN ('上证主板','深证主板','创业板','科创板') "
	}
	if keyword != "" {
		where += " AND (name LIKE ? OR code LIKE ? OR symbol LIKE ?) "
		kw := "%" + keyword + "%"
		args = append(args, kw, kw, kw)
	}
	countSQL := "SELECT COUNT(*) FROM stocks " + where
	var total int
	if err := db.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	querySQL := "SELECT symbol,code,name,market,board,trade FROM stocks " + where + " ORDER BY code LIMIT ? OFFSET ?"
	args = append(args, limit, offset)
	rows, err := db.Query(querySQL, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []StockInfo{}
	for rows.Next() {
		var s StockInfo
		if err := rows.Scan(&s.Symbol, &s.Code, &s.Name, &s.Market, &s.Board, &s.Trade); err != nil {
			return nil, 0, err
		}
		out = append(out, s)
	}
	return out, total, nil
}

// 自选股票
type WatchStock struct {
	Symbol  string `json:"symbol"`
	Name    string `json:"name"`
	AddedAt string `json:"added_at"`
}
// 添加自选股
func AddToWatchlist(symbol, name string) error {
	_, err := db.Exec("INSERT OR IGNORE INTO watchlist(symbol,name) VALUES(?,?)", symbol, name)
	return err
}
// 移除自选股
func RemoveFromWatchlist(symbol string) error {
	_, err := db.Exec("DELETE FROM watchlist WHERE symbol = ?", symbol)
	return err
}
// 获取自选股列表
func GetWatchlist() ([]WatchStock, error) {
	rows, err := db.Query("SELECT symbol,name,added_at FROM watchlist ORDER BY added_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []WatchStock{}
	for rows.Next() {
		var s WatchStock
		if err := rows.Scan(&s.Symbol, &s.Name, &s.AddedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, nil
}

// 保存K线数据
func SaveKLines(code string, klines []KLine) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT OR REPLACE INTO kline(code,date,open,close,high,low,volume,ma5,ma10,ma20,ma30,dif,dea,macd) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, k := range klines {
		if _, err := stmt.Exec(k.Code, k.Date, k.Open, k.Close, k.High, k.Low, k.Volume, k.MA5, k.MA10, k.MA20, k.MA30, k.DIF, k.DEA, k.MACD); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}
// LoadKLines 加载指定股票的最近 N 天 K 线数据
func LoadKLines(code string, days int) ([]KLine, error) {
	rows, err := db.Query("SELECT date,open,high,low,close,volume,ma5,ma10,ma20,ma30,dif,dea,macd FROM kline WHERE code=? ORDER BY date ASC LIMIT ?", code, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	res := []KLine{}
	for rows.Next() {
		var k KLine
		if err := rows.Scan(&k.Date, &k.Open, &k.High, &k.Low, &k.Close, &k.Volume, &k.MA5, &k.MA10, &k.MA20, &k.MA30, &k.DIF, &k.DEA, &k.MACD); err != nil {
			return nil, err
		}
		k.Code = code
		res = append(res, k)
	}
	return res, nil
}

// SaveResult 保存自动选股结果
func SaveResult(code, date, strategy string) error {
	_, err := db.Exec("INSERT OR REPLACE INTO results(code,date,strategy) VALUES (?,?,?)", code, date, strategy)
	return err
}

func QueryAllBoards() []string {
	return []string{"上证主板", "深证主板", "创业板", "科创板"}
}

// Utility
func ExecSQL(s string) error {
	_, err := db.Exec(s)
	return err
}

// For debug: dump count
func DumpCounts() {
	var n int
	_ = db.QueryRow("SELECT COUNT(*) FROM stocks").Scan(&n)
	log.Println("stocks:", n)
	_ = db.QueryRow("SELECT COUNT(*) FROM watchlist").Scan(&n)
	log.Println("watchlist:", n)
}
