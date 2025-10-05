## 快速概览（对 AI 代理）

这是一个轻量的 A 股数据抓取、指标计算、策略回测/调度与实时推送的全栈演示工程。
主要语言：Go（后端） + Vue/Vite（前端）。数据库为 SQLite（单文件，位于 `backend/stock.db`，路径可由 `backend/config/config.yaml` 配置）。

下面的要点足够让一个自动化编码代理快速定位修改点并安全运行项目。

---

## 一句话架构梳理

- 启动入口：`backend/main.go`。顺序：加载 config -> 初始化 DB -> 若股票表不足则调用 `fetcher.FetchAllStocks()` -> 启动 realtime poll -> 启动每日 scheduler -> 启动 web 服务（Gin，:8080）。
- 数据流：fetcher (Sina API) -> storage (SQLite) -> scheduler/strategy -> web(realtime + HTTP API) -> frontend (Vue)。

---

## 运行 / 构建（最常用操作）

- 在仓库根目录运行后端（Go 环境要求）：
  - 开发直接：`go run ./backend`（以 repo 根为工作目录，保证相对路径 `backend/config/config.yaml` 可读）
  - 构建：`go build -o bin/backend ./backend`，然后执行 `./bin/backend`
  - 注意：项目依赖 `github.com/mattn/go-sqlite3`，该包需要 CGO（本地 C 编译器，如 gcc）启用。若构建失败，请在开发环境（WSL/Unix）安装 `build-essential` 并确保 `CGO_ENABLED=1`。

- 前端（开发模式）：
  - 进入 `frontend`，安装依赖并启动：`npm install`，`npm run dev`（vite）。
  - 生产构建后端可通过 `backend/web` 的静态路由 `r.Static("/static", "../frontend/dist")` 提供静态资源，需将 `frontend/dist` 放在 `backend` 相对路径下。

---

## 关键目录与文件（快速定位）

- `backend/main.go`：程序入口，说明了启动顺序和默认阈值（例如少于 100 条股票会触发全量抓取）。
- `backend/config/config.yaml`、`backend/config/config.go`：运行时配置（DB 路径、kline 天数、调度时间、策略开关）。
- `backend/fetcher/*`：所有外部数据抓取逻辑（Sina 股票列表与 K 线），并在抓取后计算简单指标（MA、MACD）。示例：`FetchAllStocks()`、`FetchKLine()`。
- `backend/storage/db.go`：SQLite schema、CRUD、核心约定（表名：stocks、watchlist、kline、results）。注意：使用 `INSERT OR REPLACE` 做 upsert，PRAGMA 已做性能优化（WAL、synchronous NORMAL）。
- `backend/scheduler/scheduler.go`：按 `config.Cfg.UpdateHour/Minute` 定时拉取 K 线并触发 `strategy.RunAll`。
- `backend/realtime/*`：WebSocket hub（`/ws/realtime`），以及 polling/broadcast 逻辑，客户端按 `Quote.Code` 订阅。
- `backend/web/server.go`：HTTP API 路由（例如 `/api/stocks`, `/api/watchlist`, `/api/kline`）以及启动 Gin 服务。
- `backend/strategy/*`：策略实现（MA、MACD、DSL、Composite）。DSL 使用 `github.com/Knetic/govaluate` 解析表达式。

---

## 项目特有约定与模式（对修改行为很重要）

- 符号格式：项目使用 `symbol`（例如 `sz000001` / `sh600000`）与 `code`（纯数字代码）并区分 `market`（前缀）与 `board`（通过 `classifyBoard` 分配为上证主板/深证主板/创业板/科创板）。修改 symbol 相关逻辑请同时检查 `fetcher` 和 `storage`。
- Watchlist 驱动：调度器与 realtime 的 polling 都以 `storage.GetWatchlist()` 返回的 `symbol` 列表为准。改变被分析或推送的股票请修改 watchlist 表或对应 API 调用。
- 数据保存：K 线、股票信息通过 `INSERT OR REPLACE` 持久化，schema 在 `storage.InitDB` 中定义。直接修改表结构时请同时更新 InitDB。
- 外部 API 偶发性：fetcher 直接请求新浪接口并做简单容错（空判断、单引号转双引号等），网络或解析错误是常见失败点，建议在修改时尽量保持宽容的解析逻辑。

---

## 集成点与外部依赖（需要注意的地方）

- 外部数据源：新浪财经相关接口（股票列表与 K 线）。抓取代码：`backend/fetcher/stock_list.go`, `backend/fetcher/fetcher.go`。
- 数据库：SQLite 文件 `backend/stock.db`（默认）。使用 sqlite3 驱动需要 CGO，CI/容器环境需装 gcc。调试时可以直接用 `sqlite3 backend/stock.db` 查看表。
- Web/Realtime：后端提供 REST + WebSocket（Gin + gorilla/websocket）。前端使用 axios + WebSocket 连接 `/ws/realtime` 推送。
- 策略 DSL：`backend/strategy/dsl_strategy.go` 使用 `govaluate`，策略参数来自 `config.yaml`，策略结果通过 `storage.SaveResult` 写入 `results` 表。

---

## 典型修改场景示例（针对 AI 代理）

- 增加一个新的 HTTP API：在 `backend/web/server.go` 上添加路由，编写 handler（参考现有 `GetStocksHandler`），如需 DB 访问使用 `storage` 包导出的函数。
- 修改抓取逻辑的容错：在 `backend/fetcher/*` 中增强 JSON 解析或对 API 速限的退避，注意保持 `FetchAllStocks()` 的分页与 `classifyBoard()` 的符号规则。
- 添加新策略：实现 `strategy` 包下新文件，实现与现有策略相同的注册/调用方式（参考 `RunAll` 调用）。策略输出需调用 `storage.SaveResult`。

---

## 调试与诊断提示

- 运行时日志：后端大量使用 `log.Printf` 输出关键步骤（初始化、抓取失败、保存失败等），可直接在终端观察。
- 数据层验证：用 `sqlite3 backend/stock.db` 或任意 SQLite 浏览器查看 `stocks` / `kline` 表。
- 快速触发分析：最简单方式是启动整个后端（`go run ./backend`），scheduler 会在配置时间触发；要立即手动触发，可以临时调用 `strategy.RunAll`（修改代码后编译运行）或把 `StartDailyTask` 的逻辑抽成可复用函数并调用。

---

如果本文件中有任何不清楚或需要更多细节（例如 CI、容器化建议、或某些文件的具体行为），请告诉我你想补充的用例，我会基于仓库代码迭代更新此文件。
