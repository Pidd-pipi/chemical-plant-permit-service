package main

import (
	"context"
	"time"
)

// OpsWorker 后台定时任务：周期性清理已关闭许可并做收尾工作。
type OpsWorker struct {
	service  *OpsService
	interval time.Duration
	ticker   *time.Ticker
	onFinish func(cleaned int)
}

func newOpsWorker(service *OpsService, interval time.Duration) *OpsWorker {
	return &OpsWorker{service: service, interval: interval, onFinish: func(int) {}}
}

// Run 启动后台循环直到 ctx 取消；退出前执行一次收尾清理并停止定时器。
func (w *OpsWorker) Run(ctx context.Context) error {
	w.ticker = time.NewTicker(w.interval)
	defer func() {
		if w.ticker != nil {
			w.ticker.Stop()
			w.ticker = nil
		}
	}()
	for {
		select {
		case <-ctx.Done():
			cleaned := w.service.cleanupExpired(time.Now())
			if w.onFinish != nil {
				w.onFinish(cleaned)
			}
			return nil
		case <-w.ticker.C:
			w.service.cleanupExpired(time.Now())
		}
	}
}
