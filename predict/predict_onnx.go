package main

import (
	"fmt"
	"log"

	ort "github.com/yalue/onnxruntime_go"
)

func main() {
	// 1️⃣ 设置 ONNX Runtime 共享库路径
	ort.SetSharedLibraryPath("/home/sun/onnxruntimearm/lib/libonnxruntime.so.1.22.0")

	// 2️⃣ 初始化环境
	if err := ort.InitializeEnvironment(); err != nil {
		log.Fatal("Failed to initialize ONNX Runtime environment:", err)
	}
	defer ort.DestroyEnvironment()

	// 3️⃣ 构造输入数据
	inputData := []float32{12.36, 12.58, 12.20, 12.40, 5600000}
	inputShape := ort.NewShape(1, 9) // batch=1, 特征数=9
	inputTensor, err := ort.NewTensor(inputShape, inputData)
	if err != nil {
		log.Fatal("Failed to create input tensor:", err)
	}
	defer inputTensor.Destroy()

	// 4️⃣ 创建输出 Tensor
	outputShape := ort.NewShape(1) // 二分类输出 [batch_size]
	outputTensor, err := ort.NewEmptyTensor[int64](outputShape)
	if err != nil {
		log.Fatal("Failed to create output tensor:", err)
	}
	defer outputTensor.Destroy()

	// 5️⃣ 创建 AdvancedSession
	session, err := ort.NewAdvancedSession(
		"./predict/model/lgb_stock_model.onnx", // ONNX 模型路径
		[]string{"input"},                      // 输入节点名（Python 查看 sess.get_inputs()[0].name）
		[]string{"label"},                      // 输出节点名（Python 查看 sess.get_outputs()[0].name）
		[]ort.Value{inputTensor},
		[]ort.Value{outputTensor},
		nil,
	)
	if err != nil {
		log.Fatal("Failed to create session:", err)
	}
	defer session.Destroy()

	// 6️⃣ 执行推理
	if err := session.Run(); err != nil {
		log.Fatal("Inference failed:", err)
	}

	// 7️⃣ 获取输出结果
	pred := outputTensor.GetData()
	fmt.Println("预测结果:", pred)

}
