package main

import (
	"context"
	"testing"
)

func TestOpsReviewFlowComplete(t *testing.T) {
	svc := newOpsService(nil)
	rec, err := svc.Create(context.Background(), OpsRecord{ID: "rv-001", Subject: "受限空间作业", Owner: "zhao", Status: OpsStatusQueued, Priority: OpsPriorityHigh, Labels: map[string]string{"site": "west"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Transition(context.Background(), rec.ID, rec.Revision, OpsStatusReview, "auditor"); err != nil {
		t.Fatalf("queued -> review 应被允许: %v", err)
	}
	rec, err = svc.Get(context.Background(), rec.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Transition(context.Background(), rec.ID, rec.Revision, OpsStatusActive, "operator"); err != nil {
		t.Fatalf("review -> active 应被允许: %v", err)
	}
	rec, err = svc.Get(context.Background(), rec.ID)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Transition(context.Background(), rec.ID, rec.Revision, OpsStatusClosed, "operator"); err != nil {
		t.Fatalf("active -> closed 应被允许: %v", err)
	}
}

func TestOpsRecordInProgress(t *testing.T) {
	rec := OpsRecord{ID: "rv-002", Status: OpsStatusReview}
	if !rec.InProgress() {
		t.Fatal("review 状态应算办理中")
	}
}

func TestOpsCountInProgress(t *testing.T) {
	items := []OpsRecord{
		{ID: "a", Status: OpsStatusQueued},
		{ID: "b", Status: OpsStatusReview},
		{ID: "c", Status: OpsStatusActive},
		{ID: "d", Status: OpsStatusClosed},
	}
	if got := opsCountInProgress(items); got != 3 {
		t.Fatalf("办理中统计应为 3, got %d", got)
	}
}

func TestOpsInProgressFunc(t *testing.T) {
	if !opsInProgress(OpsStatusReview) {
		t.Fatal("opsInProgress 应把 review 算进办理中")
	}
}
