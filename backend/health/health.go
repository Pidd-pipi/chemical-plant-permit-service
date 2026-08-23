package health

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

var totalChecks uint64

// Handler 返回服务健康状态并累计检查次数。
func Handler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", 405)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"status": "ok", "service": "chemical-plant-permit", "checks": atomic.AddUint64(&totalChecks, 1)})
}
