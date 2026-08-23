package main

import (
	"context"
	"time"
)

// OpsWorker 后台定时任务：周期性清理已关闭许可并做收尾工作。
type OpsWorker struct {
	service  *OpsService
	interval time.Duration
	onFinish func(cleaned int)
}

func newOpsWorker(service *OpsService, interval time.Duration) *OpsWorker {
	return &OpsWorker{service: service, interval: interval, onFinish: func(int) {}}
}

// Run 启动后台循环直到 ctx 取消。
func (w *OpsWorker) Run(ctx context.Context) error {
	ticker := time.NewTicker(w.interval)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			w.service.cleanupExpired(time.Now())
		}
	}
}
