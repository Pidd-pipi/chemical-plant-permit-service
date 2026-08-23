package main

import "testing"

func TestLoadPolicyCheckerNilConfig(t *testing.T) {
	checker := loadPolicyChecker(nil)
	if checker != nil {
		t.Fatal("nil 配置应返回 nil 校验器，而不是 typed-nil")
	}
}

func TestEnforcePolicyNilConfigNoPanic(t *testing.T) {
	checker := loadPolicyChecker(nil)
	rec := OpsRecord{ID: "p-1", Labels: map[string]string{"site": "x"}}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("enforceSafetyPolicy 不应 panic: %v", r)
		}
	}()
	_ = enforceSafetyPolicy(checker, rec)
}

func TestSafetyPolicyCheckNilReceiver(t *testing.T) {
	var p *SafetyPolicy
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("nil 接收者 Check 不应 panic: %v", r)
		}
	}()
	if v := p.Check(OpsRecord{}); v != nil {
		t.Fatalf("nil 接收者应返回 nil 违规项, got %v", v)
	}
}

func TestEnforcePolicyNilCheckerSafe(t *testing.T) {
	rec := OpsRecord{ID: "p-2", Labels: map[string]string{"site": "x"}}
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("enforceSafetyPolicy(nil) 不应 panic: %v", r)
		}
	}()
	_ = enforceSafetyPolicy(nil, rec)
}

func TestRuleRequiredLabelsNotNull(t *testing.T) {
	rule := opsRule0101()
	if rule.RequiredLabels == nil {
		t.Fatal("OPS-0101 规则缺少必填标签")
	}
	missing := enforceRuleLabels(rule, OpsRecord{ID: "r-1", Labels: map[string]string{}})
	if len(missing) == 0 {
		t.Fatal("缺失必填标签的记录应产生违规项")
	}
}
