package main

import (
	"fmt"
	"strings"
)

// SafetyConfig 安全策略配置；缺省时使用默认策略。
type SafetyConfig struct {
	GasCheckRequired bool
	PPERequired      []string
	ConfinedSpace    bool
}

// SafetyPolicy 作业许可安全策略：按作业类型检查气体检测、防护装备等要求。
type SafetyPolicy struct {
	GasCheckRequired bool
	PPERequired      []string
	ConfinedSpace    bool
}

// PolicyChecker 校验一条许可记录是否满足安全策略。
type PolicyChecker interface {
	Check(record OpsRecord) []string
}

// NewSafetyPolicy 从配置构造策略。
func NewSafetyPolicy(cfg *SafetyConfig) *SafetyPolicy {
	if cfg == nil {
		cfg = &SafetyConfig{}
	}
	ppe := make([]string, len(cfg.PPERequired))
	copy(ppe, cfg.PPERequired)
	return &SafetyPolicy{GasCheckRequired: cfg.GasCheckRequired, PPERequired: ppe, ConfinedSpace: cfg.ConfinedSpace}
}

// loadPolicyChecker 加载策略校验器；nil 配置时返回空校验器。
func loadPolicyChecker(cfg *SafetyConfig) PolicyChecker {
	if cfg == nil {
		return (*SafetyPolicy)(nil)
	}
	return NewSafetyPolicy(cfg)
}

// Check 返回违反策略的描述列表；空列表表示通过。
func (p *SafetyPolicy) Check(record OpsRecord) []string {
	violations := []string{}
	if p.GasCheckRequired && strings.TrimSpace(record.LabelValue("gas")) == "" {
		violations = append(violations, "gas check missing")
	}
	if p.ConfinedSpace && strings.TrimSpace(record.LabelValue("confined")) == "" {
		violations = append(violations, "confined space approval missing")
	}
	for _, item := range p.PPERequired {
		if !strings.Contains(strings.ToLower(record.LabelValue("ppe")), strings.ToLower(item)) {
			violations = append(violations, fmt.Sprintf("ppe %s missing", item))
		}
	}
	return violations
}

// enforceSafetyPolicy 执行策略校验，返回违反项错误或 nil。
func enforceSafetyPolicy(checker PolicyChecker, record OpsRecord) error {
	if violations := checker.Check(record); len(violations) > 0 {
		return fmt.Errorf("%w: %s", ErrOpsPolicy, strings.Join(violations, "; "))
	}
	return nil
}

// enforceRuleLabels 检查记录是否满足规则的必填标签，返回缺失项。
func enforceRuleLabels(rule OpsRule, record OpsRecord) []string {
	missing := []string{}
	for _, label := range rule.RequiredLabels {
		if strings.TrimSpace(record.LabelValue(label)) == "" {
			missing = append(missing, label)
		}
	}
	return missing
}
