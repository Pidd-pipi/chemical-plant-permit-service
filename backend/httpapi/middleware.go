package httpapi

import (
	"net/http"
	"sync"
)

// Inflight 记录当前正在处理的请求，便于观察并发压力。
type Inflight struct {
	mu      sync.Mutex
	current map[string]int
}

func NewInflight() *Inflight { return &Inflight{current: map[string]int{}} }

// Wrap 返回中间件：请求进入时登记、结束时清理。
// 释放使用 defer，保证请求被取消或处理函数 panic 时也能成对递减，
// 避免计数只增不减或长时间挂着。
// 使用引用计数而非集合：多个请求可能携带相同 X-Request-ID
// （或回退到相同的 URL.Path），集合会让后到的请求“看不见”、
// 先结束的请求把别人的计数提前清掉，因此这里对同一 ID 累加。
func (f *Inflight) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id := r.Header.Get("X-Request-ID")
		if id == "" {
			id = r.URL.Path
		}
		f.register(id)
		defer f.release(id)
		next.ServeHTTP(w, r)
	})
}

// register 登记请求。
func (f *Inflight) register(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.current[id]++
}

// release 移除请求；引用计数归零时删除键。
func (f *Inflight) release(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	if n := f.current[id]; n <= 1 {
		delete(f.current, id)
	} else {
		f.current[id] = n - 1
	}
}

// Active 返回当前在处理中的请求数。
func (f *Inflight) Active() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	total := 0
	for _, n := range f.current {
		total += n
	}
	return total
}
