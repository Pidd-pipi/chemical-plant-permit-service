package store

import (
	"errors"
	"example.com/chemical-plant-permit-service/domain"
	"sync"
)

var ErrNotFound = errors.New("permit not found")

type Store struct {
	mu    sync.RWMutex
	items []domain.Permit
}

func New() *Store {
	return &Store{items: []domain.Permit{{ID: "cp-501", Facility: "西区储运厂", Operation: "酸液转运", Status: "approved", RiskLevel: "high", ExpiresOn: "2026-09-30"}, {ID: "cp-502", Facility: "东区反应车间", Operation: "年度检修", Status: "review", RiskLevel: "medium", ExpiresOn: "2026-10-15"}}}
}

// NewWithItems 使用指定初始数据构造存储。
func NewWithItems(items []domain.Permit) *Store {
	copied := append([]domain.Permit(nil), items...)
	return &Store{items: copied}
}
func (s *Store) List() []domain.Permit {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Permit, len(s.items))
	copy(out, s.items)
	return out
}
func (s *Store) UpdateStatus(id, v string) (domain.Permit, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.items {
		if s.items[i].ID == id {
			s.items[i].Status = v
			return s.items[i], nil
		}
	}
	return domain.Permit{}, ErrNotFound
}
