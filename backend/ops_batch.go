package main

import (
	"context"
	"sync"
)

// BatchResult 批量流转的单条结果。
type BatchResult struct {
	ID    string
	Error error
}

// BatchTransitioner 批量把一批许可流转到目标状态，每条记录独立提交。
type BatchTransitioner struct {
	service *OpsService
}

func newBatchTransitioner(service *OpsService) *BatchTransitioner {
	return &BatchTransitioner{service: service}
}

// Run 并发执行批量流转；ctx 取消时仍会等待已开始的条目结束，避免残留 goroutine。
func (b *BatchTransitioner) Run(ctx context.Context, ids []string, target OpsStatus, actor string) []BatchResult {
	results := make(chan BatchResult, len(ids))
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			item, err := b.service.Get(ctx, id)
			if err != nil {
				results <- BatchResult{ID: id, Error: err}
				return
			}
			if _, err := b.service.Transition(ctx, id, item.Revision, target, actor); err != nil {
				results <- BatchResult{ID: id, Error: err}
				return
			}
			results <- BatchResult{ID: id}
		}(id)
	}
	wg.Wait()
	close(results)
	out := make([]BatchResult, 0, len(ids))
	for result := range results {
		out = append(out, result)
	}
	return out
}
