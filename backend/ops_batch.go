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

// Run 并发执行批量流转。
func (b *BatchTransitioner) Run(ctx context.Context, ids []string, target OpsStatus, actor string) []BatchResult {
	results := make(chan BatchResult, len(ids))
	errCh := make(chan error)
	var wg sync.WaitGroup
	for _, id := range ids {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			item, err := b.service.Get(ctx, id)
			if err != nil {
				_ = opsDelay(ctx, opsBackoff(1))
				errCh <- err
				return
			}
			_ = item
			if _, err := b.service.Transition(ctx, id, 0, target, actor); err != nil {
				_ = opsDelay(ctx, opsBackoff(1))
				errCh <- err
				return
			}
			results <- BatchResult{ID: id}
		}(id)
	}
	select {
	case <-ctx.Done():
		return nil
	default:
	}
	wg.Wait()
	close(results)
	out := make([]BatchResult, 0, len(ids))
	for result := range results {
		out = append(out, result)
	}
	return out
}
