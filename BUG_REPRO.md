# BUG_REPRO: 关停超时配置缺失且关停可无限挂起

## Bug 是什么

- `backend/config/config.go`：`Config` 结构体缺少 `ShutdownTimeout` 字段，`Load` 不返回关停超时（默认 0）。
- `backend/runtime.go`：`shutdownServer` 忽略传入 timeout 直接用 `context.Background()` 关停；`requestTimeoutMiddleware` 用 `context.Background()` 创建超时上下文，丢弃父请求的取消传播。

## 如何触发

- 存在慢请求（如 2 秒）时调用 `shutdownServer(srv, 200ms)` → 关停无限等待，1 秒内不返回。
- 父请求 context 取消后调用 `requestTimeoutMiddleware` 包裹的 handler → 下游收到的上下文无取消信号。

## 错误信息

- `TestConfigShutdownTimeoutDefault: 缺省关停超时应为 10s, got 0s`
- `TestShutdownServerRespectsTimeout: 关停忽略超时，无限挂起`
- `TestRequestTimeoutPropagatesCancel: 请求取消未向下游传播`
