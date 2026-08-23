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

// SortPermitsByRisk 按风险等级从高到低排序。
func SortPermitsByRisk(items []Permit) []Permit {
	sort.SliceStable(items, func(i, j int) bool {
		if items[i].RiskLevel != items[j].RiskLevel {
			return riskWeight(items[i].RiskLevel) > riskWeight(items[j].RiskLevel)
		}
		return items[i].ID < items[j].ID
	})
	return items
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
