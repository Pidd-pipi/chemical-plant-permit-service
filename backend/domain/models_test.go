package domain

import "testing"

func TestSortPermitsByRiskReturnsCopy(t *testing.T) {
	input := []Permit{{ID: "cp-a", Facility: "A", Operation: "op-a", Status: "approved", RiskLevel: "low"}, {ID: "cp-b", Facility: "B", Operation: "op-b", Status: "review", RiskLevel: "high"}}
	sorted := SortPermitsByRisk(input)
	if len(sorted) != 2 || sorted[0].ID != "cp-b" {
		t.Fatalf("排序结果错误: %+v", sorted)
	}
	if input[0].ID != "cp-a" {
		t.Fatalf("入参被原地排序修改: %+v", input)
	}
}
