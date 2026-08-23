package validation

import "testing"

func TestStatusRejectsEmptyValue(t *testing.T) {
	if err := Status(""); err == nil {
		t.Fatal("空状态应被拒绝")
	}
}
