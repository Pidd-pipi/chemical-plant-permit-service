# BUG_REPRO: 页面刷新时间与健康检查计数共享状态竞态

## Bug 是什么

- `backend/web/web.go`：`lastRefreshed` 共享变量写入与读取均不加锁，并发刷新页面触发 data race。
- `backend/health/health.go`：`totalChecks` 计数用普通自增而非原子操作，并发健康检查触发 data race 且计数可能丢失。

## 如何触发

- 多个客户端并发 `GET /` 刷新页面。
- 多个探测并发 `GET /healthz`。

## 错误信息

- `-race` 下 `WARNING: DATA RACE`（web.Handler / health.Handler 内共享变量读写）。
- 并发检查后 `checks` 计数少于实际请求数。
