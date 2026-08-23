package httpapi

import (
	"encoding/json"
	"example.com/chemical-plant-permit-service/domain"
	"example.com/chemical-plant-permit-service/store"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func getPermitIDs(t *testing.T, url string) []string {
	t.Helper()
	res, err := http.Get(url + "/api/v1/permits")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body struct {
		Items []domain.Permit `json:"items"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	ids := make([]string, 0, len(body.Items))
	for _, item := range body.Items {
		ids = append(ids, item.ID)
	}
	return ids
}

func TestPermitListOrderStable(t *testing.T) {
	s := store.NewWithItems([]domain.Permit{
		{ID: "cp-a", Facility: "A", Operation: "op-a", Status: "approved", RiskLevel: "low"},
		{ID: "cp-b", Facility: "B", Operation: "op-b", Status: "review", RiskLevel: "high"},
	})
	server := httptest.NewServer(New(s))
	defer server.Close()
	ids := getPermitIDs(t, server.URL)
	if len(ids) != 2 {
		t.Fatalf("want 2 permits, got %d", len(ids))
	}
	// 存储内部顺序必须保持录入顺序（cp-a 在前），不能被列表排序污染
	internal := s.List()
	if internal[0].ID != "cp-a" {
		t.Fatalf("列表排序污染了存储内部顺序: first=%s", internal[0].ID)
	}
}

func TestPermitConcurrentListAndUpdate(t *testing.T) {
	s := store.NewWithItems([]domain.Permit{
		{ID: "cp-a", Facility: "A", Operation: "op-a", Status: "approved", RiskLevel: "low"},
		{ID: "cp-b", Facility: "B", Operation: "op-b", Status: "review", RiskLevel: "high"},
	})
	handler := New(s)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			req := httptest.NewRequest(http.MethodGet, "/api/v1/permits", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		req := httptest.NewRequest(http.MethodPost, "/api/v1/permits/cp-b/status", strings.NewReader(`{"status":"closed"}`))
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}()
	close(start)
	wg.Wait()
}
