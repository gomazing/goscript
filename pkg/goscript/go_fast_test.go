package goscript

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestParseJSXLowering(t *testing.T) {
	parser := NewJSXParser(nil)

	got, err := parser.ParseJSX(`<div class="hero"><h1>Hello</h1><img alt="hero" /></div>`)
	if err != nil {
		t.Fatalf("ParseJSX returned error: %v", err)
	}

	want := `CreateElement("div", Props{"class": "hero"}, CreateElement("h1", nil, "Hello"), CreateElement("img", Props{"alt": "hero"}))`
	if got != want {
		t.Fatalf("ParseJSX mismatch\nwant: %s\ngot:  %s", want, got)
	}
}

func TestParseJSXFragmentAndExpressionProps(t *testing.T) {
	parser := NewJSXParser(nil)

	got, err := parser.ParseJSX(`<> <span title={title}>Hi</span> </>`)
	if err != nil {
		t.Fatalf("ParseJSX returned error: %v", err)
	}

	want := `Fragment(nil, CreateElement("span", Props{"title": title}, "Hi"))`
	if got != want {
		t.Fatalf("ParseJSX mismatch\nwant: %s\ngot:  %s", want, got)
	}
}

func TestTranspileGSXLowering(t *testing.T) {
	gsx := `package components

import (
	"github.com/gomazing/goscript/pkg/goscript"
)

func Home(props goscript.Props) string {
	return <div class="hero"><h1>Hello</h1></div>
}`

	got, err := TranspileGSX(gsx)
	if err != nil {
		t.Fatalf("TranspileGSX returned error: %v", err)
	}

	if !strings.Contains(got, `return CreateElement("div", Props{"class": "hero"}, CreateElement("h1", nil, "Hello"))`) {
		t.Fatalf("TranspileGSX did not lower JSX as expected:\n%s", got)
	}
}

func BenchmarkTranspileGSXCached(b *testing.B) {
	gsx := `package components

import (
	"github.com/gomazing/goscript/pkg/goscript"
)

func Home(props goscript.Props) string {
	return <div class="hero"><h1>Hello</h1></div>
}`

	if _, err := TranspileGSX(gsx); err != nil {
		b.Fatalf("warmup TranspileGSX returned error: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		got, err := TranspileGSX(gsx)
		if err != nil {
			b.Fatalf("TranspileGSX returned error: %v", err)
		}
		if len(got) == 0 {
			b.Fatalf("expected compiled output")
		}
	}
}

func TestParseSurfaceLowering(t *testing.T) {
	parser := NewJSXParser(nil)

	node, err := parser.ParseSurface(`<section class="hero"><h1>Hello</h1><img alt="hero" /></section>`)
	if err != nil {
		t.Fatalf("ParseSurface returned error: %v", err)
	}

	want := `<section class="hero"><h1>Hello</h1><img alt="hero"/></section>`
	if got := node.Render(); got != want {
		t.Fatalf("ParseSurface render mismatch\nwant: %s\ngot:  %s", want, got)
	}
}

func TestCompileSurfaceSource(t *testing.T) {
	lowered, err := CompileSurfaceSource("pages/home.gsx", `<section class="hero"><h1>Hello</h1></section>`)
	if err != nil {
		t.Fatalf("CompileSurfaceSource returned error: %v", err)
	}

	if lowered.Source.Kind != SourceKindSurface {
		t.Fatalf("expected surface source kind, got %q", lowered.Source.Kind)
	}

	if got, want := lowered.HTML, `<section class="hero"><h1>Hello</h1></section>`; got != want {
		t.Fatalf("unexpected lowered HTML\nwant: %s\ngot:  %s", want, got)
	}
}

func TestCompileSurfaceSourceAndCache(t *testing.T) {
	core := NewLanguageCore()
	if err := core.RegisterModule(ModuleSpec{Name: "pages", Path: "/modules/pages"}); err != nil {
		t.Fatalf("register module: %v", err)
	}
	if err := core.AddSource(SourceUnit{
		Path:    "pages/home.gsx",
		Module:  "pages",
		Content: `<section class="hero"><h1>Hello</h1></section>`,
	}); err != nil {
		t.Fatalf("add source: %v", err)
	}

	parser := NewJSXParser(nil)
	compiler := NewSurfaceCompiler(core, func(unit SourceUnit) (GraftNode, []Diagnostic, error) {
		surface, err := parser.ParseSurface(unit.Content)
		if err != nil {
			return GraftNode{}, nil, err
		}
		return surface, nil, nil
	})

	first, err := compiler.Compile("pages/home.gsx")
	if err != nil {
		t.Fatalf("first compile: %v", err)
	}

	if got, want := first.Surface.Render(), `<section class="hero"><h1>Hello</h1></section>`; got != want {
		t.Fatalf("unexpected first render\nwant: %s\ngot:  %s", want, got)
	}

	if err := core.AddSource(SourceUnit{
		Path:    "pages/home.gsx",
		Module:  "pages",
		Content: `<section class="hero"><h1>Updated</h1></section>`,
	}); err != nil {
		t.Fatalf("update source: %v", err)
	}

	second, err := compiler.Compile("pages/home.gsx")
	if err != nil {
		t.Fatalf("second compile: %v", err)
	}

	if got, want := second.Surface.Render(), `<section class="hero"><h1>Updated</h1></section>`; got != want {
		t.Fatalf("unexpected second render\nwant: %s\ngot:  %s", want, got)
	}

	if first.Surface.Render() == second.Surface.Render() {
		t.Fatalf("expected cache invalidation after source change")
	}
}

func TestRouterExactRouteFastPathAndParams(t *testing.T) {
	router := NewRouter()
	router.GET("/ping", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		_, _ = w.Write([]byte("pong"))
	})
	router.GET("/users/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		_, _ = w.Write([]byte(params["id"]))
	})

	exactReq := httptest.NewRequest(http.MethodGet, "/ping", nil)
	exactRec := httptest.NewRecorder()
	router.ServeHTTP(exactRec, exactReq)
	if exactRec.Body.String() != "pong" {
		t.Fatalf("exact route mismatch: got %q", exactRec.Body.String())
	}

	paramReq := httptest.NewRequest(http.MethodGet, "/users/42", nil)
	paramRec := httptest.NewRecorder()
	router.ServeHTTP(paramRec, paramReq)
	if paramRec.Body.String() != "42" {
		t.Fatalf("param route mismatch: got %q", paramRec.Body.String())
	}
}

func BenchmarkRouterServeHTTPFastPath(b *testing.B) {
	router := NewRouter()
	router.GET("/ping", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		_, _ = w.Write([]byte("pong"))
	})
	router.GET("/users/:id", func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		_, _ = w.Write([]byte(params["id"]))
	})

	b.Run("exact", func(b *testing.B) {
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		for i := 0; i < b.N; i++ {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
		}
	})

	b.Run("param", func(b *testing.B) {
		req := httptest.NewRequest(http.MethodGet, "/users/42", nil)
		for i := 0; i < b.N; i++ {
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)
		}
	})
}
