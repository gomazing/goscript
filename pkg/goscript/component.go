package goscript

import (
	"fmt"
	"html"
	"reflect"
	"sort"
	"strings"
	"sync"
)

// Props represents component properties.
type Props map[string]interface{}

// RawHTML marks content that should be rendered without escaping.
type RawHTML string

// Raw creates an explicit HTML fragment.
func Raw(value string) RawHTML {
	return RawHTML(value)
}

// PropValidator validates a specific prop.
type PropValidator func(value interface{}) error

// ReflectKind constants for prop types.
const (
	ReflectKindInvalid reflect.Kind = reflect.Invalid
	ReflectKindBool    reflect.Kind = reflect.Bool
	ReflectKindInt     reflect.Kind = reflect.Int
	ReflectKindInt8    reflect.Kind = reflect.Int8
	ReflectKindInt16   reflect.Kind = reflect.Int16
	ReflectKindInt32   reflect.Kind = reflect.Int32
	ReflectKindInt64   reflect.Kind = reflect.Int64
	ReflectKindUint    reflect.Kind = reflect.Uint
	ReflectKindUint8   reflect.Kind = reflect.Uint8
	ReflectKindUint16  reflect.Kind = reflect.Uint16
	ReflectKindUint32  reflect.Kind = reflect.Uint32
	ReflectKindUint64  reflect.Kind = reflect.Uint64
	ReflectKindFloat32 reflect.Kind = reflect.Float32
	ReflectKindFloat64 reflect.Kind = reflect.Float64
	ReflectKindString  reflect.Kind = reflect.String
	ReflectKindSlice   reflect.Kind = reflect.Slice
	ReflectKindMap     reflect.Kind = reflect.Map
)

// PropType defines the expected type and validation for a prop.
type PropType struct {
	Type      reflect.Kind
	Required  bool
	Validator PropValidator
	Default   interface{}
}

// PropTypes defines the expected props for a component.
type PropTypes map[string]PropType

// Children represents component children.
type Children []interface{}

// Component interface defines methods all components must implement.
type Component interface {
	Render() string
	GetProps() Props
	GetChildren() Children
}

// LifecycleComponent extends Component with lifecycle methods.
type LifecycleComponent interface {
	Component
	ComponentDidMount()
	ComponentWillUnmount()
	ShouldComponentUpdate(nextProps Props) bool
}

// BaseComponent provides a basic implementation of Component.
type BaseComponent struct {
	props     Props
	children   Children
	state     map[string]interface{}
	propTypes PropTypes
	mutex     sync.RWMutex
}

// NewBaseComponent creates a new BaseComponent.
func NewBaseComponent(props Props, propTypes PropTypes, children ...interface{}) *BaseComponent {
	if props == nil {
		props = Props{}
	}

	for key, propType := range propTypes {
		if _, exists := props[key]; !exists && propType.Default != nil {
			props[key] = propType.Default
		}
	}

	return &BaseComponent{
		props:     props,
		children:   children,
		state:     make(map[string]interface{}),
		propTypes: propTypes,
	}
}

// ValidateProps validates component props against propTypes.
func (b *BaseComponent) ValidateProps() []error {
	var errors []error

	for key, propType := range b.propTypes {
		value, exists := b.props[key]

		if propType.Required && !exists {
			errors = append(errors, fmt.Errorf("required prop '%s' is missing", key))
			continue
		}

		if exists {
			if value != nil {
				t := reflect.TypeOf(value)
				actualKind := reflect.Invalid
				if t != nil {
					actualKind = t.Kind()
				}
				if t == nil || actualKind != propType.Type {
					errors = append(errors, fmt.Errorf("prop '%s' should be of type %v, got %v", key, propType.Type, actualKind))
				}
			}

			if propType.Validator != nil {
				if err := propType.Validator(value); err != nil {
					errors = append(errors, fmt.Errorf("prop '%s' validation failed: %v", key, err))
				}
			}
		}
	}

	return errors
}

// GetProps returns component props.
func (b *BaseComponent) GetProps() Props {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	return cloneProps(b.props)
}

// SetProps replaces the component props.
func (b *BaseComponent) SetProps(props Props) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.props = cloneProps(props)
}

// GetChildren returns component children.
func (b *BaseComponent) GetChildren() Children {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	out := make(Children, len(b.children))
	copy(out, b.children)
	return out
}

// SetChildren replaces the component children.
func (b *BaseComponent) SetChildren(children ...interface{}) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.children = append(Children(nil), children...)
}

// SetState updates component state.
func (b *BaseComponent) SetState(key string, value interface{}) {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	b.state[key] = value
}

// GetState retrieves component state.
func (b *BaseComponent) GetState(key string) interface{} {
	b.mutex.RLock()
	defer b.mutex.RUnlock()
	return b.state[key]
}

// SnapshotState returns a shallow copy of component state.
func (b *BaseComponent) SnapshotState() map[string]interface{} {
	b.mutex.RLock()
	defer b.mutex.RUnlock()

	out := make(map[string]interface{}, len(b.state))
	for key, value := range b.state {
		out[key] = value
	}
	return out
}

// Render implements the Component interface.
func (b *BaseComponent) Render() string {
	return ""
}

// FunctionalComponent represents a function that renders a component.
type FunctionalComponent func(props Props, children ...interface{}) string

// Render implements the Component interface for FunctionalComponent.
func (f FunctionalComponent) Render() string {
	if f == nil {
		return ""
	}
	return f(nil)
}

// GetProps implements the Component interface for FunctionalComponent.
func (f FunctionalComponent) GetProps() Props {
	return nil
}

// GetChildren implements the Component interface for FunctionalComponent.
func (f FunctionalComponent) GetChildren() Children {
	return nil
}

var selfClosingTags = map[string]bool{
	"area": true,
	"base": true,
	"br":   true,
	"col":  true,
	"embed": true,
	"hr":   true,
	"img":  true,
	"input": true,
	"link": true,
	"meta": true,
	"param": true,
	"source": true,
	"track": true,
	"wbr":   true,
}

var rawTextTags = map[string]bool{
	"script":   true,
	"style":    true,
	"textarea": true,
}

func cloneProps(props Props) Props {
	if props == nil {
		return nil
	}

	out := make(Props, len(props))
	for key, value := range props {
		out[key] = value
	}
	return out
}

func renderChild(child interface{}, raw bool) string {
	switch ch := child.(type) {
	case nil:
		return ""
	case RawHTML:
		return string(ch)
	case FunctionalComponent:
		if ch == nil {
			return ""
		}
		return ch(nil)
	case Component:
		return ch.Render()
	case string:
		if raw {
			return ch
		}
		return html.EscapeString(ch)
	case fmt.Stringer:
		if raw {
			return ch.String()
		}
		return html.EscapeString(ch.String())
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64, bool:
		return html.EscapeString(fmt.Sprintf("%v", ch))
	default:
		return html.EscapeString(fmt.Sprintf("%v", ch))
	}
}

func renderAttributeValue(value interface{}) string {
	switch v := value.(type) {
	case nil:
		return ""
	case bool:
		if v {
			return ""
		}
		return ""
	case RawHTML:
		return string(v)
	case fmt.Stringer:
		return html.EscapeString(v.String())
	default:
		return html.EscapeString(fmt.Sprintf("%v", v))
	}
}

func renderAttributes(props Props) string {
	if len(props) == 0 {
		return ""
	}

	keys := make([]string, 0, len(props))
	for key := range props {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	var result strings.Builder
	for _, key := range keys {
		value := props[key]

		if val, ok := value.(bool); ok {
			if val {
				result.WriteString(fmt.Sprintf(" %s", key))
			}
			continue
		}

		if value == nil {
			continue
		}

		if styleMap, ok := value.(map[string]interface{}); ok && key == "style" {
			styleKeys := make([]string, 0, len(styleMap))
			for styleKey := range styleMap {
				styleKeys = append(styleKeys, styleKey)
			}
			sort.Strings(styleKeys)

			var styleBuilder strings.Builder
			for _, styleKey := range styleKeys {
				styleBuilder.WriteString(fmt.Sprintf("%s:%s;", html.EscapeString(styleKey), renderAttributeValue(styleMap[styleKey])))
			}
			result.WriteString(fmt.Sprintf(" style=\"%s\"", styleBuilder.String()))
			continue
		}

		result.WriteString(fmt.Sprintf(" %s=\"%s\"", key, renderAttributeValue(value)))
	}

	return result.String()
}

func renderChildren(children []interface{}, raw bool) string {
	var result strings.Builder
	for _, child := range children {
		if child == nil {
			continue
		}

		rv := reflect.ValueOf(child)
		if rv.IsValid() && rv.Kind() == reflect.Slice && rv.Type().Elem().Kind() != reflect.Uint8 {
			for i := 0; i < rv.Len(); i++ {
				result.WriteString(renderChild(rv.Index(i).Interface(), raw))
			}
			continue
		}

		result.WriteString(renderChild(child, raw))
	}
	return result.String()
}

// CreateElement is the public version of our element creation function.
func CreateElement(component interface{}, props Props, children ...interface{}) string {
	var result strings.Builder

	switch c := component.(type) {
	case string:
		tag := strings.TrimSpace(c)
		if tag == "" {
			return ""
		}

		result.WriteString("<")
		result.WriteString(tag)
		result.WriteString(renderAttributes(props))

		if len(children) == 0 {
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
		result.WriteString(renderChildren(children, rawTextTags[tag]))
		result.WriteString("</")
		result.WriteString(tag)
		result.WriteString(">")
	case FunctionalComponent:
		if c == nil {
			return ""
		}
		result.WriteString(c(props, children...))
	case Component:
		result.WriteString(c.Render())
	case nil:
		result.WriteString("<!-- nil component -->")
	default:
		result.WriteString(fmt.Sprintf("<!-- unknown component type: %T -->", c))
	}

	return result.String()
}

// Fragment is a special component that renders only its children.
func Fragment(props Props, children ...interface{}) string {
	return renderChildren(children, false)
}
