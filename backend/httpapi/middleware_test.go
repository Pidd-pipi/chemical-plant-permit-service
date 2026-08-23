package httpapi

import (
	"context"
	"example.com/chemical-plant-permit-service/store"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

func TestInflightNoLeakOnCancel(t *testing.T) {
	inflight := NewInflight()
	handler := inflight.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/permits", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if got := inflight.Active(); got != 0 {
		t.Fatalf("取消请求后并发计数未释放: Active=%d", got)
	}
}

func TestInflightReleaseOnPanic(t *testing.T) {
	inflight := NewInflight()
	handler := inflight.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("boom")
	}))
	func() {
		defer func() { _ = recover() }()
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/x", nil)
		handler.ServeHTTP(rec, req)
	}()
	if got := inflight.Active(); got != 0 {
		t.Fatalf("panic 后计数未释放: %d", got)
	}
}

func TestInflightParallelRegister(t *testing.T) {
	inflight := NewInflight()
	handler := inflight.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			<-start
			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			req.Header.Set("X-Request-ID", fmt.Sprintf("req-%d", n))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
		}(i)
	}
	close(start)
	wg.Wait()
	if got := inflight.Active(); got != 0 {
		t.Fatalf("请求结束后应清零, got %d", got)
	}
}

func TestInflightActiveConcurrent(t *testing.T) {
	inflight := NewInflight()
	handler := inflight.Wrap(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			<-start
			req := httptest.NewRequest(http.MethodGet, "/x", nil)
			req.Header.Set("X-Request-ID", fmt.Sprintf("req-%d", n))
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			_ = inflight.Active()
		}(i)
	}
	close(start)
	wg.Wait()
}

func TestStatusRejectsEmpty(t *testing.T) {
	s := store.New()
	server := httptest.NewServer(New(s))
	defer server.Close()
	res, err := http.Post(server.URL+"/api/v1/permits/cp-501/status", "application/json", strings.NewReader(`{"status":""}`))
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 400 {
		t.Fatalf("空状态应返回 400, got %d", res.StatusCode)
	}
}
