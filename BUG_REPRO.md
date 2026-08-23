# BUG_REPRO: 后台巡检定时器泄露、收尾缺失与导出问题

## Bug 是什么

- `backend/ops_worker.go`：`Run` 启动 `time.NewTicker` 后从不 `Stop`（定时器资源泄露）；取消时直接返回不执行收尾清理；间隔配置非正数时 `NewTicker` 直接 panic。
- `backend/ops_export.go`：`BuildPermitRows` 复用包级共享缓冲，多次导出结果互相累积串数据；`WritePermitCSV` 的 deferred `Flush` 吞掉写错误。

## 如何触发

- worker 运行后取消 ctx → 定时器未释放、收尾回调不执行。
- `newOpsWorker(svc, 0)` 后 `Run` → `panic: non-positive interval for NewTicker`。
- 连续两次 `BuildPermitRows` → 第二次行数翻倍。
- 向写入失败的目标 writer 导出 CSV → 不返回错误。

## 错误信息

- `panic: non-positive interval for NewTicker`
- `取消后未执行收尾清理`
- `第二次导出行数应为 2，got 4（共享缓冲累积）`
- `flush 错误被吞掉，应返回错误`
