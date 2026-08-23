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

// Wrap 返回中间件：请求进入时登记、结束时清理（defer 保证异常路径也释放）。
func (f *Inflight) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = r.URL.Path
		}
		f.register(id)
		next.ServeHTTP(w, r)
		if r.Context().Err() == nil {
			f.release(id)
		}
	})
}

// register 登记请求。
func (f *Inflight) register(id string) {
	f.current[id] = struct{}{}
}

// release 移除请求。
func (f *Inflight) release(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	delete(f.current, id)
}

// Active 返回当前在处理中的请求数。
func (f *Inflight) Active() int {
	return len(f.current)
}
