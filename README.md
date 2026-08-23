# 化工厂作业许可服务

标准库实现的化工作业许可 HTTP 服务，默认监听 `8080`，可由 `PORT` 环境变量调整。提供 `GET /healthz`、`GET /api/v1/permits`、`POST /api/v1/permits/{id}/status`，根路径提供内置页面。

状态值包括 `draft`、`review`、`approved`、`suspended`、`closed`。

## Verification

在 `backend` 目录执行 `gofmt -w .`、`go build ./...`、`go test ./...`，结果全部通过；HTTP 定向测试通过。

## Directory Tree and API

```text
backend/       Go 模块、HTTP 服务、领域包和内嵌 web 静态资源
database/      数据库说明
output/        验证记录
README.md      项目说明
prompt.txt     任务提示
runtime_smoke.json  运行冒烟配置
```

健康检查：`GET /healthz`。API：`GET /api/v1/permits`、`POST /api/v1/permits/{id}/status`；页面入口：`GET /`。

真实服务使用端口 `18105` 启动后验证：`GET /healthz` 返回 `{"service":"chemical-plant-permit","status":"ok"}`；`GET /api/v1/permits` 返回 2 条记录；`POST /api/v1/permits/cp-501/status` 将状态更新为 `suspended` 并返回 200；非法状态返回 400；未知许可返回 404；`/` 返回 785 字节，`/app.js` 返回 835 字节。验证后服务已关闭。

## Engineering Notes

化工作业许可流程代码按领域模型、校验、状态转换、并发安全存储、审计事件和 HTTP 生命周期分层。请求会保留请求标识并经过恢复与超时保护；状态写入使用版本校验，错误通过可识别的领域错误返回。

除现有接口回归测试外，项目还保留可复用的分页、过滤、策略、工作流和运行健康能力，便于后续扩展而不把业务规则堆积到处理器中。
