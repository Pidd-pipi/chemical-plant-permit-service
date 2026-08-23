package domain

import "sort"

type Permit struct {
	ID        string `json:"id"`
	Facility  string `json:"facility"`
	Operation string `json:"operation"`
	Status    string `json:"status"`
	RiskLevel string `json:"risk_level"`
	ExpiresOn string `json:"expires_on"`
}
type StatusChange struct {
	Status string `json:"status"`
}

// SortPermitsByRisk 返回按风险等级从高到低排序的新切片，不修改入参。
func SortPermitsByRisk(items []Permit) []Permit {
	out := append([]Permit(nil), items...)
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].RiskLevel != out[j].RiskLevel {
			return riskWeight(out[i].RiskLevel) > riskWeight(out[j].RiskLevel)
		}
		return out[i].ID < out[j].ID
	})
	return out
}

func riskWeight(level string) int {
	switch level {
	case "high":
		return 3
	case "medium":
		return 2
	default:
		return 1
	}
}
