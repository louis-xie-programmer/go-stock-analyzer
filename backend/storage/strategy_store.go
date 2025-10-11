package storage

import (
	"time"
)

// Strategy 持久化策略结构
type Strategy struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Desc      string    `json:"description"`
	Code      string    `json:"code"`
	Author    string    `json:"author"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// InitStrategyTable 在 InitDB 后可调用（或合并到 InitDB）
func InitStrategyTable() error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS strategies (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		name TEXT,
		description TEXT,
		code TEXT,
		author TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		return err
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS strategy_runs (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		strategy_id INTEGER,
		target TEXT,
		matches_count INTEGER,
		duration_ms INTEGER,
		err TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	return err
}

// SaveStrategy 保存策略并返回 id
func SaveStrategyDB(s *Strategy) (int64, error) {
	res, err := db.Exec(`INSERT INTO strategies(name, description, code, author, created_at, updated_at) VALUES(?,?,?,?,?,?)`,
		s.Name, s.Desc, s.Code, s.Author, time.Now(), time.Now())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

// UpdateStrategy 更新策略
func UpdateStrategyDB(s *Strategy) error {
	_, err := db.Exec(`UPDATE strategies SET name=?, description=?, code=?, author=?, updated_at=? WHERE id=?`,
		s.Name, s.Desc, s.Code, s.Author, time.Now(), s.ID)
	return err
}

// GetStrategyDB
func GetStrategyDB(id int64) (*Strategy, error) {
	row := db.QueryRow(`SELECT id,name,description,code,author,created_at,updated_at FROM strategies WHERE id=?`, id)
	var s Strategy
	var created, updated string
	if err := row.Scan(&s.ID, &s.Name, &s.Desc, &s.Code, &s.Author, &created, &updated); err != nil {
		return nil, err
	}
	s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
	s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updated)
	return &s, nil
}

// ListStrategiesDB
func ListStrategiesDB() ([]Strategy, error) {
	rows, err := db.Query(`SELECT id,name,description,author,created_at,updated_at FROM strategies ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Strategy
	for rows.Next() {
		var s Strategy
		var created, updated string
		if err := rows.Scan(&s.ID, &s.Name, &s.Desc, &s.Author, &created, &updated); err != nil {
			return nil, err
		}
		s.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", created)
		s.UpdatedAt, _ = time.Parse("2006-01-02 15:04:05", updated)
		out = append(out, s)
	}
	return out, nil
}

// SaveStrategyRunLog 保存一次运行记录（optional）
func SaveStrategyRunLog(strategyID int64, target string, matchesCount int, durationMs int64, errStr string) error {
	_, err := db.Exec(`INSERT INTO strategy_runs(strategy_id, target, matches_count, duration_ms, err, created_at) VALUES(?,?,?,?,?,?)`,
		strategyID, target, matchesCount, durationMs, errStr, time.Now())
	return err
}

// DeleteStrategyDB 根据 id 删除策略
func DeleteStrategyDB(id int64) error {
	_, err := db.Exec(`DELETE FROM strategies WHERE id=?`, id)
	return err
}
