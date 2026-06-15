package goscript

import "testing"

func TestPaintRectGeometry(t *testing.T) {
	rect := PaintRect{X: 10, Y: 20, W: 100, H: 80}

	if !rect.Contains(PaintPoint{X: 20, Y: 30}) {
		t.Fatalf("expected point to be inside rect")
	}
	if rect.Contains(PaintPoint{X: 1, Y: 1}) {
		t.Fatalf("expected point to be outside rect")
	}

	translated := rect.Translate(5, -5)
	if translated.X != 15 || translated.Y != 15 {
		t.Fatalf("unexpected translated rect: %#v", translated)
	}

	center := rect.Center()
	if center.X != 60 || center.Y != 60 {
		t.Fatalf("unexpected center point: %#v", center)
	}
}

func TestPaintHitTest(t *testing.T) {
	nodes := []PaintNode{
		{ID: "background", Bounds: PaintRect{X: 0, Y: 0, W: 200, H: 200}, Z: 0, Visible: true},
		{ID: "card", Bounds: PaintRect{X: 20, Y: 20, W: 60, H: 60}, Z: 2, Visible: true},
		{ID: "overlay", Bounds: PaintRect{X: 10, Y: 10, W: 80, H: 80}, Z: 5, Visible: true},
	}

	node, ok := PaintHitTest(nodes, PaintPoint{X: 30, Y: 30})
	if !ok {
		t.Fatalf("expected a hit")
	}
	if node.ID != "overlay" {
		t.Fatalf("expected overlay to win hit test, got %q", node.ID)
	}
}

func TestPaintIndexHitTest(t *testing.T) {
	nodes := []PaintNode{
		{ID: "spreadsheet", Bounds: PaintRect{X: 0, Y: 0, W: 1024, H: 768}, Z: 0, Visible: true},
		{ID: "overlay", Bounds: PaintRect{X: 700, Y: 500, W: 220, H: 120}, Z: 9, Visible: true},
		{ID: "side-panel", Bounds: PaintRect{X: 900, Y: 0, W: 300, H: 768}, Z: 4, Visible: true},
	}

	index := NewPaintIndex(128, nodes)
	node, ok := index.HitTest(PaintPoint{X: 760, Y: 560})
	if !ok {
		t.Fatalf("expected indexed hit")
	}
	if node.ID != "overlay" {
		t.Fatalf("expected overlay to win indexed hit test, got %q", node.ID)
	}

	node, ok = index.HitTest(PaintPoint{X: 960, Y: 40})
	if !ok {
		t.Fatalf("expected indexed hit in side panel")
	}
	if node.ID != "side-panel" {
		t.Fatalf("expected side-panel to win indexed hit test, got %q", node.ID)
	}
}

func BenchmarkPaintIndexHitTest(b *testing.B) {
	nodes := make([]PaintNode, 0, 1024)
	for row := 0; row < 32; row++ {
		for col := 0; col < 32; col++ {
			nodes = append(nodes, PaintNode{
				ID:      "cell",
				Bounds:  PaintRect{X: float64(col * 32), Y: float64(row * 24), W: 32, H: 24},
				Z:       row,
				Visible: true,
			})
		}
	}

	index := NewPaintIndex(64, nodes)
	point := PaintPoint{X: 512, Y: 288}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		if _, ok := index.HitTest(point); !ok {
			b.Fatalf("expected indexed hit")
		}
	}
}
