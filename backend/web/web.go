package web

import (
	"embed"
	"net/http"
	"sync"
	"time"
)

//go:embed index.html app.js
var files embed.FS

var (
	refreshMu     sync.Mutex
	lastRefreshed time.Time
)

// Handler 返回内置页面并记录最近一次刷新时间。
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" && r.URL.Path != "/app.js" {
		http.NotFound(w, r)
		return
	}
	refreshMu.Lock()
	lastRefreshed = time.Now()
	refreshMu.Unlock()
	w.Header().Set("X-Refreshed-At", lastRefreshed.UTC().Format(time.RFC3339))
	http.FileServer(http.FS(files)).ServeHTTP(w, r)
}
