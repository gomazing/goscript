package goscript

import "testing"

func TestLanguageCoreSnapshotAndLowering(t *testing.T) {
	core := NewLanguageCore()

	if err := core.RegisterModule(ModuleSpec{
		Name: "pages",
		Path: "/modules/pages",
	}); err != nil {
		t.Fatalf("register pages module: %v", err)
	}

	if err := core.RegisterModule(ModuleSpec{
		Name: "calc",
		Path: "/modules/calc",
		Dependencies: map[string]string{
			"pages": "",
		},
	}); err != nil {
		t.Fatalf("register calc module: %v", err)
	}

	if err := core.AddSource(SourceUnit{
		Path:    "pages/home.gsx",
		Content: "TODO build the home surface",
	}); err != nil {
		t.Fatalf("add source: %v", err)
	}

	lowered, err := core.LowerSource("pages/home.gsx", func(unit SourceUnit) (GraftNode, []Diagnostic, error) {
		if unit.Module != "pages" {
			t.Fatalf("expected inferred module pages, got %q", unit.Module)
		}

		return Graft("div", Props{"id": "shell"}, GraftText("hello")), []Diagnostic{
			{
				Code:     "lowered",
				Severity: "info",
				Message:  "lowered by language core",
			},
		}, nil
	})
	if err != nil {
		t.Fatalf("lower source: %v", err)
	}

	if lowered.HTML != `<div id="shell">hello</div>` {
		t.Fatalf("unexpected lowered html: %s", lowered.HTML)
	}

	if lowered.Semantics.Plane != RuntimePlaneFrontend {
		t.Fatalf("expected frontend semantics, got %q", lowered.Semantics.Plane)
	}

	if len(lowered.Diagnostics) < 2 {
		t.Fatalf("expected source and lowerer diagnostics, got %#v", lowered.Diagnostics)
	}

	snapshot, err := core.Snapshot()
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	if len(snapshot.Modules) != 2 {
		t.Fatalf("expected 2 modules, got %d", len(snapshot.Modules))
	}

	if len(snapshot.Sources) != 1 {
		t.Fatalf("expected 1 source, got %d", len(snapshot.Sources))
	}

	if len(snapshot.BuildOrder) != 2 || snapshot.BuildOrder[0] != "pages" || snapshot.BuildOrder[1] != "calc" {
		t.Fatalf("unexpected build order: %#v", snapshot.BuildOrder)
	}

	if err := core.Validate(); err != nil {
		t.Fatalf("validate: %v", err)
	}
}

func TestLanguageCoreValidateMissingModule(t *testing.T) {
	core := NewLanguageCore()

	if err := core.AddSource(SourceUnit{
		Path:    "workspace/editor.gsx",
		Content: "<div>editor</div>",
	}); err != nil {
		t.Fatalf("add source: %v", err)
	}

	if err := core.Validate(); err == nil {
		t.Fatalf("expected validation error for missing module")
	}
}

func BenchmarkLanguageCoreLowerSource(b *testing.B) {
	core := NewLanguageCore()

	_ = core.RegisterModule(ModuleSpec{
		Name: "pages",
		Path: "/modules/pages",
	})
	_ = core.AddSource(SourceUnit{
		Path:    "pages/home.gsx",
		Content: "hello",
	})

	lowerer := func(unit SourceUnit) (GraftNode, []Diagnostic, error) {
		return Graft("div", Props{"id": "root"}, GraftText(unit.Content)), nil, nil
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := core.LowerSource("pages/home.gsx", lowerer); err != nil {
			b.Fatal(err)
		}
	}
}
