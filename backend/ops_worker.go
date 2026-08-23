package main

import (
	"context"
	"fmt"
	"time"
)

// OpsWorker 后台定时任务：周期性清理已关闭许可并做收尾工作。
type OpsWorker struct {
	service  *OpsService
	interval time.Duration
	onFinish func(cleaned int)
}

// minOpsInterval 是后台巡检允许的最小周期，过小会拖垮存储。
const minOpsInterval = 50 * time.Millisecond

func newOpsWorker(service *OpsService, interval time.Duration) *OpsWorker {
	if interval < minOpsInterval {
		interval = minOpsInterval
	}
	return &OpsWorker{service: service, interval: interval, onFinish: func(int) {}}
}

// Run 启动后台循环直到 ctx 取消。退出时停止 ticker 并调用 onFinish 收尾。
func (w *OpsWorker) Run(ctx context.Context) error {
	if w.interval <= 0 {
		return fmt.Errorf("ops worker: non-positive interval %v", w.interval)
	}
	ticker := time.NewTicker(w.interval)
	defer func() {
		// 退出前必须停掉 ticker，避免泄漏后台 goroutine / 定时器，并执行收尾回调。
		ticker.Stop()
		w.onFinish(0)
	}()
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if cleaned := w.service.cleanupExpired(time.Now()); cleaned > 0 {
				w.onFinish(cleaned)
			}
		}
	}
}
