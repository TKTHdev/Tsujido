package tsujido

import (
	"sync"
	"testing"
)

func newTestEndpoint(t *testing.T) *Endpoint {
	t.Helper()
	ep := NewEndpoint()
	// Simulate consensus: forward ProposeCh to CommitCh.
	go func() {
		for req := range ep.ProposeCh {
			ep.CommitCh <- req
		}
	}()
	t.Cleanup(func() {
		close(ep.CommitCh)
		ep.Stop()
	})
	return ep
}

func TestSubmitAndQuery(t *testing.T) {
	ep := newTestEndpoint(t)

	res := ep.Handle("c1", Operation{Key: "x", Value: "hello"})
	if !res.Success || res.Value != "OK" {
		t.Fatalf("Handle got %+v", res)
	}

	res = ep.Query("x")
	if !res.Success || res.Value != "hello" {
		t.Fatalf("Query got %+v", res)
	}
}

func TestQueryMissing(t *testing.T) {
	ep := newTestEndpoint(t)

	res := ep.Query("missing")
	if !res.Success || res.Value != "" {
		t.Fatalf("Query missing got %+v", res)
	}
}

func TestOverwrite(t *testing.T) {
	ep := newTestEndpoint(t)

	ep.Handle("c1", Operation{Key: "x", Value: "v1"})
	ep.Handle("c1", Operation{Key: "x", Value: "v2"})

	res := ep.Query("x")
	if res.Value != "v2" {
		t.Fatalf("expected v2, got %s", res.Value)
	}
}

func TestSessionSeqIncrement(t *testing.T) {
	ep := newTestEndpoint(t)

	ep.Handle("c1", Operation{Key: "a", Value: "1"})
	ep.Handle("c1", Operation{Key: "b", Value: "2"})

	ep.sm.mu.RLock()
	s := ep.sm.sessions["c1"]
	ep.sm.mu.RUnlock()

	if s == nil || s.seq != 2 {
		t.Fatalf("expected seq 2, got %+v", s)
	}
}

func TestMultipleClients(t *testing.T) {
	ep := newTestEndpoint(t)

	ep.Handle("c1", Operation{Key: "x", Value: "from-c1"})
	ep.Handle("c2", Operation{Key: "x", Value: "from-c2"})

	res := ep.Query("x")
	if res.Value != "from-c2" {
		t.Fatalf("expected from-c2, got %s", res.Value)
	}

	ep.sm.mu.RLock()
	s1 := ep.sm.sessions["c1"]
	s2 := ep.sm.sessions["c2"]
	ep.sm.mu.RUnlock()

	if s1.seq != 1 || s2.seq != 1 {
		t.Fatalf("expected seq 1 each, got c1=%d c2=%d", s1.seq, s2.seq)
	}
}

func TestConcurrentSubmit(t *testing.T) {
	ep := newTestEndpoint(t)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ep.Handle("c1", Operation{Key: "x", Value: "v"})
		}()
	}
	wg.Wait()

	ep.sm.mu.RLock()
	seq := ep.sm.sessions["c1"].seq
	ep.sm.mu.RUnlock()

	if seq != 100 {
		t.Fatalf("expected seq 100, got %d", seq)
	}
}

func TestConcurrentQueryDuringSubmit(t *testing.T) {
	ep := newTestEndpoint(t)

	ep.Handle("c1", Operation{Key: "x", Value: "init"})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := ep.Query("x")
			if !res.Success {
				t.Errorf("Query failed")
			}
		}()
	}
	wg.Wait()
}
