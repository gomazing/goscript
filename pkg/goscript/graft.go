package goscript

import (
	"html"
	"strings"
)

// GraftKind identifies the shape of a structured UI node.
type GraftKind string

const (
	GraftKindElement  GraftKind = "element"
	GraftKindText     GraftKind = "text"
	GraftKindRaw      GraftKind = "raw"
	GraftKindFragment GraftKind = "fragment"
)

// GraftNode is the human-facing structural UI graph used by Go Graft.
type GraftNode struct {
	Kind     GraftKind
	Tag      string
	Props    Props
	Value    string
	Children []GraftNode
}

// Graft builds an element node inside the structural UI graph.
func Graft(tag string, props Props, children ...GraftNode) GraftNode {
	return GraftNode{
		Kind:     GraftKindElement,
		Tag:      strings.TrimSpace(tag),
		Props:    cloneProps(props),
		Children: cloneGraftNodes(children),
	}
}

// GraftText builds a text node inside the structural UI graph.
func GraftText(value string) GraftNode {
	return GraftNode{
		Kind:  GraftKindText,
		Value: value,
	}
}

// GraftRaw builds a raw node that renders without escaping.
func GraftRaw(value string) GraftNode {
	return GraftNode{
		Kind:  GraftKindRaw,
		Value: value,
	}
}

// GraftFragment builds a fragment node that only lowers its children.
func GraftFragment(children ...GraftNode) GraftNode {
	return GraftNode{
		Kind:     GraftKindFragment,
		Children: cloneGraftNodes(children),
	}
}

// Lower converts a Go Graft node into the Go RUN representation.
func (n GraftNode) Lower() RunNode {
	switch n.Kind {
	case GraftKindText:
		return RunText(n.Value)
	case GraftKindRaw:
		return RunRaw(n.Value)
	case GraftKindFragment:
		children := make([]RunNode, 0, len(n.Children))
		for _, child := range n.Children {
			children = append(children, child.Lower())
		}
		return RunFragment(children...)
	default:
		children := make([]RunNode, 0, len(n.Children))
		for _, child := range n.Children {
			children = append(children, child.Lower())
		}
		return Run(n.Tag, n.Props, children...)
	}
}

// Render converts the graph into HTML by lowering through Go RUN.
func (n GraftNode) Render() string {
	return n.Lower().Render()
}

// GetProps satisfies the Component interface.
func (n GraftNode) GetProps() Props {
	return cloneProps(n.Props)
}

// GetChildren satisfies the Component interface.
func (n GraftNode) GetChildren() Children {
	out := make(Children, 0, len(n.Children))
	for _, child := range n.Children {
		out = append(out, child)
	}
	return out
}

// RunNode is the machine-facing lowered UI representation used by Go RUN.
type RunNode struct {
	Kind     GraftKind
	Tag      string
	Props    Props
	Value    string
	Children []RunNode
}

// Run builds an element node inside the lowered runtime graph.
func Run(tag string, props Props, children ...RunNode) RunNode {
	return RunNode{
		Kind:     GraftKindElement,
		Tag:      strings.TrimSpace(tag),
		Props:    cloneProps(props),
		Children: cloneRunNodes(children),
	}
}

// RunText builds a lowered text node.
func RunText(value string) RunNode {
	return RunNode{
		Kind:  GraftKindText,
		Value: value,
	}
}

// RunRaw builds a lowered raw node.
func RunRaw(value string) RunNode {
	return RunNode{
		Kind:  GraftKindRaw,
		Value: value,
	}
}

// RunFragment builds a lowered fragment node.
func RunFragment(children ...RunNode) RunNode {
	return RunNode{
		Kind:     GraftKindFragment,
		Children: cloneRunNodes(children),
	}
}

// Render converts the lowered runtime graph into HTML.
func (n RunNode) Render() string {
	return renderRunNode(n, false)
}

// GetProps satisfies the Component interface.
func (n RunNode) GetProps() Props {
	return cloneProps(n.Props)
}

// GetChildren satisfies the Component interface.
func (n RunNode) GetChildren() Children {
	out := make(Children, 0, len(n.Children))
	for _, child := range n.Children {
		out = append(out, child)
	}
	return out
}

func renderRunNode(node RunNode, raw bool) string {
	switch node.Kind {
	case GraftKindText:
		if raw {
			return node.Value
		}
		return html.EscapeString(node.Value)
	case GraftKindRaw:
		return node.Value
	case GraftKindFragment:
		return renderRunChildren(node.Children, raw)
	default:
		tag := strings.TrimSpace(node.Tag)
		if tag == "" {
			return ""
		}

		var result strings.Builder
		result.WriteString("<")
		result.WriteString(tag)
		result.WriteString(renderAttributes(node.Props))

		if len(node.Children) == 0 {
			if selfClosingTags[tag] {
				result.WriteString("/>")
			} else {
				result.WriteString("></")
				result.WriteString(tag)
				result.WriteString(">")
			}
			return result.String()
		}

		result.WriteString(">")
		result.WriteString(renderRunChildren(node.Children, rawTextTags[tag]))
		result.WriteString("</")
		result.WriteString(tag)
		result.WriteString(">")
		return result.String()
	}
}

func renderRunChildren(children []RunNode, raw bool) string {
	var result strings.Builder
	for _, child := range children {
		result.WriteString(renderRunNode(child, raw))
	}
	return result.String()
}

func cloneGraftNodes(children []GraftNode) []GraftNode {
	if len(children) == 0 {
		return nil
	}
	out := make([]GraftNode, len(children))
	copy(out, children)
	return out
}

func cloneRunNodes(children []RunNode) []RunNode {
	if len(children) == 0 {
		return nil
	}
	out := make([]RunNode, len(children))
	copy(out, children)
	return out
}
