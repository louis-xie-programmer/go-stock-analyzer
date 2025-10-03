package storage

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
)

type KLine struct {
	Code                  string
	Date                  string
	Open                  float64
	High                  float64
	Low                   float64
	Close                 float64
	Volume                float64
	MA5, MA10, MA20, MA30 float64
	DIF, DEA, MACD        float64
}

var DB *sql.DB

func InitDB(path string) {
	var err error
	DB, err = sql.Open("sqlite3", path)
	if err != nil {
		log.Fatal(err)
	}

	DB.Exec(`CREATE TABLE IF NOT EXISTS kline (
        code TEXT, date TEXT, open REAL, close REAL, high REAL, low REAL, volume REAL,
        ma5 REAL, ma10 REAL, ma20 REAL, ma30 REAL,
        dif REAL, dea REAL, macd REAL,
        PRIMARY KEY (code,date)
    )`)

	DB.Exec(`CREATE TABLE IF NOT EXISTS results (
        code TEXT, date TEXT, strategy TEXT,
        PRIMARY KEY(code,date,strategy)
    )`)
}

func SaveKLines(code string, klines []KLine) {
	for _, k := range klines {
		stmt, _ := DB.Prepare(`INSERT OR REPLACE INTO kline
            (code,date,open,close,high,low,volume,ma5,ma10,ma20,ma30,dif,dea,macd)
            VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?)`)
		stmt.Exec(k.Code, k.Date, k.Open, k.Close, k.High, k.Low, k.Volume, k.MA5, k.MA10, k.MA20, k.MA30, k.DIF, k.DEA, k.MACD)
	}
}

func SaveResult(code, date, strategy string) {
	stmt, _ := DB.Prepare("INSERT OR REPLACE INTO results(code,date,strategy) VALUES (?,?,?)")
	stmt.Exec(code, date, strategy)
}

func LoadKLines(code string, days int) ([]KLine, error) {
	rows, _ := DB.Query("SELECT date,open,high,low,close,volume,ma5,ma10,ma20,ma30,dif,dea,macd FROM kline WHERE code=? ORDER BY date ASC LIMIT ?", code, days)
	defer rows.Close()
	klines := []KLine{}
	for rows.Next() {
		var k KLine
		rows.Scan(&k.Date, &k.Open, &k.High, &k.Low, &k.Close, &k.Volume, &k.MA5, &k.MA10, &k.MA20, &k.MA30, &k.DIF, &k.DEA, &k.MACD)
		k.Code = code
		klines = append(klines, k)
	}
	return klines, nil
}
