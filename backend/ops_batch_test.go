package main

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"
)

func batchSeed() []OpsRecord {
	records := make([]OpsRecord, 0, 60)
	for i := 0; i < 60; i++ {
		id := fmt.Sprintf("b-%03d", i)
		records = append(records, OpsRecord{ID: id, Subject: "单元检修", Owner: "op", Status: OpsStatusQueued, Priority: OpsPriorityNormal, Labels: map[string]string{"site": "x"}})
	}
	return records
}

func batchIDs(n int) []string {
	ids := make([]string, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, fmt.Sprintf("b-%03d", i))
	}
	return ids
}

func TestBatchRunAllCompleted(t *testing.T) {
	svc := newOpsService(batchSeed())
	b := newBatchTransitioner(svc)
	ids := batchIDs(60)
	done := make(chan []BatchResult, 1)
	go func() { done <- b.Run(context.Background(), ids, OpsStatusClosed, "tester") }()
	select {
	case res := <-done:
		if len(res) != len(ids) {
			t.Fatalf("批量结果不全: got %d/%d", len(res), len(ids))
		}
	case <-time.After(3 * time.Second):
		t.Fatal("批量任务挂死")
	}
}

func TestBatchRunFailureReturnsPromptly(t *testing.T) {
	svc := newOpsService(batchSeed())
	b := newBatchTransitioner(svc)
	ids := []string{"b-000", "b-001", "missing-1"}
	done := make(chan []BatchResult, 1)
	go func() {
		done <- b.Run(context.Background(), ids, OpsStatusClosed, "tester")
	}()
	select {
	case res := <-done:
		if len(res) != len(ids) {
			t.Fatalf("want %d results, got %d", len(ids), len(res))
		}
		found := false
		for _, r := range res {
			if r.ID == "missing-1" && r.Error != nil {
				found = true
			}
		}
		if !found {
			t.Fatal("不存在的记录应带错误结果返回")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("批量任务在失败分支挂死")
	}
}

func TestBatchCancelNoGoroutineLeak(t *testing.T) {
	svc := newOpsService(batchSeed())
	b := newBatchTransitioner(svc)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	ids := []string{"b-000", "b-001", "b-002"}
	before := runtime.NumGoroutine()
	done := make(chan []BatchResult, 1)
	go func() { done <- b.Run(ctx, ids, OpsStatusClosed, "tester") }()
	select {
	case res := <-done:
		if len(res) != len(ids) {
			t.Fatalf("want %d results, got %d", len(ids), len(res))
		}
	case <-time.After(2 * time.Second):
		t.Fatal("取消后批量任务挂死")
	}
	time.Sleep(100 * time.Millisecond)
	after := runtime.NumGoroutine()
	if after > before {
		t.Fatalf("goroutine 泄露: before=%d after=%d", before, after)
	}
}

func TestBatchRunRevisionChecked(t *testing.T) {
	svc := newOpsService([]OpsRecord{{ID: "r-1", Subject: "动火作业", Owner: "wu", Status: OpsStatusQueued, Priority: OpsPriorityHigh, Labels: map[string]string{"site": "x"}}})
	b := newBatchTransitioner(svc)
	res := b.Run(context.Background(), []string{"r-1", "r-1"}, OpsStatusActive, "tester")
	errors := 0
	for _, r := range res {
		if r.Error != nil {
			errors++
		}
	}
	if errors == 0 {
		t.Fatal("并发流转同一记录应产生版本冲突")
	}
}

func TestOpsDelayHonorsCancel(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	done := make(chan error, 1)
	go func() { done <- opsDelay(ctx, 10*time.Second) }()
	select {
	case err := <-done:
		if err == nil {
			t.Fatal("opsDelay 应返回 context 取消错误")
		}
	case <-time.After(500 * time.Millisecond):
		t.Fatal("opsDelay 忽略取消，硬等 10 秒")
	}
}

func TestBatchRunConcurrentRuns(t *testing.T) {
	svc := newOpsService(batchSeed())
	b := newBatchTransitioner(svc)
	ids := []string{"b-000", "b-001", "missing-1", "b-002"}
	start := make(chan struct{})
	results := make(chan []BatchResult, 2)
	for i := 0; i < 2; i++ {
		go func() {
			<-start
			results <- b.Run(context.Background(), ids, OpsStatusClosed, "tester")
		}()
	}
	close(start)
	for i := 0; i < 2; i++ {
		select {
		case res := <-results:
			if len(res) != len(ids) {
				t.Fatalf("结果条数不对: %d/%d", len(res), len(ids))
			}
		case <-time.After(2 * time.Second):
			t.Fatal("并发批量任务在失败分支挂死")
		}
	}
}
