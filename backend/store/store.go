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
// 为避免外部切片与存储内部状态共享底层数组（调用方后续改动会污染库存），
// 这里对传入数据做一次深拷贝，使库存与外部切片彼此独立。
func NewWithItems(items []domain.Permit) *Store {
	out := make([]domain.Permit, len(items))
	copy(out, items)
	return &Store{items: out}
}

// List 返回库存中所有许可的副本。
// 返回副本而非内部切片，确保调用方排序/改写返回值都不会影响库存，
// 多次刷新的结果彼此独立、可安全并发。
func (s *Store) List() []domain.Permit {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Permit, len(s.items))
	copy(out, s.items)
	return out
}

// UpdateStatus 更新指定许可的状态并加写锁，避免与并发的 List 读写竞争。
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
