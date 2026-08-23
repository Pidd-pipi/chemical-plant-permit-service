# BUG_REPRO: 许可列表共享底层数组导致顺序串场与竞态

## Bug 是什么

- `backend/store/store.go`：`List` 直接返回内部切片、`NewWithItems` 保留外部输入切片引用；`UpdateStatus` 不加锁。
- `backend/domain/models.go`：`SortPermitsByRisk` 原地排序并返回原切片。

## 如何触发

- 连续两次请求列表 → 第一次排序把存储内部顺序改掉，第二次顺序与录入顺序不一致。
- 用外部数据初始化存储后修改原始列表 → 库存数据跟着变。
- 并发 GET 列表 + POST 更新状态 → data race。

## 错误信息

- `列表排序污染了存储内部顺序: first=cp-b`
- `List 返回内部切片: got "closed" want "approved"`
- `NewWithItems 保留外部引用: got "closed" want "approved"`
- `入参被原地排序修改`
- `-race` 报 `WARNING: DATA RACE`
