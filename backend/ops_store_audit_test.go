package main

import (
	"context"
	"sync"
	"testing"
)

func seedOpsRecords() []OpsRecord {
	return []OpsRecord{
		{ID: "p-001", Subject: "酸液转运", Owner: "zhang", Status: OpsStatusActive, Priority: OpsPriorityHigh, Labels: map[string]string{"site": "west", "gas": "ok"}},
		{ID: "p-002", Subject: "年度检修", Owner: "li", Status: OpsStatusQueued, Priority: OpsPriorityNormal, Labels: map[string]string{"site": "east"}},
	}
}

func TestOpsStoreGetIsolation(t *testing.T) {
	s := newOpsStore(seedOpsRecords())
	got, err := s.Get(context.Background(), "p-001")
	if err != nil {
		t.Fatal(err)
	}
	got.Labels["site"] = "mutated"
	again, err := s.Get(context.Background(), "p-001")
	if err != nil {
		t.Fatal(err)
	}
	if again.Labels["site"] != "west" {
		t.Fatalf("Get 返回的记录与内部存储共享 Labels: got %q want %q", again.Labels["site"], "west")
	}
}

func TestOpsStoreListIsolation(t *testing.T) {
	s := newOpsStore(seedOpsRecords())
	items, err := s.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	items[0].Labels["gas"] = "bad"
	again, err := s.List(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if again[0].Labels["gas"] != "ok" {
		t.Fatalf("List 返回的记录与内部存储共享 Labels: got %q want %q", again[0].Labels["gas"], "ok")
	}
}

func TestOpsAuditForIsolation(t *testing.T) {
	a := newOpsAudit()
	a.Add("p-001", "created", "zhang")
	events := a.For("p-001")
	if len(events) != 1 {
		t.Fatalf("want 1 event, got %d", len(events))
	}
	events[0].Details["hack"] = "yes"
	again := a.For("p-001")
	if _, ok := again[0].Details["hack"]; ok {
		t.Fatal("For 返回的事件与审计内部共享 Details")
	}
}

func TestOpsStoreRacingReadWrite(t *testing.T) {
	s := newOpsStore(seedOpsRecords())
	start := make(chan struct{})
	var wg sync.WaitGroup
	for i := 0; i < 3; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			_ = s.Count()
			_, _ = s.Get(context.Background(), "p-001")
			_, _ = s.List(context.Background())
		}()
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		<-start
		item := seedOpsRecords()[0]
		item.Status = OpsStatusClosed
		_ = s.Update(context.Background(), item, 1)
	}()
	close(start)
	wg.Wait()
}
