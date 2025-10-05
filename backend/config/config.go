package config

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v2"
)

type StrategyConfig struct {
	Name    string                 `yaml:"name"`
	Enabled bool                   `yaml:"enabled"`
	Params  map[string]interface{} `yaml:"params"`
}

type Config struct {
	DBPath        string           `yaml:"db_path"`
	KLineDays     int              `yaml:"kline_days"`
	UpdateHour    int              `yaml:"update_hour"`
	UpdateMinute  int              `yaml:"update_minute"`
	Combination   string           `yaml:"combination"`
	Strategies    []StrategyConfig `yaml:"strategies"`
	// Worker pool settings for startup watchlist KLine fetch
	WorkerConcurrency  int `yaml:"worker_concurrency"`
	WorkerRetries      int `yaml:"worker_retries"`
	WorkerDelayMs      int `yaml:"worker_delay_ms"`
	WorkerBackoffMs    int `yaml:"worker_backoff_ms"`
	WatchlistKlineDays int `yaml:"watchlist_kline_days"`
}

var Cfg Config

func LoadConfig(path string) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	if err := yaml.Unmarshal(data, &Cfg); err != nil {
		log.Fatal(err)
	}
}
