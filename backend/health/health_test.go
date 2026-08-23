package health

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestHealthParallelChecks(t *testing.T) {
	totalChecks = 0
	handler := http.HandlerFunc(Handler)
	start := make(chan struct{})
	var wg sync.WaitGroup
	const n = 20
	results := make(chan int, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
			var body map[string]any
			_ = json.NewDecoder(rec.Body).Decode(&body)
			results <- rec.Code
		}()
	}
	close(start)
	wg.Wait()
	for i := 0; i < n; i++ {
		if code := <-results; code != 200 {
			t.Fatalf("healthz 应返回 200, got %d", code)
		}
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	var body map[string]any
	_ = json.NewDecoder(rec.Body).Decode(&body)
	if got := int(body["checks"].(float64)); got != n+1 {
		t.Fatalf("检查次数计数错误: got %d want %d", got, n+1)
	}
}
