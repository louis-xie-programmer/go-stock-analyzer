#!/bin/bash

# 项目根目录
ROOT_DIR=$(pwd)

echo "🚀 开始构建 Stock Analyzer 项目..."

# 后端部分
echo "📦 构建 Go 后端..."
cd $ROOT_DIR/backend
go mod init backend >/dev/null 2>&1 || true
go mod tidy
go build -o go-stock-analyzer

# 前端部分
echo "📦 构建 Vue 前端..."
cd $ROOT_DIR/frontend
npm install
npm run build

# 启动服务
echo "✅ 构建完成！"
echo "👉 启动后端 API 服务: ./backend/go-stock-analyzer"
echo "👉 前端静态文件已打包到 frontend/dist/，可用 nginx/静态服务器部署"

# 提供一个一键运行方式（默认前后端同时跑）
echo
read -p "是否直接运行（后端+前端开发模式）？ [y/n] " run_now

if [ "$run_now" == "y" ]; then
  # 启动后端
  echo "🚀 启动后端 API 服务 (端口 :8080)..."
  cd $ROOT_DIR/backend
  ./go-stock-analyzer &
  BACKEND_PID=$!

  # 启动前端 (Vite Dev Server)
  echo "🌐 启动前端服务 (端口 :5173)..."
  cd $ROOT_DIR/frontend
  npm run dev

  # 前端退出时，杀掉后端进程
  kill $BACKEND_PID
fi
