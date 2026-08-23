package store

import (
	"example.com/chemical-plant-permit-service/domain"
	"testing"
)

func TestStoreListReturnsCopy(t *testing.T) {
	s := NewWithItems([]domain.Permit{{ID: "cp-x", Facility: "X", Operation: "op", Status: "approved", RiskLevel: "low"}})
	items := s.List()
	items[0].Status = "closed"
	again := s.List()
	if again[0].Status != "approved" {
		t.Fatalf("List 返回内部切片: got %q want %q", again[0].Status, "approved")
	}
}

func TestNewWithItemsCopies(t *testing.T) {
	input := []domain.Permit{{ID: "cp-y", Facility: "Y", Operation: "op", Status: "approved", RiskLevel: "low"}}
	s := NewWithItems(input)
	input[0].Status = "closed"
	if got := s.List()[0].Status; got != "approved" {
		t.Fatalf("NewWithItems 保留外部引用: got %q want %q", got, "approved")
	}
}
