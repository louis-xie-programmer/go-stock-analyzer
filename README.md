# go-stock-analyzer

轻量级 A 股数据抓取、指标计算、策略回测/调度与实时推送示例工程。

本仓库以 Go 作为后端、Vue（Vite）作为前端，使用 SQLite 做为单文件数据库，演示了从数据抓取到策略执行和 Web / WebSocket 推送的完整流程。适合作为学习、二次开发或构建原型的起点。

详细的内容介绍全在微信公众号中。干货持续更新，敬请关注「代码扳手」微信公众号：
<img width="430" height="430" alt="image" src="https://github.com/user-attachments/assets/bf5948e8-5d4f-431e-b1b3-e80a135630a8" />

## 目录结构（摘录）

- `backend/`：后端 Go 服务代码（入口：`backend/main.go`）
  - `config/`：配置加载（`config.yaml`, `config.go`）
  - `fetcher/`：外部行情抓取逻辑（Sina API）、指标计算
  - `storage/`：SQLite 初始化与 CRUD（`db.go`）
  - `strategy/`：示例策略（MA、MACD、DSL、Composite）
  - `strategyexec/`：动态策略执行引擎（基于 yaegi 解释器）
  - `scheduler/`：定时任务调度（拉取 K 线并触发策略）
  - `realtime/`：WebSocket Hub 与 polling 广播逻辑
  - `web/`：HTTP API 路由与处理器
  - `stock.db`：示例 SQLite 数据库文件（运行时生成/更新）

- `frontend/`：Vue + Vite 前端代码（简单 UI，连接后端 REST + WS）
  - `src/`：组件与页面（`StockChart.vue`、`WatchlistView.vue` 等）

根目录包含 `go.mod` / `go.sum`，以及 `bin/backend`（构建产物示例）。

## 项目概述（一句话）

后端定期从新浪等公开接口抓取股票列表和 K 线数据，计算常见技术指标（MA、MACD 等），把结果存入 `SQLite`，同时通过 WebSocket 将实时行情/策略信号推送给前端。前端使用 Vite + Vue 显示行情、K 线与策略回测/结果。

## 快速上手（开发环境，推荐使用 WSL / Linux）

先决条件：Go 1.18+、Node.js (建议 16+)、npm、gcc（构建 sqlite3 cgo 时需要）。在 Windows 上建议通过 WSL 运行后端以避免 CGO 问题。

1) 克隆仓库（已在本 workspace）

2) 启动后端（开发模式）

```bash
# 进入仓库根目录（WSL）
cd /mnt/d/2025/go-stock-analyzer
# 直接运行（会读取 backend/config/config.yaml）
go run ./backend
```

说明：后端入口 `backend/main.go` 启动流程：加载配置 -> 初始化 DB -> 若 `stocks` 表条目少于阈值则调用 `fetcher.FetchAllStocks()` 做全量抓取 -> 启动 realtime poll -> 启动每日 scheduler -> 启动 Gin HTTP 服务（默认 :8080）。

3) 构建后端二进制

```bash
go build -o ./bin/backend ./backend
./bin/backend
```

注意：`github.com/mattn/go-sqlite3` 依赖 CGO，构建时需启用 CGO 并在系统安装 C 编译工具链（如 `gcc`）。在 WSL/Ubuntu 上可以使用 `sudo apt update && sudo apt install build-essential`。

4) 启动前端（开发模式）

```bash
cd frontend
npm install
npm run dev
```

前端默认会连接后端提供的 REST 和 WebSocket：
- REST 示例：`/api/stocks`, `/api/watchlist`, `/api/kline`
- WebSocket：`/ws/realtime`（客户端按 `Quote.Code` 订阅）

如果想让后端提供前端生产静态资源，先在 `frontend` 中执行 `npm run build`，然后把 `dist` 放到 `frontend/dist`，后端在 `backend/web/server.go` 中有将 `../frontend/dist` 暴露为静态文件的路由。

## 核心实现要点（开发者速览）

- 抓取（backend/fetcher）
  - `stock_list.go`：抓取并解析股票列表，生成 `symbol`（示例：`sz000001` / `sh600000`）
  - `fetcher.go`：按 symbol 拉取 K 线数据并解析，抓取后会计算部分指标（MA/MACD）以便策略使用

- 存储（backend/storage/db.go）
  - 初始化 schema（tables: `stocks`, `kline`, `watchlist`, `results` 等）
  - 统一使用 `INSERT OR REPLACE` 做 upsert
  - 已为性能做了 PRAGMA 调优（WAL、synchronous NORMAL）

- 策略（backend/strategy）
  - 提供多种策略实现：`ma_strategy.go`, `macd_strategy.go`, `dsl_strategy.go`, `composite_strategy.go`
  - DSL 使用 `github.com/Knetic/govaluate` 解析表达式，可在前端或配置里输入简单逻辑表达式进行回测

- 调度（backend/scheduler/scheduler.go）
  - 按 `config.yaml` 中设置的时间周期拉取最新 K 线并触发 `strategy.RunAll`

- 实时推送（backend/realtime）
  - WebSocket Hub（`hub.go`）管理连接，`fetcher.go` 负责轮询并把最新 Quote 广播给订阅的客户端

## 常见问题与排查

- 后端构建失败（通常是 sqlite3/CGO 相关）
  - 错误信息通常包含 `gcc`、`cgo` 或 `sqlite3`，在 WSL 上安装 `build-essential` 并确保 `CGO_ENABLED=1` 通常可解决。

- 数据不够或全量抓取没跑
  - 后端在 `InitDB` 时会检查 `stocks` 表条目，若少于配置阈值会调用 `fetcher.FetchAllStocks()`。手动触发可以临时修改或调用相应函数进行测试（开发时直接在 `main.go` 添加触发点）。

- 外部 API 不稳定
  - 抓取模块对 Sina 接口的解析有一定容错（空判断等）。遇到接口变更时优先调整 `backend/fetcher/*` 中的解析逻辑。

## 可扩展项 / 下一步建议

- 增加单元测试与集成测试：当前为演示工程，建议为关键模块（`fetcher`、`storage`、`strategy`）补充自动化测试。
- 把 SQLite 替换为可选的持久层（Postgres / MySQL），并提供迁移脚本。
- 优化抓取并发与限频策略（加入重试/退避、rate limiter）。
- 增强前端交互：策略编辑器、回测可视化、自定义 watchlist 管理。
- 扩展动态策略执行：
  - 为策略提供更多内置函数（技术指标、统计函数）
  - 支持策略之间的组合与引用
  - 添加策略性能分析与优化建议
  - 提供更多策略模板与示例

## 贡献

欢迎 Issue / PR。请保持 commit 小而专注，确保修改伴随必要的说明或测试用例。

## 许可证

仓库默认无明确 LICENSE 文件 — 如需对外开源请添加合适的 LICENSE（推荐 MIT / Apache-2.0）。

---

作者：本仓库为示例/教学用途，主要维护者信息见仓库提交历史。

## 详细示例

下列示例旨在快速帮助你复制环境、调试后端、调用 API 与使用策略 DSL（可直接复制到终端 / 客户端进行测试）。

### 后端配置示例（`backend/config/config.yaml`）
下面是一个合理的配置示例；仓库中的 `config.yaml` 可能略有不同，但字段含义相同：

```yaml
# 数据库文件（相对 backend 路径或绝对路径）
db_path: "backend/stock.db"

# 抓取与调度设置
update_hour: 15       # 每日触发策略的小时（24h）
update_minute: 10     # 每日触发策略的分钟
watchlist_poll_interval_seconds: 5  # 实时推送 polling 间隔（秒）
kline_days: 200       # 抓取历史 K 线天数

# Web 服务
http_port: 8080

# 抓取并发与限频（示例）
fetch_workers: 4
fetch_rate_limit_per_second: 2
```

### 常用环境变量（构建/运行）

- 在 Windows/WSL 上构建时为 sqlite3 启用 CGO（确保已安装 gcc）：

```bash
# 在 WSL/Ubuntu 下
export CGO_ENABLED=1
go build -o ./bin/backend ./backend
```

### REST API 示例

1) 获取股票列表（分页/简单示例）

```bash
curl -s "http://localhost:8080/api/stocks"
```

示例响应（JSON）：

```json
[
  {
    "symbol": "sz000001",
    "code": "000001",
    "name": "平安银行",
    "market": "sz",
    "board": "深证主板"
  },
  ...
]
```

2) 获取单支股票的 K 线（按天）

```bash
curl -s "http://localhost:8080/api/kline?symbol=sz000001&days=60"
```

示例响应（结构化，省略部分字段）：

```json
{
  "symbol": "sz000001",
  "kline": [
    {"date":"2025-10-01","open":10.5,"high":11.0,"low":10.3,"close":10.9,"volume":1234567,"ma5":10.2,"ma10":9.8},
    ...
  ]
}
```

3) Watchlist 与策略结果

```bash
curl -s "http://localhost:8080/api/watchlist"
curl -s "http://localhost:8080/api/results?symbol=sz000001"
```

### WebSocket（实时推送）示例

后端 WebSocket 路径：`ws://localhost:8080/ws/realtime`

客户端连接后通常先发送一条订阅消息（这里以 JSON 为例，后端实现会根据 `Quote.Code` 分配订阅）

示例订阅消息：

```json
{ "action": "subscribe", "symbol": "sz000001" }
```

示例服务器推送（行情/策略信号混合）：

```json
{
  "type": "quote",
  "symbol": "sz000001",
  "price": 10.92,
  "volume": 120000,
  "time": "2025-10-05T15:10:02+08:00"
}

{
  "type": "signal",
  "symbol": "sz000001",
  "strategy": "ma_cross",
  "signal": "buy",
  "meta": {"ma_short":5,"ma_long":20}
}
```

注意：前端实现里 `backend/realtime/client.go` 与 `hub.go` 负责订阅管理与广播，实际字段名与格式请以代码为准。

### 数据库（SQLite）schema 快照

下面给出常见表的简化字段，具体定义在 `backend/storage/db.go` 中（InitDB）：

- `stocks`：
  - `symbol` TEXT PRIMARY KEY
  - `code` TEXT
  - `name` TEXT
  - `market` TEXT
  - `board` TEXT

- `kline`：按日/分钟存储（示例为日线）
  - `symbol` TEXT
  - `date` TEXT (YYYY-MM-DD)
  - `open` REAL
  - `high` REAL
  - `low` REAL
  - `close` REAL
  - `volume` INTEGER
  - `ma5` REAL
  - `ma10` REAL
  - `macd` REAL
  - PRIMARY KEY (`symbol`, `date`)

- `watchlist`：订阅或关注的 symbol 列表
  - `id` INTEGER PRIMARY KEY
  - `symbol` TEXT

- `results`：策略运行结果/信号
  - `id` INTEGER PRIMARY KEY
  - `symbol` TEXT
  - `strategy` TEXT
  - `signal` TEXT
  - `timestamp` TEXT
  - `payload` TEXT (JSON)

可以用 `sqlite3 backend/stock.db` 或 `sqlitebrowser` 打开并检查表数据；常用 SQL：

```sql
-- 最近 10 条 kline
SELECT * FROM kline WHERE symbol='sz000001' ORDER BY date DESC LIMIT 10;

-- 查看 watchlist
SELECT * FROM watchlist;
```

### DSL 策略示例与回测

仓库中的 DSL 策略基于 `govaluate` 表达式解析器，示例表达式：

```
ma5 > ma10 && volume > sma(volume, 20) * 1.5
```

含义：5 日均线上穿 10 日均线，且当前成交量大于 20 日均量的 1.5 倍。

回测流程（简化）

1. 在前端 DSL 编辑器输入表达式并保存。
2. 后端 `strategy.RunAll` 会读取 DSL 策略并对 `kline` 数据序列逐日评估表达式（实现细节在 `backend/strategy/dsl_strategy.go`）。
3. 将命中的信号写入 `results` 表，并（可选）通过 WebSocket 推送给在线客户端。

示例：在前端提供回测按钮会触发 API（或在本地调用策略函数）生成回测报告，报告包含命中日期、收益统计与基本可视化数据。

### 动态策略执行（Go 源码）

除了 DSL 表达式，系统还支持直接执行 Go 源码形式的策略。这种方式更灵活，可以使用完整的 Go 语言特性，适合复杂策略的实现。

策略需要实现以下函数签名（在前端编辑器或配置中提供源码）：

```go
// Match 函数接收股票代码和 K 线数据，返回是否满足策略条件
func Match(symbol string, klines []map[string]interface{}) bool {
    // 在这里实现策略逻辑
    if len(klines) < 20 {
        return false
    }
    
    // 示例：判断最近收盘价是否连续上涨
    latest := klines[len(klines)-1]["Close"].(float64)
    prev := klines[len(klines)-2]["Close"].(float64)
    prevPrev := klines[len(klines)-3]["Close"].(float64)
    
    return latest > prev && prev > prevPrev
}
```

系统使用 yaegi 解释器在安全的环境中执行策略代码。主要特点：

1. 超时控制
   - 总执行超时（默认 30s）
   - 每只股票超时（默认 800ms）

2. 安全性
   - 代码在隔离环境中执行
   - 运行时错误不会影响主程序
   - panic 会被捕获并跳过当前股票

3. 上下文与数据
   - K 线数据以 map 形式提供
   - 支持访问标准库函数
   - 可以在策略中实现技术指标计算

使用示例：

```go
code := `
func Match(symbol string, klines []map[string]interface{}) bool {
    // 你的策略逻辑
    return true
}
`
matches, err := strategyexec.ExecuteStrategy(code, symbols, 60, storage.GetKLines, strategyexec.DefaultExecConfig)
```

注意：动态执行会比预编译策略稍慢，建议在回测场景使用，实时信号生成优先使用预编译策略。

### 常用调试命令与快速检查

- 检查后端是否成功监听端口（Linux/WSL）：

```bash
ss -ltnp | grep 8080
```

- 在没有前端时手动发送一条测试 WebSocket（可用 websocat 或 node）

```bash
# 安装 websocat（Linux）
sudo apt install websocat
websocat ws://localhost:8080/ws/realtime
```

- 清理并重建 DB（慎用，会丢数据）：

```bash
rm -f backend/stock.db
go run ./backend
```

### 调试与常见错误定位

- 若遇到 `database is locked`：查看是否有多个进程访问同一 sqlite 文件并发写入，或 WAL 模式未正确生效。
- 若抓取返回空数据：检查外部 API 是否可访问（网络问题或接口变更），可在 `backend/fetcher` 中增加更严格的日志以定位解析失败行。

## 总结

本文档补充了配置样例、REST/WS 使用样例、数据库 schema 快照、DSL 策略示例与常用调试命令，便于复制运行和快速调试代码。若你希望我把其中的某些示例变成可运行的脚本（例如自动化回测脚本、短小的 Postman 集合或前端 Mock），我可以继续实现并在仓库中添加对应文件。
