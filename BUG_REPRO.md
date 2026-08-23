# BUG_REPRO: 在途请求计数泄漏与空状态校验缺口

## Bug 是什么

- `backend/httpapi/middleware.go`：`Inflight.Wrap` 只在请求正常结束时释放计数，取消的请求或处理函数 panic 时跳过释放，`current` map 持续增长；`register` 与 `Active` 无锁读写共享 map，并发请求触发 data race。
- `backend/validation/validation.go`：`Status` 放行空字符串，空状态可被提交并写入许可记录，污染状态。

## 如何触发

- 用已取消的 context 发起请求，或让处理函数 panic → `Inflight.Active()` 不归零。
- 并发 `ServeHTTP` + `Active()` → `-race` 报 data race。
- `POST /api/v1/permits/cp-501/status` 携带 `{"status":""}` → 返回 200 并写入空状态（应 400）。

## 错误信息

- `取消请求后并发计数未释放: Active=1`
- `panic 后计数未释放: 1`
- `-race` 下 `WARNING: DATA RACE`
- `空状态应返回 400, got 200`
