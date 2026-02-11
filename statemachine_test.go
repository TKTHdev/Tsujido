package tsujido

import (
	"sync"
	"testing"
)

func TestSubmitAndQuery(t *testing.T) {
	sm := NewStateMachine()
	defer sm.Stop()

	res := sm.Submit("c1", Operation{Key: "x", Value: "hello"})
	if !res.Success || res.Value != "OK" {
		t.Fatalf("Submit got %+v", res)
	}

	res = sm.Query("x")
	if !res.Success || res.Value != "hello" {
		t.Fatalf("Query got %+v", res)
	}
}

func TestQueryMissing(t *testing.T) {
	sm := NewStateMachine()
	defer sm.Stop()

	res := sm.Query("missing")
	if !res.Success || res.Value != "" {
		t.Fatalf("Query missing got %+v", res)
	}
}

func TestOverwrite(t *testing.T) {
	sm := NewStateMachine()
	defer sm.Stop()

	sm.Submit("c1", Operation{Key: "x", Value: "v1"})
	sm.Submit("c1", Operation{Key: "x", Value: "v2"})

	res := sm.Query("x")
	if res.Value != "v2" {
		t.Fatalf("expected v2, got %s", res.Value)
	}
}

func TestSessionSeqIncrement(t *testing.T) {
	sm := NewStateMachine()
	defer sm.Stop()

	sm.Submit("c1", Operation{Key: "a", Value: "1"})
	sm.Submit("c1", Operation{Key: "b", Value: "2"})

	sm.mu.RLock()
	s := sm.sessions["c1"]
	sm.mu.RUnlock()

	if s == nil || s.seq != 2 {
		t.Fatalf("expected seq 2, got %+v", s)
	}
}

func TestMultipleClients(t *testing.T) {
	sm := NewStateMachine()
	defer sm.Stop()

	sm.Submit("c1", Operation{Key: "x", Value: "from-c1"})
	sm.Submit("c2", Operation{Key: "x", Value: "from-c2"})

	res := sm.Query("x")
	if res.Value != "from-c2" {
		t.Fatalf("expected from-c2, got %s", res.Value)
	}

	sm.mu.RLock()
	s1 := sm.sessions["c1"]
	s2 := sm.sessions["c2"]
	sm.mu.RUnlock()

	if s1.seq != 1 || s2.seq != 1 {
		t.Fatalf("expected seq 1 each, got c1=%d c2=%d", s1.seq, s2.seq)
	}
}

func TestConcurrentSubmit(t *testing.T) {
	sm := NewStateMachine()
	defer sm.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sm.Submit("c1", Operation{Key: "x", Value: "v"})
		}()
	}
	wg.Wait()

	sm.mu.RLock()
	seq := sm.sessions["c1"].seq
	sm.mu.RUnlock()

	if seq != 100 {
		t.Fatalf("expected seq 100, got %d", seq)
	}
}

func TestConcurrentQueryDuringSubmit(t *testing.T) {
	sm := NewStateMachine()
	defer sm.Stop()

	sm.Submit("c1", Operation{Key: "x", Value: "init"})

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := sm.Query("x")
			if !res.Success {
				t.Errorf("Query failed")
			}
		}()
	}
	wg.Wait()
}
