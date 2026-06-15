package goscript

import "testing"

func TestModuleGraphBuildOrder(t *testing.T) {
	graph := NewModuleGraph(nil)

	if err := graph.Register(ModuleSpec{Name: "storage", Path: "/modules/storage"}); err != nil {
		t.Fatalf("register storage: %v", err)
	}
	if err := graph.Register(ModuleSpec{Name: "ui", Path: "/modules/ui", Dependencies: map[string]string{"storage": ""}}); err != nil {
		t.Fatalf("register ui: %v", err)
	}

	order, err := graph.BuildOrder()
	if err != nil {
		t.Fatalf("build order: %v", err)
	}

	if len(order) != 2 || order[0] != "storage" || order[1] != "ui" {
		t.Fatalf("unexpected order: %#v", order)
	}

	if err := graph.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}
