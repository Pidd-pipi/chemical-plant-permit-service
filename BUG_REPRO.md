# BUG_REPRO: 复核（review）中间态流转与统计错位

## Bug 是什么

- `backend/ops_state.go`：状态机转换表缺少 `queued→review`、`review→paused`、`review→closed` 等边，记录无法进入或离开复核状态（`ErrOpsTransition`）。
- `backend/ops_model.go` / `backend/ops_query.go`：`InProgress`、`opsInProgress`、`opsCountInProgress` 未把 `review` 计入办理中，统计漏数。

## 如何触发

- 调用 `Transition(id, revision, OpsStatusReview, ...)` 从 `queued` 进入复核 → 被拒绝。
- 统计办理中记录数时，`review` 状态不被计入。

## 错误信息

- `queued -> review 应被允许: operations status transition is not allowed: queued to review`
- `review 状态应算办理中`
- `办理中统计应为 3, got 2`
