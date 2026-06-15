package goscript

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"sync"
)

// TalkTextCodec carries human-readable TALK payloads.
type TalkTextCodec struct{}

// ContentType returns the codec MIME type.
func (TalkTextCodec) ContentType() string { return "text/plain; charset=utf-8" }

// Encode serializes a frame into text payload bytes.
func (TalkTextCodec) Encode(frame TalkFrame) ([]byte, error) {
	frame = frame.Normalize()
	switch value := frame.Payload.(type) {
	case nil:
		return []byte{}, nil
	case []byte:
		return append([]byte(nil), value...), nil
	case string:
		return []byte(value), nil
	default:
		return json.Marshal(value)
	}
}

// Decode deserializes a text payload into a TALK frame.
func (TalkTextCodec) Decode(data []byte) (TalkFrame, error) {
	return TalkFrame{
		Kind:      TalkFrameKindText,
		Direction: TalkDirectionBidirectional,
		Payload:   string(data),
	}.Normalize(), nil
}

// TalkBinaryCodec carries raw or sensor-style payloads.
type TalkBinaryCodec struct{}

// ContentType returns the codec MIME type.
func (TalkBinaryCodec) ContentType() string { return "application/octet-stream" }

// Encode serializes a frame into binary payload bytes.
func (TalkBinaryCodec) Encode(frame TalkFrame) ([]byte, error) {
	frame = frame.Normalize()
	switch value := frame.Payload.(type) {
	case nil:
		return []byte{}, nil
	case []byte:
		return append([]byte(nil), value...), nil
	case string:
		return []byte(value), nil
	default:
		return json.Marshal(value)
	}
}

// Decode deserializes a binary payload into a TALK frame.
func (TalkBinaryCodec) Decode(data []byte) (TalkFrame, error) {
	return TalkFrame{
		Kind:      TalkFrameKindData,
		Direction: TalkDirectionBidirectional,
		Payload:   append([]byte(nil), data...),
	}.Normalize(), nil
}

// TalkCodecRegistry resolves codecs for content negotiation.
type TalkCodecRegistry struct {
	mu           sync.RWMutex
	codecs       map[string]TalkCodec
	order        []string
	defaultCodec TalkCodec
}

// NewTalkCodecRegistry creates a registry with the canonical codecs installed.
func NewTalkCodecRegistry() *TalkCodecRegistry {
	registry := &TalkCodecRegistry{
		codecs: make(map[string]TalkCodec),
	}
	registry.Register(TalkJSONCodec{})
	registry.Register(TalkTextCodec{})
	registry.Register(TalkBinaryCodec{})
	registry.defaultCodec = TalkJSONCodec{}
	return registry
}

// Register stores or replaces a codec by content type.
func (r *TalkCodecRegistry) Register(codec TalkCodec) {
	if r == nil || codec == nil {
		return
	}

	contentType := normalizeTalkContentType(codec.ContentType())
	if contentType == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.codecs == nil {
		r.codecs = make(map[string]TalkCodec)
	}
	if _, exists := r.codecs[contentType]; !exists {
		r.order = append(r.order, contentType)
	}
	r.codecs[contentType] = codec
	if r.defaultCodec == nil {
		r.defaultCodec = codec
	}
}

// Resolve finds a codec by content type.
func (r *TalkCodecRegistry) Resolve(contentType string) (TalkCodec, bool) {
	if r == nil {
		return nil, false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	codec, ok := r.codecs[normalizeTalkContentType(contentType)]
	return codec, ok
}

// ContentTypes returns registered codec content types in registration order.
func (r *TalkCodecRegistry) ContentTypes() []string {
	if r == nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]string, len(r.order))
	copy(out, r.order)
	return out
}

// Negotiate chooses the best codec for a contract and request headers.
func (r *TalkCodecRegistry) Negotiate(profile TalkProfile, acceptHeader, contentTypeHeader string) TalkCodec {
	if r == nil {
		return TalkJSONCodec{}
	}

	if codec, ok := r.Resolve(contentTypeHeader); ok {
		return codec
	}

	for _, token := range parseTalkAcceptHeader(acceptHeader) {
		if codec, ok := r.Resolve(token); ok {
			return codec
		}
	}

	if profile.Media || profile.Sensors {
		if codec, ok := r.Resolve("application/octet-stream"); ok {
			return codec
		}
	}

	if profile.Text && !profile.Data {
		if codec, ok := r.Resolve("text/plain"); ok {
			return codec
		}
	}

	if profile.Streaming && !profile.Text {
		if codec, ok := r.Resolve("application/octet-stream"); ok {
			return codec
		}
	}

	if codec, ok := r.Resolve("application/json"); ok {
		return codec
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.defaultCodec != nil {
		return r.defaultCodec
	}

	return TalkJSONCodec{}
}

// TalkRuntime manages TALK endpoints, codecs, and transport sessions.
type TalkRuntime struct {
	mu        sync.RWMutex
	codecs    *TalkCodecRegistry
	bus       *TalkBus
	endpoints map[string]TalkEndpoint
	router    *Router
}

// NewTalkRuntime creates a ready-to-use TALK runtime.
func NewTalkRuntime() *TalkRuntime {
	return &TalkRuntime{
		codecs:    NewTalkCodecRegistry(),
		bus:       NewTalkBus(),
		endpoints: make(map[string]TalkEndpoint),
		router:    NewRouter(),
	}
}

func (r *TalkRuntime) ensure() {
	if r.codecs == nil {
		r.codecs = NewTalkCodecRegistry()
	}
	if r.bus == nil {
		r.bus = NewTalkBus()
	}
	if r.endpoints == nil {
		r.endpoints = make(map[string]TalkEndpoint)
	}
	if r.router == nil {
		r.router = NewRouter()
	}
}

// RegisterCodec adds a codec to the negotiation registry.
func (r *TalkRuntime) RegisterCodec(codec TalkCodec) {
	if r == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()
	r.ensure()
	r.codecs.Register(codec)
	r.rebuildRouterLocked()
}

// Codec resolves a codec by content type.
func (r *TalkRuntime) Codec(contentType string) (TalkCodec, bool) {
	if r == nil {
		return nil, false
	}

	r.mu.RLock()
	codecs := r.codecs
	r.mu.RUnlock()
	if codecs == nil {
		return nil, false
	}
	return codecs.Resolve(contentType)
}

// RegisterEndpoint stores a TALK endpoint and mounts it into the runtime router.
func (r *TalkRuntime) RegisterEndpoint(endpoint TalkEndpoint) error {
	if r == nil {
		return fmt.Errorf("talk runtime is nil")
	}

	endpoint = endpoint.Normalize()
	if err := endpoint.Validate(); err != nil {
		return err
	}

	key := talkEndpointKey(endpoint.Contract.Method, endpoint.Contract.Path)

	r.mu.Lock()
	defer r.mu.Unlock()
	r.ensure()
	r.endpoints[key] = endpoint
	r.rebuildRouterLocked()
	return nil
}

// Endpoint returns a registered endpoint by method and path.
func (r *TalkRuntime) Endpoint(method, path string) (TalkEndpoint, bool) {
	if r == nil {
		return TalkEndpoint{}, false
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	endpoint, ok := r.endpoints[talkEndpointKey(method, path)]
	return endpoint, ok
}

// Endpoints returns all registered endpoints in deterministic order.
func (r *TalkRuntime) Endpoints() []TalkEndpoint {
	if r == nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	out := make([]TalkEndpoint, 0, len(r.endpoints))
	for _, endpoint := range r.endpoints {
		out = append(out, endpoint)
	}
	sort.SliceStable(out, func(i, j int) bool {
		if out[i].Contract.Path == out[j].Contract.Path {
			return strings.ToUpper(out[i].Contract.Method) < strings.ToUpper(out[j].Contract.Method)
		}
		return out[i].Contract.Path < out[j].Contract.Path
	})
	return out
}

// Bus returns the session bus used by the TALK runtime.
func (r *TalkRuntime) Bus() *TalkBus {
	if r == nil {
		return nil
	}

	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.bus
}

// OpenSession opens a transport session on the internal bus.
func (r *TalkRuntime) OpenSession(id, topic string, profile TalkProfile) *TalkSession {
	if r == nil {
		return nil
	}

	r.mu.RLock()
	bus := r.bus
	r.mu.RUnlock()
	if bus == nil {
		return nil
	}
	return bus.OpenSession(id, topic, profile)
}

// Publish forwards a frame through the named session.
func (r *TalkRuntime) Publish(sessionID string, frame TalkFrame) (TalkFrame, error) {
	if r == nil {
		return TalkFrame{}, fmt.Errorf("talk runtime is nil")
	}

	r.mu.RLock()
	bus := r.bus
	r.mu.RUnlock()
	if bus == nil {
		return TalkFrame{}, fmt.Errorf("talk bus is unavailable")
	}
	return bus.Publish(sessionID, frame)
}

// Subscribe attaches a subscriber to the named session.
func (r *TalkRuntime) Subscribe(sessionID, subscriberID string, buffer int) (<-chan TalkFrame, error) {
	if r == nil {
		return nil, fmt.Errorf("talk runtime is nil")
	}

	r.mu.RLock()
	bus := r.bus
	r.mu.RUnlock()
	if bus == nil {
		return nil, fmt.Errorf("talk bus is unavailable")
	}
	return bus.Subscribe(sessionID, subscriberID, buffer)
}

// Unsubscribe removes a subscriber from the named session.
func (r *TalkRuntime) Unsubscribe(sessionID, subscriberID string) {
	if r == nil {
		return
	}

	r.mu.RLock()
	bus := r.bus
	r.mu.RUnlock()
	if bus == nil {
		return
	}
	bus.Unsubscribe(sessionID, subscriberID)
}

// CloseSession shuts down and removes a named session.
func (r *TalkRuntime) CloseSession(sessionID string) {
	if r == nil {
		return
	}

	r.mu.RLock()
	bus := r.bus
	r.mu.RUnlock()
	if bus == nil {
		return
	}
	bus.CloseSession(sessionID)
}

// Mount registers the runtime endpoints into an external router.
func (r *TalkRuntime) Mount(router *Router) {
	if r == nil || router == nil {
		return
	}

	for _, endpoint := range r.Endpoints() {
		router.Handle(endpoint.Contract.Method, endpoint.Contract.Path, endpoint.RouteHandlerWithRegistry(r.codecs))
	}
}

// ServeHTTP makes the TALK runtime an http.Handler for direct use.
func (r *TalkRuntime) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r == nil {
		http.NotFound(w, req)
		return
	}

	r.mu.RLock()
	router := r.router
	r.mu.RUnlock()
	if router == nil {
		http.NotFound(w, req)
		return
	}

	router.ServeHTTP(w, req)
}

func (r *TalkRuntime) rebuildRouterLocked() {
	router := NewRouter()
	endpoints := make([]TalkEndpoint, 0, len(r.endpoints))
	for _, endpoint := range r.endpoints {
		endpoints = append(endpoints, endpoint)
	}
	sort.SliceStable(endpoints, func(i, j int) bool {
		if endpoints[i].Contract.Path == endpoints[j].Contract.Path {
			return strings.ToUpper(endpoints[i].Contract.Method) < strings.ToUpper(endpoints[j].Contract.Method)
		}
		return endpoints[i].Contract.Path < endpoints[j].Contract.Path
	})
	for _, endpoint := range endpoints {
		router.Handle(endpoint.Contract.Method, endpoint.Contract.Path, endpoint.RouteHandlerWithRegistry(r.codecs))
	}
	r.router = router
}

// RouteHandlerWithRegistry adapts a TALK endpoint with negotiated codecs.
func (e TalkEndpoint) RouteHandlerWithRegistry(registry *TalkCodecRegistry) RouteHandler {
	e = e.Normalize()
	if registry == nil {
		registry = NewTalkCodecRegistry()
	}

	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		if strings.ToUpper(r.Method) != strings.ToUpper(e.Contract.Method) {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
		}

		codec := registry.Negotiate(e.Contract.Profile, r.Header.Get("Accept"), r.Header.Get("Content-Type"))
		if codec == nil {
			codec = TalkJSONCodec{}
		}

		body, err := readTalkBody(r, codec)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		frame := TalkFrame{
			Session:   e.Contract.Name,
			Topic:     e.Contract.Path,
			Kind:      TalkFrameKindData,
			Direction: TalkDirectionInbound,
			Payload:   body,
			Metadata:  cloneStringMap(e.Contract.Metadata),
		}.Normalize()

		if err := e.validatePayload(frame.Payload, e.Contract.RequestSchema); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		response, err := e.Handler(r.Context(), TalkRequest{
			Contract: e.Contract.Name,
			Frame:    frame,
			Params:   cloneStringMap(params),
			Query:    queryValuesToMap(r),
			Headers:  headersToMap(r),
		})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if response.Status == 0 {
			response.Status = http.StatusOK
		}
		for key, value := range response.Headers {
			w.Header().Set(key, value)
		}

		if err := e.validatePayload(response.Frame.Payload, e.Contract.ResponseSchema); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		responsePayload := response.Frame.Payload
		if responsePayload == nil || !isJSONTalkCodec(codec) {
			responsePayload = response.Frame.Normalize()
		}

		writeTalkBody(w, responsePayload, response.Status, codec)
	}
}

func talkEndpointKey(method, path string) string {
	return strings.ToUpper(strings.TrimSpace(method)) + " " + normalizePath(path)
}

func normalizeTalkContentType(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return ""
	}
	if index := strings.Index(value, ";"); index >= 0 {
		value = value[:index]
	}
	return strings.TrimSpace(value)
}

func parseTalkAcceptHeader(acceptHeader string) []string {
	if strings.TrimSpace(acceptHeader) == "" {
		return nil
	}

	parts := strings.Split(acceptHeader, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		contentType := normalizeTalkContentType(part)
		if contentType == "" || contentType == "*/*" {
			continue
		}
		out = append(out, contentType)
	}
	return out
}

func isJSONTalkCodec(codec TalkCodec) bool {
	if codec == nil {
		return true
	}

	contentType := normalizeTalkContentType(codec.ContentType())
	return strings.HasSuffix(contentType, "+json") || contentType == "application/json"
}
