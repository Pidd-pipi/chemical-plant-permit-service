package main

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestWorkerCleanupOnCancel(t *testing.T) {
	svc := newOpsService(nil)
	w := newOpsWorker(svc, 10*time.Millisecond)
	cleaned := make(chan int, 1)
	w.onFinish = func(n int) { cleaned <- n }
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- w.Run(ctx) }()
	time.Sleep(30 * time.Millisecond)
	cancel()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("worker 未退出")
	}
	select {
	case <-cleaned:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("取消后未执行收尾清理")
	}
}

func TestWorkerZeroIntervalSafe(t *testing.T) {
	svc := newOpsService(nil)
	w := newOpsWorker(svc, 0)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan error, 1)
	go func() { done <- w.Run(ctx) }()
	cancel()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("worker 未退出")
	}
}

type failingWriter struct {
	buf   []byte
	limit int
}

func (w *failingWriter) Write(p []byte) (int, error) {
	if len(w.buf)+len(p) > w.limit {
		return 0, errors.New("disk full")
	}
	w.buf = append(w.buf, p...)
	return len(p), nil
}

func TestExportFlushErrorPreserved(t *testing.T) {
	rows := []ExportRow{
		{ID: "e-1", Status: "active", Priority: "high"},
		{ID: "e-2", Status: "closed", Priority: "low"},
	}
	fw := &failingWriter{limit: 20}
	err := WritePermitCSV(fw, rows)
	if err == nil {
		t.Fatal("flush 错误被吞掉，应返回错误")
	}
}

func TestBuildPermitRowsIndependent(t *testing.T) {
	svc := newOpsService([]OpsRecord{
		{ID: "x-1", Subject: "作业一", Owner: "a", Status: OpsStatusActive, Priority: OpsPriorityHigh, Labels: map[string]string{"site": "s"}},
		{ID: "x-2", Subject: "作业二", Owner: "b", Status: OpsStatusQueued, Priority: OpsPriorityNormal, Labels: map[string]string{"site": "s"}},
	})
	first, err := BuildPermitRows(svc, 10)
	if err != nil {
		t.Fatal(err)
	}
	second, err := BuildPermitRows(svc, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(second) != 2 {
		t.Fatalf("第二次导出行数应为 2，got %d（共享缓冲累积）", len(second))
	}
	if len(first) != 2 {
		t.Fatalf("第一次导出行数应为 2，got %d", len(first))
	}
}
