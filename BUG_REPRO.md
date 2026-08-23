# BUG_REPRO: 许可台账并发读写竞态与引用串扰

## Bug 是什么

- `backend/ops_store.go`：`Get`/`List` 直接返回内部记录引用（未深拷贝），读路径 `Count` 不加锁；外部拿到记录修改 Labels 会污染库存，多 goroutine 并发读写触发 data race（`concurrent map read and map write`）。
- `backend/ops_audit.go`：`For`/`Since`/`Latest` 返回的事件浅拷贝共享 `Details` map，调用方修改详情会污染审计记录。

## 如何触发

并发调用 `OpsStore.Get`/`List`/`Count` 与 `OpsStore.Update`，并修改取回记录的 `Labels`；并发读取审计事件的 `Details`。

## 错误信息

- `-race` 下报 `WARNING: DATA RACE`（`concurrent map read and map write`）。
- 断言 `Get`/`List` 返回记录与内部存储共享 Labels 时失败：`Get 返回的记录与内部存储共享 Labels: got "mutated" want "west"`。
- `For 返回的事件与审计内部共享 Details`。
