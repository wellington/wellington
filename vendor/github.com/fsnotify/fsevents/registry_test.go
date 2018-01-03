package fsevents

import "testing"

func TestRegistry(t *testing.T) {
	if registry.m == nil {
		t.Fatal("registry not initialized at start")
	}

	es := &EventStream{}
	i := registry.Add(es)

	if registry.Get(i) == nil {
		t.Fatal("failed to retrieve es from registry")
	}

	if es != registry.Get(i) {
		t.Errorf("eventstream did not match what was found in the registry")
	}

	registry.Delete(i)
	if registry.Get(i) != nil {
		t.Error("failed to delete registry")
	}
}
