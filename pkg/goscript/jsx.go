package goscript

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// JSXParser parses JSX-like syntax into CreateElement calls.
type JSXParser struct {
	// Configuration options
	options map[string]interface{}
}

var (
	jsxCompileCache sync.Map
	gsxCompileCache sync.Map

	packageRegex = regexp.MustCompile(`package\s+([a-zA-Z][a-zA-Z0-9]*)`)
	importRegex  = regexp.MustCompile(`import\s+\(([^)]+)\)`)
	funcRegex    = regexp.MustCompile(`func\s+([a-zA-Z][a-zA-Z0-9]*)\s*\(([^)]*)\)\s*string\s*{([^}]+)}`)
)

// NewJSXParser creates a new JSX parser.
func NewJSXParser(options map[string]interface{}) *JSXParser {
	return &JSXParser{
		options: options,
	}
}

// ParseJSX parses a JSX-like string into Go code.
//
// The parser intentionally favors deterministic lowering over runtime regex replacement.
// It converts a single JSX fragment into nested CreateElement calls, which is what Go FAST
// wants for build-time lowering and hot-path reuse.
func (p *JSXParser) ParseJSX(jsx string) (string, error) {
	normalized := strings.TrimSpace(jsx)
	if normalized == "" {
		return "", nil
	}

	if cached, ok := jsxCompileCache.Load(normalized); ok {
		return cached.(string), nil
	}

	parser := &jsxExpressionParser{src: normalized}
	expr, err := parser.parseExpression()
	if err != nil {
		return "", err
	}

	jsxCompileCache.Store(normalized, expr)
	return expr, nil
}

// ParseSurface lowers JSX-like syntax directly into a Graft graph.
func (p *JSXParser) ParseSurface(jsx string) (GraftNode, error) {
	normalized := strings.TrimSpace(jsx)
	if normalized == "" {
		return GraftFragment(), nil
	}

	parser := &jsxExpressionParser{src: normalized}
	node, err := parser.parseNode()
	if err != nil {
		return GraftNode{}, err
	}
	if node == nil {
		return GraftText(normalized), nil
	}

	return lowerJSXNodeToGraft(node), nil
}

// ParseAttributes parses JSX attributes into Props.
func (p *JSXParser) ParseAttributes(attrs string) (Props, error) {
	parser := &jsxExpressionParser{src: strings.TrimSpace(attrs)}
	list, err := parser.parseAttributeList()
	if err != nil {
		return nil, err
	}

	props := Props{}
	for _, attr := range list {
		switch attr.kind {
		case jsxAttrBool:
			props[attr.key] = true
		case jsxAttrExpr:
			props[attr.key] = RawHTML(attr.value)
		default:
			props[attr.key] = attr.value
		}
	}
	return props, nil
}

// TranspileGSX transpiles a .gsx file to a .go file.
func TranspileGSX(gsxContent string) (string, error) {
	normalized := strings.TrimSpace(gsxContent)
	if normalized == "" {
		return "", nil
	}

	if cached, ok := gsxCompileCache.Load(normalized); ok {
		return cached.(string), nil
	}

	// This still keeps file-level parsing lightweight; the expensive bit is JSX lowering,
	// which is now handled by a real parser instead of regex replacement.
	importsMatch := importRegex.FindStringSubmatch(normalized)
	imports := ""
	if len(importsMatch) > 1 {
		imports = importsMatch[1]
	}

	packageMatch := packageRegex.FindStringSubmatch(normalized)
	packageName := ""
	if len(packageMatch) > 1 {
		packageName = packageMatch[1]
	}
	if packageName == "" {
		packageName = "main"
	}

	funcMatches := funcRegex.FindAllStringSubmatch(normalized, -1)
	var functions []string

	parser := NewJSXParser(nil)
	for _, match := range funcMatches {
		if len(match) < 4 {
			continue
		}

		funcName := match[1]
		funcParams := match[2]
		funcBody := strings.TrimSpace(match[3])

		hasReturn := strings.HasPrefix(funcBody, "return ")
		if hasReturn {
			funcBody = strings.TrimSpace(strings.TrimPrefix(funcBody, "return"))
		}

		compiledBody := funcBody
		if strings.Contains(funcBody, "<") {
			parsedBody, err := parser.ParseJSX(funcBody)
			if err != nil {
				return "", fmt.Errorf("error parsing JSX in function %s: %v", funcName, err)
			}
			compiledBody = parsedBody
		}

		compiledBody = strings.ReplaceAll(compiledBody, "goscript.createElement", "goscript.CreateElement")
		compiledBody = strings.ReplaceAll(compiledBody, "createElement(", "CreateElement(")

		if hasReturn && !strings.HasPrefix(strings.TrimSpace(compiledBody), "return ") {
			compiledBody = "return " + compiledBody
		}

		function := fmt.Sprintf("func %s(%s) string {\n%s\n}", funcName, funcParams, compiledBody)
		functions = append(functions, function)
	}

	var result strings.Builder
	result.WriteString("package ")
	result.WriteString(packageName)
	result.WriteString("\n\n")

	if strings.TrimSpace(imports) != "" {
		result.WriteString("import (\n")
		result.WriteString(imports)
		result.WriteString("\n)\n\n")
	}

	if len(functions) > 0 {
		result.WriteString(strings.Join(functions, "\n\n"))
		result.WriteString("\n")
	} else {
		result.WriteString("// no function declarations found\n")
	}

	compiled := result.String()
	gsxCompileCache.Store(normalized, compiled)
	return compiled, nil
}

// GSXCompiler compiles .gsx files to .go files.
type GSXCompiler struct {
	// Configuration options
	options map[string]interface{}
}

// NewGSXCompiler creates a new GSX compiler.
func NewGSXCompiler(options map[string]interface{}) *GSXCompiler {
	return &GSXCompiler{
		options: options,
	}
}

// CompileFile compiles a .gsx file to a .go file.
func (c *GSXCompiler) CompileFile(inputPath, outputPath string) error {
	if strings.TrimSpace(inputPath) == "" {
		return fmt.Errorf("input path is required")
	}

	if strings.TrimSpace(outputPath) == "" {
		ext := filepath.Ext(inputPath)
		outputPath = strings.TrimSuffix(inputPath, ext) + ".go"
	}

	contents, err := os.ReadFile(inputPath)
	if err != nil {
		return err
	}

	compiled, err := TranspileGSX(string(contents))
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return err
	}

	return os.WriteFile(outputPath, []byte(compiled), 0644)
}

type jsxAttrKind int

const (
	jsxAttrString jsxAttrKind = iota
	jsxAttrExpr
	jsxAttrBool
)

type jsxAttribute struct {
	key   string
	value string
	kind  jsxAttrKind
}

type jsxChild struct {
	text     string
	element  *jsxNode
}

type jsxNode struct {
	tag        string
	fragment   bool
	selfClosing bool
	attrs      []jsxAttribute
	children   []jsxChild
}

type jsxExpressionParser struct {
	src string
	pos int
}

func (p *jsxExpressionParser) parseExpression() (string, error) {
	node, err := p.parseNode()
	if err != nil {
		return "", err
	}
	if node == nil {
		return strings.TrimSpace(p.src), nil
	}

	return emitJSXNode(node), nil
}

func (p *jsxExpressionParser) parseNode() (*jsxNode, error) {
	idx := strings.Index(p.src[p.pos:], "<")
	if idx < 0 {
		return nil, nil
	}
	p.pos += idx

	node, err := p.parseElement()
	if err != nil {
		return nil, err
	}

	return node, nil
}

func (p *jsxExpressionParser) parseElement() (*jsxNode, error) {
	if !p.consume("<") {
		return nil, fmt.Errorf("expected < at position %d", p.pos)
	}

	node := &jsxNode{}
	if p.consume(">") {
		node.fragment = true
	} else {
		tag, err := p.parseName()
		if err != nil {
			return nil, err
		}
		if tag == "" {
			return nil, fmt.Errorf("expected tag name at position %d", p.pos)
		}
		node.tag = tag

		attrs, selfClosing, err := p.parseAttributeList()
		if err != nil {
			return nil, err
		}
		node.attrs = attrs
		node.selfClosing = selfClosing

		if node.selfClosing {
			return node, nil
		}
	}

	if node.fragment {
		children, err := p.parseChildren("", true)
		if err != nil {
			return nil, err
		}
		node.children = children
		return node, nil
	}

	if rawTextTags[node.tag] {
		children, err := p.parseRawTextChildren(node.tag)
		if err != nil {
			return nil, err
		}
		node.children = children
		return node, nil
	}

	children, err := p.parseChildren(node.tag, false)
	if err != nil {
		return nil, err
	}
	node.children = children
	return node, nil
}

func (p *jsxExpressionParser) parseAttributeList() ([]jsxAttribute, bool, error) {
	var attrs []jsxAttribute

	for {
		p.skipSpaces()
		if p.eof() {
			return attrs, false, nil
		}

		if p.consume("/>") {
			return attrs, true, nil
		}
		if p.consume(">") {
			return attrs, false, nil
		}

		key, err := p.parseName()
		if err != nil {
			return nil, false, err
		}
		if key == "" {
			return nil, false, fmt.Errorf("expected attribute name at position %d", p.pos)
		}

		p.skipSpaces()
		if !p.consume("=") {
			attrs = append(attrs, jsxAttribute{key: key, kind: jsxAttrBool, value: "true"})
			continue
		}

		p.skipSpaces()
		kind, value, err := p.parseAttributeValue()
		if err != nil {
			return nil, false, err
		}
		attrs = append(attrs, jsxAttribute{key: key, kind: kind, value: value})
	}
}

func (p *jsxExpressionParser) parseAttributeValue() (jsxAttrKind, string, error) {
	if p.eof() {
		return jsxAttrString, "", fmt.Errorf("unexpected end of input while parsing attribute value")
	}

	switch p.src[p.pos] {
	case '\'', '"':
		quote := p.src[p.pos]
		p.pos++
		value, err := p.parseQuotedValue(quote)
		if err != nil {
			return jsxAttrString, "", err
		}
		return jsxAttrString, value, nil
	case '{':
		expr, err := p.parseBracedExpression()
		if err != nil {
			return jsxAttrExpr, "", err
		}
		return jsxAttrExpr, expr, nil
	default:
		token := p.parseBareToken()
		switch token {
		case "true", "false":
			return jsxAttrBool, token, nil
		default:
			return jsxAttrString, token, nil
		}
	}
}

func (p *jsxExpressionParser) parseQuotedValue(quote byte) (string, error) {
	var out strings.Builder
	for !p.eof() {
		ch := p.src[p.pos]
		p.pos++
		if ch == quote {
			return out.String(), nil
		}
		if ch == '\\' && !p.eof() {
			next := p.src[p.pos]
			p.pos++
			out.WriteByte(next)
			continue
		}
		out.WriteByte(ch)
	}
	return "", fmt.Errorf("unterminated quoted string")
}

func (p *jsxExpressionParser) parseBracedExpression() (string, error) {
	if !p.consume("{") {
		return "", fmt.Errorf("expected { at position %d", p.pos)
	}

	depth := 1
	start := p.pos
	inString := byte(0)

	for !p.eof() {
		ch := p.src[p.pos]
		p.pos++

		if inString != 0 {
			if ch == '\\' && !p.eof() {
				p.pos++
				continue
			}
			if ch == inString {
				inString = 0
			}
			continue
		}

		switch ch {
		case '\'', '"', '`':
			inString = ch
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return strings.TrimSpace(p.src[start : p.pos-1]), nil
			}
		}
	}

	return "", fmt.Errorf("unterminated braced expression")
}

func (p *jsxExpressionParser) parseBareToken() string {
	start := p.pos
	for !p.eof() {
		ch := p.src[p.pos]
		if isTokenTerminator(ch) {
			break
		}
		p.pos++
	}
	return strings.TrimSpace(p.src[start:p.pos])
}

func (p *jsxExpressionParser) parseName() (string, error) {
	start := p.pos
	if p.eof() {
		return "", nil
	}
	if !isNameStart(p.src[p.pos]) {
		return "", nil
	}
	p.pos++
	for !p.eof() {
		ch := p.src[p.pos]
		if !isNamePart(ch) {
			break
		}
		p.pos++
	}
	return p.src[start:p.pos], nil
}

func (p *jsxExpressionParser) parseChildren(tag string, fragment bool) ([]jsxChild, error) {
	var children []jsxChild

	for !p.eof() {
		p.skipSpacesBetweenNodes()

		if fragment {
			if p.consume("</>") {
				return children, nil
			}
		} else if strings.HasPrefix(p.src[p.pos:], "</") {
			if !p.consume("</") {
				return nil, fmt.Errorf("expected closing tag at position %d", p.pos)
			}
			closeTag, err := p.parseName()
			if err != nil {
				return nil, err
			}
			if closeTag != tag {
				return nil, fmt.Errorf("expected closing tag </%s> but found </%s>", tag, closeTag)
			}
			if !p.consume(">") {
				return nil, fmt.Errorf("expected > after closing tag </%s>", tag)
			}
			return children, nil
		}

		if p.eof() {
			break
		}

		if p.src[p.pos] == '<' {
			child, err := p.parseElement()
			if err != nil {
				return nil, err
			}
			children = append(children, jsxChild{element: child})
			continue
		}

		text := p.parseText()
		if strings.TrimSpace(text) == "" {
			continue
		}
		children = append(children, jsxChild{text: text})
	}

	if fragment {
		return nil, fmt.Errorf("missing closing fragment tag")
	}

	return nil, fmt.Errorf("missing closing tag </%s>", tag)
}

func (p *jsxExpressionParser) parseRawTextChildren(tag string) ([]jsxChild, error) {
	closing := "</" + tag + ">"
	offset := strings.Index(p.src[p.pos:], closing)
	if offset < 0 {
		return nil, fmt.Errorf("missing closing tag </%s>", tag)
	}

	raw := p.src[p.pos : p.pos+offset]
	p.pos += offset
	if !p.consume(closing) {
		return nil, fmt.Errorf("missing closing tag </%s>", tag)
	}

	if strings.TrimSpace(raw) == "" {
		return nil, nil
	}

	return []jsxChild{{text: raw}}, nil
}

func (p *jsxExpressionParser) parseText() string {
	start := p.pos
	for !p.eof() && p.src[p.pos] != '<' {
		p.pos++
	}
	return p.src[start:p.pos]
}

func (p *jsxExpressionParser) skipSpaces() {
	for !p.eof() {
		if !isSpace(p.src[p.pos]) {
			return
		}
		p.pos++
	}
}

func (p *jsxExpressionParser) skipSpacesBetweenNodes() {
	for !p.eof() {
		if !isSpace(p.src[p.pos]) {
			return
		}
		p.pos++
	}
}

func (p *jsxExpressionParser) consume(s string) bool {
	if strings.HasPrefix(p.src[p.pos:], s) {
		p.pos += len(s)
		return true
	}
	return false
}

func (p *jsxExpressionParser) eof() bool {
	return p.pos >= len(p.src)
}

func isSpace(ch byte) bool {
	switch ch {
	case ' ', '\t', '\n', '\r':
		return true
	default:
		return false
	}
}

func isTokenTerminator(ch byte) bool {
	return isSpace(ch) || ch == '>' || ch == '/' || ch == '='
}

func isNameStart(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' || ch == ':'
}

func isNamePart(ch byte) bool {
	return isNameStart(ch) || (ch >= '0' && ch <= '9') || ch == '-' || ch == '.'
}

func emitJSXNode(node *jsxNode) string {
	if node == nil {
		return ""
	}

	if node.fragment {
		return emitFragment(node.children)
	}

	var builder strings.Builder
	builder.WriteString("CreateElement(")
	builder.WriteString(strconv.Quote(node.tag))
	builder.WriteString(", ")

	propsLiteral := emitPropsLiteral(node.attrs)
	if propsLiteral == "" {
		builder.WriteString("nil")
	} else {
		builder.WriteString(propsLiteral)
	}

	for _, child := range node.children {
		if child.element == nil && strings.TrimSpace(child.text) == "" {
			continue
		}
		builder.WriteString(", ")
		builder.WriteString(emitJSXChild(child))
	}

	builder.WriteString(")")
	return builder.String()
}

func emitFragment(children []jsxChild) string {
	var builder strings.Builder
	builder.WriteString("Fragment(nil")
	for _, child := range children {
		if child.element == nil && strings.TrimSpace(child.text) == "" {
			continue
		}
		builder.WriteString(", ")
		builder.WriteString(emitJSXChild(child))
	}
	builder.WriteString(")")
	return builder.String()
}

func emitJSXChild(child jsxChild) string {
	if child.element != nil {
		return emitJSXNode(child.element)
	}
	return strconv.Quote(child.text)
}

func emitPropsLiteral(attrs []jsxAttribute) string {
	if len(attrs) == 0 {
		return ""
	}

	sorted := append([]jsxAttribute(nil), attrs...)
	sort.SliceStable(sorted, func(i, j int) bool {
		return sorted[i].key < sorted[j].key
	})

	var builder strings.Builder
	builder.WriteString("Props{")
	needsComma := false
	for _, attr := range sorted {
		switch attr.kind {
		case jsxAttrBool:
			if attr.value == "false" {
				continue
			}
			if needsComma {
				builder.WriteString(", ")
			}
			builder.WriteString(strconv.Quote(attr.key))
			builder.WriteString(": true")
			needsComma = true
		case jsxAttrExpr:
			if needsComma {
				builder.WriteString(", ")
			}
			builder.WriteString(strconv.Quote(attr.key))
			builder.WriteString(": ")
			builder.WriteString(attr.value)
			needsComma = true
		default:
			if needsComma {
				builder.WriteString(", ")
			}
			builder.WriteString(strconv.Quote(attr.key))
			builder.WriteString(": ")
			builder.WriteString(strconv.Quote(attr.value))
			needsComma = true
		}
	}
	builder.WriteString("}")

	if !needsComma {
		return ""
	}

	return builder.String()
}

func lowerJSXNodeToGraft(node *jsxNode) GraftNode {
	if node == nil {
		return GraftFragment()
	}

	if node.fragment {
		return GraftFragment(lowerJSXChildrenToGraft(node, node.children)...)
	}

	props := lowerJSXProps(node.attrs)
	children := lowerJSXChildrenToGraft(node, node.children)
	return Graft(node.tag, props, children...)
}

func lowerJSXChildrenToGraft(parent *jsxNode, children []jsxChild) []GraftNode {
	if len(children) == 0 {
		return nil
	}

	out := make([]GraftNode, 0, len(children))
	for _, child := range children {
		if child.element != nil {
			out = append(out, lowerJSXNodeToGraft(child.element))
			continue
		}

		text := strings.TrimSpace(child.text)
		if text == "" {
			continue
		}

		if parent != nil && rawTextTags[parent.tag] {
			out = append(out, GraftRaw(child.text))
			continue
		}

		out = append(out, GraftText(child.text))
	}
	return out
}

func lowerJSXProps(attrs []jsxAttribute) Props {
	if len(attrs) == 0 {
		return nil
	}

	props := make(Props, len(attrs))
	for _, attr := range attrs {
		switch attr.kind {
		case jsxAttrBool:
			if attr.value != "false" {
				props[attr.key] = true
			}
		case jsxAttrExpr:
			props[attr.key] = RawHTML(attr.value)
		default:
			props[attr.key] = attr.value
		}
	}
	if len(props) == 0 {
		return nil
	}
	return props
}
