package storage

import (
	"database/sql"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

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
