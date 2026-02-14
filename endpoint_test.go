package tsujido

import "testing"

func TestEndpointHandle(t *testing.T) {
	ep := newTestEndpoint(t)

	res := ep.Handle("c1", Operation{Key: "x", Value: "1"})
	if !res.Success {
		t.Fatalf("expected success, got %+v", res)
	}

	got := ep.Query("x")
	if got.Value != "1" {
		t.Fatalf("expected value '1', got %q", got.Value)
	}
}
