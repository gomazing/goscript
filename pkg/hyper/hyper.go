package hyper

import (
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
	"time"
)

type node struct {
	Name     string
	Attr     map[string]string
	Text     string
	Children []*node
}

// MarshalIndent serializes a value into the Hyper text format.
func MarshalIndent(v interface{}, prefix, indent string) ([]byte, error) {
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	if prefix != "" || indent != "" {
		enc.Indent(prefix, indent)
	}

	rootKind := "value"
	rv := reflect.ValueOf(v)
	if rv.IsValid() {
		for rv.Kind() == reflect.Pointer || rv.Kind() == reflect.Interface {
			if rv.IsNil() {
				rv = reflect.Value{}
				break
			}
			rv = rv.Elem()
		}
		if rv.IsValid() && rv.Kind() == reflect.Struct {
			rootKind = rv.Type().Name()
			if rootKind == "" {
				rootKind = "struct"
			}
		}
	}

	root := xml.StartElement{
		Name: xml.Name{Local: "hyper"},
		Attr: []xml.Attr{
			{Name: xml.Name{Local: "kind"}, Value: rootKind},
			{Name: xml.Name{Local: "version"}, Value: "1"},
		},
	}
	if err := enc.EncodeToken(root); err != nil {
		return nil, err
	}

	if rv.IsValid() {
		if rv.Kind() == reflect.Struct {
			if err := encodeStructFields(enc, rv); err != nil {
				return nil, err
			}
		} else {
			if err := encodeValue(enc, rv, "value"); err != nil {
				return nil, err
			}
		}
	}

	if err := enc.EncodeToken(root.End()); err != nil {
		return nil, err
	}
	if err := enc.Flush(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// Marshal is a convenience wrapper around MarshalIndent.
func Marshal(v interface{}) ([]byte, error) {
	return MarshalIndent(v, "", "  ")
}

// WriteFile serializes a value to disk in Hyper format.
func WriteFile(path string, v interface{}) error {
	data, err := MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return ioWriteFile(path, data)
}

// NewEncoder creates a stream encoder for Hyper documents.
type Encoder struct {
	w      io.Writer
	prefix string
	indent string
}

// NewEncoder constructs an encoder that writes Hyper documents to w.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w: w, indent: "  "}
}

// SetIndent adjusts the indentation used by the encoder.
func (e *Encoder) SetIndent(prefix, indent string) {
	e.prefix = prefix
	e.indent = indent
}

// Encode writes a Hyper document to the underlying writer.
func (e *Encoder) Encode(v interface{}) error {
	data, err := MarshalIndent(v, e.prefix, e.indent)
	if err != nil {
		return err
	}
	_, err = e.w.Write(data)
	return err
}

// NewDecoder creates a stream decoder for Hyper documents.
type Decoder struct {
	data []byte
}

// NewDecoder reads all input and prepares it for decoding.
func NewDecoder(r io.Reader) *Decoder {
	data, _ := io.ReadAll(r)
	return &Decoder{data: data}
}

// Decode populates v from the Hyper document.
func (d *Decoder) Decode(v interface{}) error {
	return Unmarshal(d.data, v)
}

// Unmarshal decodes Hyper data into v.
func Unmarshal(data []byte, v interface{}) error {
	if v == nil {
		return errors.New("hyper: destination must not be nil")
	}

	target := reflect.ValueOf(v)
	if target.Kind() != reflect.Pointer || target.IsNil() {
		return errors.New("hyper: destination must be a non-nil pointer")
	}

	root, err := parseNode(bytes.NewReader(data))
	if err != nil {
		return err
	}

	dst := target.Elem()
	return decodeInto(root, dst)
}

func encodeValue(enc *xml.Encoder, value reflect.Value, name string) error {
	value = indirectValue(value)
	if !value.IsValid() {
		return nil
	}

	if marshaler, ok := value.Interface().(interface{ MarshalText() ([]byte, error) }); ok {
		text, err := marshaler.MarshalText()
		if err != nil {
			return err
		}
		return writeTextElement(enc, name, string(text), nil)
	}

	switch value.Kind() {
	case reflect.Struct:
		if value.Type() == reflect.TypeOf(time.Time{}) {
			return writeTextElement(enc, name, value.Interface().(time.Time).UTC().Format(time.RFC3339), nil)
		}
		start := xml.StartElement{Name: xml.Name{Local: name}}
		if err := enc.EncodeToken(start); err != nil {
			return err
		}
		if err := encodeStructFields(enc, value); err != nil {
			return err
		}
		return enc.EncodeToken(start.End())
	case reflect.Slice, reflect.Array:
		start := xml.StartElement{Name: xml.Name{Local: name}}
		if err := enc.EncodeToken(start); err != nil {
			return err
		}
		for i := 0; i < value.Len(); i++ {
			if err := encodeValue(enc, value.Index(i), "item"); err != nil {
				return err
			}
		}
		return enc.EncodeToken(start.End())
	case reflect.Map:
		start := xml.StartElement{Name: xml.Name{Local: name}}
		if err := enc.EncodeToken(start); err != nil {
			return err
		}

		keys := value.MapKeys()
		keyStrings := make([]string, 0, len(keys))
		keyValues := make(map[string]reflect.Value, len(keys))
		for _, key := range keys {
			keyString := fmt.Sprint(key.Interface())
			keyStrings = append(keyStrings, keyString)
			keyValues[keyString] = key
		}
		sort.Strings(keyStrings)

		for _, keyString := range keyStrings {
			entry := xml.StartElement{
				Name: xml.Name{Local: "item"},
				Attr: []xml.Attr{{Name: xml.Name{Local: "key"}, Value: keyString}},
			}
			if err := enc.EncodeToken(entry); err != nil {
				return err
			}
			if err := encodeValue(enc, value.MapIndex(keyValues[keyString]), "value"); err != nil {
				return err
			}
			if err := enc.EncodeToken(entry.End()); err != nil {
				return err
			}
		}
		return enc.EncodeToken(start.End())
	case reflect.String:
		return writeTextElement(enc, name, value.String(), nil)
	case reflect.Bool:
		return writeTextElement(enc, name, strconv.FormatBool(value.Bool()), nil)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if value.Type() == reflect.TypeOf(time.Duration(0)) {
			return writeTextElement(enc, name, time.Duration(value.Int()).String(), nil)
		}
		return writeTextElement(enc, name, strconv.FormatInt(value.Int(), 10), nil)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return writeTextElement(enc, name, strconv.FormatUint(value.Uint(), 10), nil)
	case reflect.Float32, reflect.Float64:
		return writeTextElement(enc, name, strconv.FormatFloat(value.Float(), 'f', -1, 64), nil)
	case reflect.Interface:
		if value.IsNil() {
			return nil
		}
		return encodeValue(enc, value.Elem(), name)
	case reflect.Pointer:
		if value.IsNil() {
			return nil
		}
		return encodeValue(enc, value.Elem(), name)
	default:
		return writeTextElement(enc, name, fmt.Sprint(value.Interface()), nil)
	}
}

func encodeStructFields(enc *xml.Encoder, value reflect.Value) error {
	typ := value.Type()
	for i := 0; i < value.NumField(); i++ {
		field := typ.Field(i)
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}

		name, omitempty, skip := hyperFieldName(field)
		if skip {
			continue
		}

		fieldValue := value.Field(i)
		if omitempty && fieldValue.IsZero() {
			continue
		}

		if field.Anonymous && fieldValue.Kind() == reflect.Struct && name == "" {
			if err := encodeStructFields(enc, fieldValue); err != nil {
				return err
			}
			continue
		}

		if name == "" {
			name = strings.ToLower(field.Name)
		}

		if err := encodeValue(enc, fieldValue, name); err != nil {
			return err
		}
	}
	return nil
}

func hyperFieldName(field reflect.StructField) (name string, omitempty bool, skip bool) {
	tag := field.Tag.Get("hyper")
	if tag == "" {
		tag = field.Tag.Get("json")
	}
	if tag == "-" {
		return "", false, true
	}
	if tag != "" {
		parts := strings.Split(tag, ",")
		name = parts[0]
		for _, part := range parts[1:] {
			if part == "omitempty" {
				omitempty = true
			}
		}
	}
	return name, omitempty, false
}

func writeTextElement(enc *xml.Encoder, name, text string, attrs []xml.Attr) error {
	start := xml.StartElement{Name: xml.Name{Local: name}, Attr: attrs}
	if err := enc.EncodeToken(start); err != nil {
		return err
	}
	if text != "" {
		if err := enc.EncodeToken(xml.CharData([]byte(text))); err != nil {
			return err
		}
	}
	return enc.EncodeToken(start.End())
}

func parseNode(r io.Reader) (*node, error) {
	dec := xml.NewDecoder(r)
	var stack []*node
	var root *node

	for {
		tok, err := dec.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}

		switch typed := tok.(type) {
		case xml.StartElement:
			current := &node{Name: typed.Name.Local, Attr: map[string]string{}}
			for _, attr := range typed.Attr {
				current.Attr[attr.Name.Local] = attr.Value
			}
			stack = append(stack, current)
		case xml.CharData:
			if len(stack) == 0 {
				continue
			}
			stack[len(stack)-1].Text += string(typed)
		case xml.EndElement:
			if len(stack) == 0 {
				continue
			}
			current := stack[len(stack)-1]
			stack = stack[:len(stack)-1]
			current.Text = strings.TrimSpace(current.Text)
			if len(stack) == 0 {
				if root == nil {
					root = current
				} else {
					root.Children = append(root.Children, current)
				}
			} else {
				parent := stack[len(stack)-1]
				parent.Children = append(parent.Children, current)
			}
		}
	}

	if root == nil {
		return nil, errors.New("hyper: document is empty")
	}

	return root, nil
}

func decodeInto(n *node, dst reflect.Value) error {
	if !dst.IsValid() {
		return nil
	}
	if dst.Kind() == reflect.Pointer {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return decodeInto(n, dst.Elem())
	}

	if dst.CanAddr() {
		if unmarshaler, ok := dst.Addr().Interface().(interface{ UnmarshalText([]byte) error }); ok {
			return unmarshaler.UnmarshalText([]byte(nodeText(n)))
		}
	}

	switch dst.Kind() {
	case reflect.Struct:
		if dst.Type() == reflect.TypeOf(time.Time{}) {
			parsed, err := time.Parse(time.RFC3339, nodeText(n))
			if err != nil {
				return err
			}
			dst.Set(reflect.ValueOf(parsed))
			return nil
		}

		for i := 0; i < dst.NumField(); i++ {
			field := dst.Type().Field(i)
			if field.PkgPath != "" && !field.Anonymous {
				continue
			}
			name, _, skip := hyperFieldName(field)
			if skip {
				continue
			}
			if field.Anonymous && dst.Field(i).Kind() == reflect.Struct && name == "" {
				if err := decodeInto(n, dst.Field(i)); err != nil {
					return err
				}
				continue
			}
			if name == "" {
				name = strings.ToLower(field.Name)
			}
			child := firstChild(n, name)
			if child == nil {
				continue
			}
			if err := decodeField(child, dst.Field(i)); err != nil {
				return fmt.Errorf("hyper: decode field %s: %w", field.Name, err)
			}
		}
		return nil
	case reflect.Slice:
		items := childrenForSlice(n)
		slice := reflect.MakeSlice(dst.Type(), 0, len(items))
		for _, item := range items {
			elem := reflect.New(dst.Type().Elem()).Elem()
			if err := decodeInto(item, elem); err != nil {
				return err
			}
			slice = reflect.Append(slice, elem)
		}
		dst.Set(slice)
		return nil
	case reflect.Array:
		items := childrenForSlice(n)
		limit := len(items)
		if limit > dst.Len() {
			limit = dst.Len()
		}
		for i := 0; i < limit; i++ {
			if err := decodeInto(items[i], dst.Index(i)); err != nil {
				return err
			}
		}
		return nil
	case reflect.Map:
		if dst.IsNil() {
			dst.Set(reflect.MakeMap(dst.Type()))
		}
		for _, child := range n.Children {
			key := child.Attr["key"]
			if key == "" {
				key = child.Name
			}
			elem := reflect.New(dst.Type().Elem()).Elem()
			if err := decodeInto(child, elem); err != nil {
				return err
			}
			mapKey := reflect.ValueOf(key).Convert(dst.Type().Key())
			dst.SetMapIndex(mapKey, elem)
		}
		return nil
	case reflect.Interface:
		dst.Set(reflect.ValueOf(nodeToAny(n)))
		return nil
	case reflect.String:
		dst.SetString(nodeText(n))
		return nil
	case reflect.Bool:
		parsed, err := strconv.ParseBool(nodeText(n))
		if err != nil {
			return err
		}
		dst.SetBool(parsed)
		return nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if dst.Type() == reflect.TypeOf(time.Duration(0)) {
			parsed, err := time.ParseDuration(nodeText(n))
			if err != nil {
				return err
			}
			dst.SetInt(int64(parsed))
			return nil
		}
		parsed, err := strconv.ParseInt(nodeText(n), 10, 64)
		if err != nil {
			return err
		}
		dst.SetInt(parsed)
		return nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		parsed, err := strconv.ParseUint(nodeText(n), 10, 64)
		if err != nil {
			return err
		}
		dst.SetUint(parsed)
		return nil
	case reflect.Float32, reflect.Float64:
		parsed, err := strconv.ParseFloat(nodeText(n), 64)
		if err != nil {
			return err
		}
		dst.SetFloat(parsed)
		return nil
	default:
		if dst.CanSet() && dst.Kind() == reflect.Pointer && dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
			return decodeInto(n, dst.Elem())
		}
	}

	if dst.CanAddr() {
		if unmarshaler, ok := dst.Addr().Interface().(interface{ UnmarshalText([]byte) error }); ok {
			return unmarshaler.UnmarshalText([]byte(nodeText(n)))
		}
	}

	return nil
}

func decodeField(n *node, dst reflect.Value) error {
	if dst.Kind() == reflect.Pointer {
		if dst.IsNil() {
			dst.Set(reflect.New(dst.Type().Elem()))
		}
		return decodeInto(n, dst.Elem())
	}

	switch dst.Kind() {
	case reflect.Struct, reflect.Slice, reflect.Array, reflect.Map, reflect.Interface:
		return decodeInto(n, dst)
	default:
		return decodeInto(n, dst)
	}
}

func nodeText(n *node) string {
	text := strings.TrimSpace(n.Text)
	if text != "" {
		return text
	}
	if len(n.Children) == 0 {
		return ""
	}
	var parts []string
	for _, child := range n.Children {
		if child.Text != "" {
			parts = append(parts, strings.TrimSpace(child.Text))
		}
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}

func firstChild(n *node, name string) *node {
	for _, child := range n.Children {
		if child.Name == name {
			return child
		}
	}
	return nil
}

func childrenForSlice(n *node) []*node {
	if len(n.Children) == 0 {
		if strings.TrimSpace(n.Text) == "" {
			return nil
		}
		return []*node{{Name: "item", Text: strings.TrimSpace(n.Text)}}
	}

	items := make([]*node, 0, len(n.Children))
	for _, child := range n.Children {
		if child.Name == "item" || child.Name == "entry" {
			items = append(items, child)
			continue
		}
		if len(n.Children) == 1 {
			items = append(items, child)
		}
	}

	if len(items) == 0 {
		items = append(items, n.Children...)
	}
	return items
}

func nodeToAny(n *node) interface{} {
	if len(n.Children) == 0 {
		return parseScalar(nodeText(n))
	}

	if len(n.Children) > 0 && allChildrenNamed(n.Children, "item") && len(n.Children[0].Attr) > 0 {
		result := map[string]interface{}{}
		for _, child := range n.Children {
			key := child.Attr["key"]
			if key == "" {
				key = child.Name
			}
			result[key] = nodeToAny(child)
		}
		return result
	}

	if allChildrenNamed(n.Children, "item") {
		result := make([]interface{}, 0, len(n.Children))
		for _, child := range n.Children {
			result = append(result, nodeToAny(child))
		}
		return result
	}

	result := map[string]interface{}{}
	for _, child := range n.Children {
		value := nodeToAny(child)
		if existing, ok := result[child.Name]; ok {
			switch typed := existing.(type) {
			case []interface{}:
				result[child.Name] = append(typed, value)
			default:
				result[child.Name] = []interface{}{typed, value}
			}
			continue
		}
		result[child.Name] = value
	}
	return result
}

func allChildrenNamed(children []*node, name string) bool {
	if len(children) == 0 {
		return false
	}
	for _, child := range children {
		if child.Name != name {
			return false
		}
	}
	return true
}

func parseScalar(raw string) interface{} {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	if trimmed == "true" {
		return true
	}
	if trimmed == "false" {
		return false
	}
	if strings.ContainsAny(trimmed, ".eE") {
		if f, err := strconv.ParseFloat(trimmed, 64); err == nil {
			return f
		}
	}
	if i, err := strconv.ParseInt(trimmed, 10, 64); err == nil {
		return i
	}
	return trimmed
}

func indirectValue(value reflect.Value) reflect.Value {
	for value.IsValid() && (value.Kind() == reflect.Pointer || value.Kind() == reflect.Interface) {
		if value.IsNil() {
			return reflect.Value{}
		}
		value = value.Elem()
	}
	return value
}

func ioWriteFile(path string, data []byte) error {
	return writeFile(path, data)
}

func writeFile(path string, data []byte) error {
	if err := os.MkdirAll(filepathDir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func filepathDir(path string) string {
	if idx := strings.LastIndexAny(path, `/\`); idx >= 0 {
		return path[:idx]
	}
	return "."
}
