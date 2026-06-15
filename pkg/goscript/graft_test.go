package goscript

import "testing"

func TestGraftLowerAndRender(t *testing.T) {
	node := Graft("section", Props{
		"class": "hero",
	},
		Graft("h1", nil, GraftText("Hello <World>")),
		GraftRaw("<hr>"),
	)

	run := node.Lower()
	if run.Kind != GraftKindElement {
		t.Fatalf("expected lowered element node, got %q", run.Kind)
	}

	got := node.Render()
	want := `<section class="hero"><h1>Hello &lt;World&gt;</h1><hr></section>`
	if got != want {
		t.Fatalf("unexpected render output\nwant: %s\ngot:  %s", want, got)
	}

	if rendered := CreateElement(node, nil); rendered != want {
		t.Fatalf("CreateElement should render graft nodes the same way\nwant: %s\ngot:  %s", want, rendered)
	}
}

func TestRunRender(t *testing.T) {
	node := Run("p", Props{"id": "note"}, RunText("Safe <bold>"))

	got := node.Render()
	want := `<p id="note">Safe &lt;bold&gt;</p>`
	if got != want {
		t.Fatalf("unexpected render output\nwant: %s\ngot:  %s", want, got)
	}
}

