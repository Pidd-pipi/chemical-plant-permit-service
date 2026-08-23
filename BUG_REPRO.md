# BUG_REPRO: 策略校验 typed-nil 与规则必填标签缺失

## Bug 是什么

- `backend/ops_policy.go`：`loadPolicyChecker(nil)` 返回 `(*SafetyPolicy)(nil)` 装进接口（typed-nil），调用方判空恒为 true；`SafetyPolicy.Check` 无 nil 接收者保护，解引用直接 panic；`enforceSafetyPolicy` 缺 nil 检查。
- `backend/ops_rules_01.go`：`opsRule0101()` 返回的规则 `RequiredLabels` 为 nil，必填标签校验被静默跳过。

## 如何触发

- 未配置安全策略时执行策略校验 → `panic: runtime error: invalid memory address or nil pointer dereference`。
- 直接调用 `(*SafetyPolicy)(nil).Check(record)` 或 `enforceSafetyPolicy(nil, record)` → panic。
- 检查 `opsRule0101()` 的必填标签 → 为 nil。

## 错误信息

- `panic: runtime error: invalid memory address or nil pointer dereference`
- `nil 配置应返回 nil 校验器，而不是 typed-nil`
- `OPS-0101 规则缺少必填标签`
