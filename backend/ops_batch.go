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

// batchConcurrency 限制单次批量任务最多同时运行的 goroutine 数量，
// 避免大批量提交时 goroutine 只增不减、瞬时挤爆调度器。
const batchConcurrency = 8

// BatchTransitioner 批量把一批许可流转到目标状态，每条记录独立提交。
type BatchTransitioner struct {
	service *OpsService
}

func newBatchTransitioner(service *OpsService) *BatchTransitioner {
	return &BatchTransitioner{service: service}
}

// Run 并发执行批量流转。
//
// 设计要点：
//   - 每条记录的成败互不影响：失败的记录只记录错误，不会阻塞其它记录。
//   - 结果直接写入按 id 下标预分配的切片，由 WaitGroup 保护并发写入互不干扰，
//     发送端永远不必等待接收端，因此 ctx 取消时 goroutine 能及时退出，不必硬等。
//   - 用信号量限制并发 goroutine 数量，避免 goroutine 只增不减。
//   - 流转时传入当前记录的 Revision 做乐观锁校验，同一张单子并发流转时
//     只有先提交的那条能成功，后到的那条会拿到 ErrOpsConflict。
func (b *BatchTransitioner) Run(ctx context.Context, ids []string, target OpsStatus, actor string) []BatchResult {
	if len(ids) == 0 {
		return nil
	}
	results := make([]BatchResult, len(ids))
	sem := make(chan struct{}, batchConcurrency)

	var wg sync.WaitGroup
	for i, id := range ids {
		// ctx 已取消就不再派发新任务；已派发的任务也会在内部感知到取消并尽快退出。
		select {
		case <-ctx.Done():
			results[i] = BatchResult{ID: id, Error: ctx.Err()}
			continue
		case sem <- struct{}{}:
		}
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			defer func() { <-sem }()
			results[i] = b.transitionOne(ctx, id, target, actor)
		}(i, id)
	}

	// 等待所有已派发的任务结束。任务内部会感知 ctx 取消，
	// 不会因无缓冲 channel 而死等。
	wg.Wait()
	return results
}

// transitionOne 流转单条记录：读取当前 Revision 并用其做乐观锁，
// 失败时只返回错误，不影响调用方对其它记录的处理。
func (b *BatchTransitioner) transitionOne(ctx context.Context, id string, target OpsStatus, actor string) BatchResult {
	item, err := b.service.Get(ctx, id)
	if err != nil {
		return BatchResult{ID: id, Error: err}
	}
	if _, err := b.service.Transition(ctx, id, item.Revision, target, actor); err != nil {
		return BatchResult{ID: id, Error: err}
	}
	return BatchResult{ID: id}
}
