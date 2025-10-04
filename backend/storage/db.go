package storage

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// StockInfo 股票基础信息
type StockInfo struct {
	Symbol string `json:"symbol"`
	Code   string `json:"code"`
	Name   string `json:"name"`
	Trade  string `json:"trade"`
}

func QueryStocks(keyword string, offset, limit int) ([]StockInfo, int, error) {
	query := "SELECT symbol, code, name, trade FROM stocks"
	args := []interface{}{}
	countSQL := "SELECT COUNT(*) FROM stocks"

	if keyword != "" {
		query += " WHERE name LIKE ? OR code LIKE ?"
		countSQL += " WHERE name LIKE ? OR code LIKE ?"
		kw := "%" + keyword + "%"
		args = append(args, kw, kw)
	}

	query += " ORDER BY code LIMIT ? OFFSET ?"
	args = append(args, limit, offset)

	rows, err := DB.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []StockInfo
	for rows.Next() {
		var s StockInfo
		err := rows.Scan(&s.Symbol, &s.Code, &s.Name, &s.Trade)
		if err != nil {
			return nil, 0, err
		}
		list = append(list, s)
	}

	var total int
	err = DB.QueryRow(countSQL, args[:len(args)-2]...).Scan(&total)
	if err != nil {
		return list, 0, fmt.Errorf("统计总数失败: %v", err)
	}
	return list, total, nil
}

func SaveStocks(stocks []StockInfo) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT OR REPLACE INTO stocks(symbol, code, name, trade) VALUES (?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, s := range stocks {
		trade, _ := strconv.ParseFloat(s.Trade, 64)
		_, err = stmt.Exec(s.Symbol, s.Code, s.Name, trade)
		if err != nil {
			log.Printf("保存股票失败 %s: %v", s.Code, err)
			continue
		}
	}
	return tx.Commit()
}

type KLine struct {
	Code   string
	Date   string
	Open   float64
	High   float64
	Low    float64
	Close  float64
	Volume float64
	MA5    float64
	MA10   float64
	MA20   float64
	MA30   float64
	DIF    float64
	DEA    float64
	MACD   float64
}

var DB *sql.DB

func InitDB(path string) {
	var err error
	DB, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}

	sqlStmt := `
CREATE TABLE IF NOT EXISTS stocks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	symbol TEXT UNIQUE,
	code TEXT,
	name TEXT,
	trade REAL
);`
	_, err = DB.Exec(sqlStmt)
	if err != nil {
		log.Fatalf("创建 stocks 表失败: %v", err)
	}

	_, err = DB.Exec(`
	CREATE TABLE IF NOT EXISTS watchlist (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		symbol TEXT UNIQUE,
		name TEXT,
		added_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`)
	if err != nil {
		log.Fatalf("自选股 表格失败 %v", err)
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS kline (
        code TEXT, date TEXT, open REAL, close REAL, high REAL, low REAL, volume REAL,
        ma5 REAL, ma10 REAL, ma20 REAL, ma30 REAL,
        dif REAL, dea REAL, macd REAL,
        PRIMARY KEY (code,date)
    )`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = DB.Exec(`CREATE TABLE IF NOT EXISTS results (
        code TEXT, date TEXT, strategy TEXT,
        PRIMARY KEY(code,date,strategy)
    )`)
	if err != nil {
		log.Fatal(err)
	}
}

type WatchStock struct {
	Symbol  string    `json:"symbol"`
	Name    string    `json:"name"`
	AddedAt time.Time `json:"added_at"`
}

// 添加股票到自选池
func AddToWatchlist(symbol, name string) error {
	_, err := DB.Exec(`INSERT OR IGNORE INTO watchlist(symbol, name) VALUES(?, ?)`, symbol, name)
	return err
}

// 从自选池删除
func RemoveFromWatchlist(symbol string) error {
	_, err := DB.Exec(`DELETE FROM watchlist WHERE symbol = ?`, symbol)
	return err
}

// 查询所有自选股
func GetWatchlist() ([]WatchStock, error) {
	rows, err := DB.Query(`SELECT symbol, name, added_at FROM watchlist ORDER BY added_at DESC`)
	if err != nil {
		if err == sql.ErrNoRows {
			return []WatchStock{}, nil
		}
		return nil, err
	}
	defer rows.Close()

	var list []WatchStock
	for rows.Next() {
		var s WatchStock
		if err := rows.Scan(&s.Symbol, &s.Name, &s.AddedAt); err != nil {
			return nil, err
		}
		list = append(list, s)
	}
	return list, nil
}

func SaveKLines(code string, klines []KLine) error {
	tx, err := DB.Begin()
	if err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT OR REPLACE INTO kline
    (code,date,open,close,high,low,volume,ma5,ma10,ma20,ma30,dif,dea,macd)
    VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, k := range klines {
		if _, err := stmt.Exec(k.Code, k.Date, k.Open, k.Close, k.High, k.Low, k.Volume,
			k.MA5, k.MA10, k.MA20, k.MA30, k.DIF, k.DEA, k.MACD); err != nil {
			tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func SaveResult(code, date, strategy string) error {
	_, err := DB.Exec("INSERT OR REPLACE INTO results(code,date,strategy) VALUES (?,?,?)", code, date, strategy)
	return err
}

func LoadKLines(code string, days int) ([]KLine, error) {
	rows, err := DB.Query("SELECT date,open,high,low,close,volume,ma5,ma10,ma20,ma30,dif,dea,macd FROM kline WHERE code=? ORDER BY date ASC LIMIT ?", code, days)
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

func LoadResults(limit int) ([]map[string]string, error) {
	rows, err := DB.Query("SELECT code,date,strategy FROM results ORDER BY date DESC LIMIT ?", limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []map[string]string{}
	for rows.Next() {
		var code, date, strategy string
		if err := rows.Scan(&code, &date, &strategy); err != nil {
			return nil, err
		}
		out = append(out, map[string]string{"code": code, "date": date, "strategy": strategy})
	}
	return out, nil
}
