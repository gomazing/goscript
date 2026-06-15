package goscript

import "math"

// PaintPoint describes a 2D coordinate in Go PAINT space.
type PaintPoint struct {
	X float64
	Y float64
}

// PaintRect describes a 2D spatial region in Go PAINT space.
type PaintRect struct {
	X float64
	Y float64
	W float64
	H float64
}

// Contains reports whether the point lies inside the rectangle.
func (r PaintRect) Contains(p PaintPoint) bool {
	return p.X >= r.X && p.Y >= r.Y && p.X <= r.X+r.W && p.Y <= r.Y+r.H
}

// Intersects reports whether two rectangles overlap.
func (r PaintRect) Intersects(other PaintRect) bool {
	return r.X < other.X+other.W &&
		r.X+r.W > other.X &&
		r.Y < other.Y+other.H &&
		r.Y+r.H > other.Y
}

// Translate moves the rectangle by the supplied delta.
func (r PaintRect) Translate(dx, dy float64) PaintRect {
	r.X += dx
	r.Y += dy
	return r
}

// Inset shrinks or expands the rectangle symmetrically.
func (r PaintRect) Inset(dx, dy float64) PaintRect {
	r.X += dx
	r.Y += dy
	r.W -= dx * 2
	r.H -= dy * 2
	return r
}

// Center returns the geometric center point.
func (r PaintRect) Center() PaintPoint {
	return PaintPoint{
		X: r.X + (r.W / 2),
		Y: r.Y + (r.H / 2),
	}
}

// Empty reports whether the rectangle has no visible area.
func (r PaintRect) Empty() bool {
	return r.W <= 0 || r.H <= 0
}

// PaintNode describes a spatial node that can be drawn and hit-tested.
type PaintNode struct {
	ID       string
	Bounds   PaintRect
	Z        int
	Visible  bool
	Payload  interface{}
}

// Hit reports whether the point intersects the node.
func (n PaintNode) Hit(p PaintPoint) bool {
	if !n.Visible {
		return false
	}
	return n.Bounds.Contains(p)
}

// PaintHitTest returns the top-most visible node containing the point.
func PaintHitTest(nodes []PaintNode, point PaintPoint) (PaintNode, bool) {
	var (
		best    PaintNode
		found   bool
		bestIdx int
	)

	for idx, node := range nodes {
		if !node.Hit(point) {
			continue
		}

		if !found || node.Z > best.Z || (node.Z == best.Z && idx > bestIdx) {
			best = node
			bestIdx = idx
			found = true
		}
	}

	return best, found
}

type paintCellKey struct {
	X int
	Y int
}

// PaintIndex accelerates hit testing by grouping nodes into spatial cells.
type PaintIndex struct {
	CellSize float64

	nodes []PaintNode
	cells map[paintCellKey][]int
}

// NewPaintIndex creates a spatial hit-test index for a set of nodes.
func NewPaintIndex(cellSize float64, nodes []PaintNode) *PaintIndex {
	index := &PaintIndex{
		CellSize: cellSize,
		cells:    make(map[paintCellKey][]int),
	}
	index.Rebuild(nodes)
	return index
}

// Rebuild refreshes the spatial index from a new node set.
func (i *PaintIndex) Rebuild(nodes []PaintNode) {
	if i == nil {
		return
	}

	if i.CellSize <= 0 {
		i.CellSize = 128
	}

	if i.cells == nil {
		i.cells = make(map[paintCellKey][]int)
	} else {
		clear(i.cells)
	}

	i.nodes = append(i.nodes[:0], nodes...)
	for idx, node := range nodes {
		if !node.Visible || node.Bounds.Empty() {
			continue
		}

		minX := paintCellCoord(node.Bounds.X, i.CellSize)
		maxX := paintCellCoord(node.Bounds.X+node.Bounds.W, i.CellSize)
		minY := paintCellCoord(node.Bounds.Y, i.CellSize)
		maxY := paintCellCoord(node.Bounds.Y+node.Bounds.H, i.CellSize)

		for x := minX; x <= maxX; x++ {
			for y := minY; y <= maxY; y++ {
				key := paintCellKey{X: x, Y: y}
				i.cells[key] = append(i.cells[key], idx)
			}
		}
	}
}

// HitTest returns the top-most visible node inside the index that contains the point.
func (i *PaintIndex) HitTest(point PaintPoint) (PaintNode, bool) {
	if i == nil {
		return PaintNode{}, false
	}
	if i.CellSize <= 0 {
		i.CellSize = 128
	}

	key := paintCellKey{
		X: paintCellCoord(point.X, i.CellSize),
		Y: paintCellCoord(point.Y, i.CellSize),
	}

	indices := i.cells[key]
	if len(indices) == 0 {
		return PaintHitTest(i.nodes, point)
	}

	var (
		best    PaintNode
		bestIdx int
		found   bool
		seen    = make(map[int]struct{}, len(indices))
	)

	for _, idx := range indices {
		if _, ok := seen[idx]; ok {
			continue
		}
		seen[idx] = struct{}{}
		if idx < 0 || idx >= len(i.nodes) {
			continue
		}

		node := i.nodes[idx]
		if !node.Hit(point) {
			continue
		}

		if !found || node.Z > best.Z || (node.Z == best.Z && idx > bestIdx) {
			best = node
			bestIdx = idx
			found = true
		}
	}

	if found {
		return best, true
	}

	return PaintHitTest(i.nodes, point)
}

// PaintClamp clamps a point to the bounds of a rectangle.
func PaintClamp(point PaintPoint, bounds PaintRect) PaintPoint {
	if point.X < bounds.X {
		point.X = bounds.X
	}
	if point.Y < bounds.Y {
		point.Y = bounds.Y
	}
	if point.X > bounds.X+bounds.W {
		point.X = bounds.X + bounds.W
	}
	if point.Y > bounds.Y+bounds.H {
		point.Y = bounds.Y + bounds.H
	}
	return point
}

func paintCellCoord(value, cellSize float64) int {
	return int(math.Floor(value / cellSize))
}
