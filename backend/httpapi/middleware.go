package httpapi

import (
	"net/http"
	"sync"
)

// Inflight 记录当前正在处理的请求，便于观察并发压力。
type Inflight struct {
	mu      sync.Mutex
	current map[string]struct{}
}

func NewInflight() *Inflight { return &Inflight{current: map[string]struct{}{}} }

// Wrap 返回中间件：请求进入时登记、结束时清理。
func (f *Inflight) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = r.URL.Path
		}
		f.mu.Lock()
		f.current[id] = struct{}{}
		f.mu.Unlock()
		defer func() {
			f.mu.Lock()
			delete(f.current, id)
			f.mu.Unlock()
		}()
		next.ServeHTTP(w, r)
	})
}

// Active 返回当前在处理中的请求数。
func (f *Inflight) Active() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.current)
}
