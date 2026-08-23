package main

import (
	"context"
	"errors"
	"testing"
)

func errorChainSeed() []OpsRecord {
	return []OpsRecord{
		{ID: "ec-001", Subject: "动火作业", Owner: "wang", Status: OpsStatusActive, Priority: OpsPriorityHigh, Labels: map[string]string{"site": "north"}},
	}
}

func TestOpsGetNotFoundChain(t *testing.T) {
	svc := newOpsService(nil)
	_, err := svc.Get(context.Background(), "nope-000")
	if err == nil {
		t.Fatal("want error for missing record")
	}
	if !errors.Is(err, ErrOpsNotFound) {
		t.Fatalf("错误链断裂: errors.Is(err, ErrOpsNotFound)=false, err=%v", err)
	}
}

func TestOpsCreateConflictChain(t *testing.T) {
	seed := errorChainSeed()
	svc := newOpsService(seed)
	_, err := svc.Create(context.Background(), seed[0])
	if err == nil {
		t.Fatal("want conflict error for duplicate create")
	}
	if !errors.Is(err, ErrOpsConflict) {
		t.Fatalf("错误链断裂: errors.Is(err, ErrOpsConflict)=false, err=%v", err)
	}
}

func TestOpsStatusFromErrorMapping(t *testing.T) {
	cases := []struct {
		err  error
		want int
	}{
		{ErrOpsNotFound, 404},
		{ErrOpsConflict, 409},
		{ErrOpsInvalid, 400},
		{ErrOpsTransition, 409},
		{ErrOpsPolicy, 403},
	}
	for _, tc := range cases {
		if got := opsStatusFromError(tc.err); got != tc.want {
			t.Fatalf("opsStatusFromError(%v)=%d want %d", tc.err, got, tc.want)
		}
	}
}

func TestOpsStatusFromErrorWrapped(t *testing.T) {
	svc := newOpsService(nil)
	_, err := svc.Get(context.Background(), "missing-1")
	if err == nil {
		t.Fatal("want error")
	}
	if got := opsStatusFromError(err); got != 404 {
		t.Fatalf("不存在的许可应映射 404, got %d", got)
	}
}

func TestOpsTransitionErrorChain(t *testing.T) {
	seed := errorChainSeed()
	svc := newOpsService(seed)
	_, err := svc.Transition(context.Background(), "ec-001", 1, OpsStatusQueued, "tester")
	if err == nil {
		t.Fatal("want error for invalid transition")
	}
	if !errors.Is(err, ErrOpsTransition) {
		t.Fatalf("错误链断裂: errors.Is(err, ErrOpsTransition)=false, err=%v", err)
	}
}
