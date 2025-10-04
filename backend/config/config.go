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
	StockListPath string           `yaml:"stock_list_path"`
	KLineDays     int              `yaml:"kline_days"`
	UpdateHour    int              `yaml:"update_hour"`
	UpdateMinute  int              `yaml:"update_minute"`
	Combination   string           `yaml:"combination"`
	Strategies    []StrategyConfig `yaml:"strategies"`
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
