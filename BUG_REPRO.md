# BUG_REPRO: 批量审批挂死、取消不生效与并发覆盖

## Bug 是什么

- `backend/ops_batch.go`：批量流转 `Run` 中错误写入无缓冲且无人读取的 `errCh`，失败条目导致 worker 永久阻塞、任务挂死；取消时提前返回不等待不排空，worker 残留；流转时丢失版本号（`expected=0`），同一记录并发流转可两边都成功。
- `backend/ops_clock.go`：`opsDelay` 用 `time.Sleep` 忽略 ctx 取消，取消后仍硬等满时长。

## 如何触发

- 批量中包含不存在的许可号 → 任务挂死，2 秒内不返回。
- 批量运行前取消 ctx → goroutine 数量持续增长。
- 同一记录放入同一批次并发流转 → 无版本冲突，两边都成功。
- 已取消的 ctx 调用 `opsDelay` → 仍硬等满时长。

## 错误信息

- 失败分支：`批量任务在失败分支挂死`（2 秒超时）。
- 取消后：`goroutine 泄露: before=N after=N+K`。
- 并发流转：`并发流转同一记录应产生版本冲突`。
- `opsDelay 忽略取消，硬等 10 秒`。
