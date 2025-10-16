// predict.go
package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/dmitryikh/leaves"
)

// 简单的特征处理函数，与训练保持一致
func extractFeatures(record []string) ([]float64, error) {
	// 假设 CSV 格式与训练时一致: Date,Open,High,Low,Close,Volume,ma5,ma10,rsi,return
	// 我们只取与训练相同的特征列
	feats := make([]float64, 9)
	for i := 1; i <= 9; i++ {
		v, err := strconv.ParseFloat(record[i], 64)
		if err != nil {
			return nil, err
		}
		feats[i-1] = v
	}
	return feats, nil
}

func main() {
	// 1. 加载 LightGBM 模型
	model, err := leaves.LGBMModelFromFile("model/lgb_stock_model.txt")
	if err != nil {
		log.Fatalf("加载模型失败: %v", err)
	}
	defer model.Free()

	// 2. 读取最新样本数据（例如 data/latest.csv）
	file, err := os.Open("data/latest.csv")
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	_, _ = reader.Read() // 跳过header

	record, _ := reader.Read()
	feats, err := extractFeatures(record)
	if err != nil {
		log.Fatal(err)
	}

	// 3. 预测
	pred := model.PredictSingle(feats, 0)
	fmt.Printf("上涨概率: %.4f\n", pred)
	if pred > 0.5 {
		fmt.Println("👉 预测：明日上涨")
	} else {
		fmt.Println("👎 预测：明日下跌")
	}
}
