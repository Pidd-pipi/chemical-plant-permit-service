# BUG_REPRO: 错误链断裂导致 404/409 被映射成 500

## Bug 是什么

- `backend/ops_service.go`：`Get` 与 `Transition` 用 `%v` 包装底层错误，丢失错误链。
- `backend/ops_errors.go`：`wrapOps` 丢弃 `Cause`，错误链断链。
- `backend/ops_http.go`：`opsStatusFromError` 的状态映射漏掉 `not_found` 分支。

## 如何触发

- 查询不存在的许可 → `errors.Is(err, ErrOpsNotFound)` 为 false，状态映射返回 500（应为 404）。
- 重复创建已存在记录 → 错误链断裂，无法识别为冲突。
- 非法状态流转 → 无法识别为 transition 错误。

## 错误信息

- `TestOpsGetNotFoundChain: 错误链断裂: errors.Is(err, ErrOpsNotFound)=false, err=load nope-000: operations record not found`
- `opsStatusFromError(ErrOpsNotFound)=500 want 404`
- `TestOpsTransitionErrorChain: errors.Is(err, ErrOpsTransition)=false`
