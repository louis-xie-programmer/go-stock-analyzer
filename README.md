# go-stock-analyzer-realtime

## 项目简介

本项目是一个基于 Go 语言和 Vue3 前后端分离的股票分析与实时行情订阅系统。支持多种选股策略（MA、MACD、复合、DSL），可自动定时更新数据，支持实时行情 WebSocket 订阅，并提供 DSL 策略表达式测试器。

## 目录结构

```
go-stock-analyzer-realtime-fixed/
├── backend/      # Go 后端服务
│   ├── main.go   # 后端入口
│   ├── config/   # 配置相关
│   ├── fetcher/  # 数据抓取与指标计算
│   ├── realtime/ # 实时行情 WebSocket
│   ├── scheduler/# 定时任务
│   ├── storage/  # 数据库操作
│   ├── strategy/ # 选股策略
│   └── web/      # HTTP API 与静态资源服务
├── frontend/     # Vue3 前端
│   ├── src/      # 前端源码
│   ├── index.html
│   ├── package.json
│   └── vite.config.js
├── go.mod
├── go.sum
└── .gitignore
```

## 主要功能

- **选股策略**：支持 MA、MACD、复合策略、DSL表达式自定义。
- **定时任务**：每日定时自动更新股票数据并运行策略。
- **实时行情**：WebSocket 订阅股票实时行情。
- **DSL测试器**：前端可输入 DSL 表达式测试选股逻辑。
- **结果展示**：前端展示符合策略的股票列表。

## 后端架构说明

- 使用 [Gin](https://github.com/gin-gonic/gin) 提供 RESTful API 和 WebSocket 服务。
- 数据存储采用 SQLite，自动建表。
- 通过 `config.yaml` 配置策略、股票列表、定时参数等。
- 主要模块：
  - `fetcher`：抓取新浪行情数据，计算 MA、MACD 等指标。
  - `strategy`：实现多种选股策略，支持 DSL 表达式。
  - `realtime`：WebSocket 推送实时行情，支持订阅/退订。
  - `scheduler`：定时任务，每日自动更新数据并运行策略。
  - `storage`：数据库读写，保存 K线与选股结果。
  - `web`：API 路由与静态资源服务。

## 前端架构说明

- 使用 Vue3 + Vite 构建 SPA。
- 主要页面：
  - 选股结果展示
  - DSL策略测试器
  - 实时行情订阅
- 通过 axios 调用后端 API，通过 WebSocket 订阅实时行情。

## 配置文件说明

`backend/config/config.yaml` 示例：

```yaml
db_path: "backend/stock.db"
stock_list_path: "stock_list.csv"
kline_days: 120
update_hour: 15
update_minute: 0
combination: "all"
strategies:
  - name: "MA"
    enabled: true
    params:
      ma: 20
      hold_days: 3
  - name: "MACD"
    enabled: true
  - name: "Composite"
    enabled: true
    params:
      hold_days: 3
  - name: "DSL"
    enabled: true
    params:
      expr: "close > ma20 AND macd_dif > macd_dea"
```

## 启动方式

### 后端

1. 安装 Go 1.25+，并确保 `go mod tidy` 安装依赖。
2. 准备股票列表文件（如 `stock_list.csv`），每行一个股票代码。
3. 启动后端：

```bash
go run backend/main.go
```

后端默认监听 `8080` 端口，API 路径为 `/api/*`，WebSocket 路径为 `/ws/realtime`。

### 前端

1. 进入 `frontend` 目录，安装依赖：

```bash
npm install
```

2. 启动开发服务器：

```bash
npm run dev
```

前端通过代理 `/api` 到后端，开发时访问 `http://localhost:5173`。

## 主要依赖

- 后端：Gin、Gorilla WebSocket、govaluate、go-sqlite3、yaml.v2
- 前端：Vue3、Vue Router、Axios、Vite

## API 说明

- `GET /api/results`：获取选股结果
- `POST /api/dsl/test`：DSL策略测试
- `GET /ws/realtime`：WebSocket 实时行情订阅

## 示例

- DSL表达式示例：`close > ma20 AND macd_dif > macd_dea`
- 订阅实时行情：在前端输入股票代码（如 `sz000001`），点击订阅即可收到实时数据。

## 贡献与许可

欢迎提交 issue 和 PR。代码遵循 MIT 许可协议。
