package web

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func TestWebConcurrentRequests(t *testing.T) {
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
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
			if rec.Header().Get("X-Refreshed-At") == "" {
				results <- -1
				return
			}
			results <- rec.Code
		}()
	}
	close(start)
	wg.Wait()
	for i := 0; i < n; i++ {
		if code := <-results; code != 200 {
			t.Fatalf("页面应返回 200 且带刷新时间头, code=%d", code)
		}
	}
}
