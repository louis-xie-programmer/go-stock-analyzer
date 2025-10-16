# train_lightgbm.py
import os
import sys
import pandas as pd
import numpy as np
import joblib
import lightgbm as lgb
from sklearn.model_selection import train_test_split
from sklearn.metrics import accuracy_score

# ✅ 关键：用 onnxmltools 转换 LightGBM 模型
from onnxmltools import convert_lightgbm
from onnxmltools.utils import save_model
from onnxmltools.convert.common.data_types import FloatTensorType

def make_features(df):
    df = df.copy()
    df["return"] = df["Close"].pct_change()
    df["ma5"] = df["Close"].rolling(5).mean()
    df["ma10"] = df["Close"].rolling(10).mean()
    delta = df["Close"].diff()
    up = delta.clip(lower=0).rolling(14).mean()
    down = (-delta.clip(upper=0)).rolling(14).mean()
    df["rsi"] = 100 - 100 / (1 + (up / (down + 1e-9)))
    df = df.dropna().reset_index(drop=True)
    return df

def main(csv_path="data/stock_history.csv"):
    if not os.path.exists(csv_path):
        raise SystemExit(f"数据文件不存在: {csv_path}")

    df = pd.read_csv(csv_path)
    df = df.sort_values("Date").reset_index(drop=True)
    df = make_features(df)

    df["label"] = (df["Close"].shift(-1) > df["Close"]).astype(int)
    df = df.dropna().reset_index(drop=True)

    feature_cols = ["Open", "High", "Low", "Close", "Volume", "ma5", "ma10", "rsi", "return"]
    X = df[feature_cols].astype(float)
    y = df["label"].astype(int)

    X_train, X_test, y_train, y_test = train_test_split(X, y, shuffle=False, test_size=0.2)

    # ✅ 使用 LightGBM 原生训练接口
    train_data = lgb.Dataset(X_train, label=y_train)
    valid_data = lgb.Dataset(X_test, label=y_test)

    params = {
        "objective": "binary",
        "metric": "binary_error",
        "learning_rate": 0.05,
        "num_leaves": 31,
        "verbose": -1,
    }

    model = lgb.train(params, train_data, valid_sets=[valid_data], num_boost_round=200)

    # 验证准确率
    y_pred = (model.predict(X_test) > 0.5).astype(int)
    acc = accuracy_score(y_test, y_pred)
    print(f"✅ Test accuracy: {acc:.4f}")

    os.makedirs("model", exist_ok=True)

    # 保存 LightGBM 模型（文本格式 + Booster）
    model.save_model("model/lgb_stock_model.txt")
    joblib.dump(model, "model/lgb_stock_model.joblib")
    print("✅ 模型保存完成")

    # 2) 初始化 input 类型
    n_features = X_train.shape[1]
    initial_types = [("float_input", FloatTensorType([None, n_features]))]

    # 3) 转换 LightGBM Booster 为 ONNX
    onnx_model = convert_lightgbm(
        model,
        name="LGBMClassifier",
        initial_types=initial_types,
        target_opset=15
    )

    # 4) 保存 ONNX
    onnx_path = "model/lgb_stock_model.onnx"
    save_model(onnx_model, onnx_path)
    print(f"✅ 导出 ONNX 成功 -> {onnx_path}")

if __name__ == "__main__":
    csv = sys.argv[1] if len(sys.argv) > 1 else "data/stock_history.csv"
    main(csv)
