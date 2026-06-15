package goscript

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"
)

// TalkFrameKind identifies the payload shape carried by a TALK frame.
type TalkFrameKind string

const (
	TalkFrameKindText    TalkFrameKind = "text"
	TalkFrameKindData    TalkFrameKind = "data"
	TalkFrameKindMedia   TalkFrameKind = "media"
	TalkFrameKindStream  TalkFrameKind = "stream"
	TalkFrameKindSensor  TalkFrameKind = "sensor"
	TalkFrameKindControl TalkFrameKind = "control"
)

// TalkDirection identifies the exchange direction for a frame or contract.
type TalkDirection string

const (
	TalkDirectionInbound       TalkDirection = "inbound"
	TalkDirectionOutbound      TalkDirection = "outbound"
	TalkDirectionBidirectional  TalkDirection = "bidirectional"
)

// TalkProfile describes the transport capabilities a peer or session can support.
type TalkProfile struct {
	Name          string            `json:"name,omitempty"`
	Text          bool              `json:"text,omitempty"`
	Data          bool              `json:"data,omitempty"`
	Media         bool              `json:"media,omitempty"`
	Streaming     bool              `json:"streaming,omitempty"`
	Sensors       bool              `json:"sensors,omitempty"`
	LiveTracking  bool              `json:"liveTracking,omitempty"`
	Bidirectional  bool              `json:"bidirectional,omitempty"`
	MultiThreaded  bool              `json:"multiThreaded,omitempty"`
	MaxFrameBytes  int               `json:"maxFrameBytes,omitempty"`
	MaxStreams     int               `json:"maxStreams,omitempty"`
	Metadata       map[string]string `json:"metadata,omitempty"`
}

// Normalize fills safe defaults for the profile.
func (p TalkProfile) Normalize() TalkProfile {
	if strings.TrimSpace(p.Name) == "" {
		p.Name = "talk"
	}
	if !p.Text && !p.Data && !p.Media && !p.Streaming && !p.Sensors {
		p.Data = true
	}
	if p.Metadata == nil {
		p.Metadata = map[string]string{}
	}
	return p
}

// Supports reports whether the profile can carry the supplied frame kind.
func (p TalkProfile) Supports(kind TalkFrameKind) bool {
	p = p.Normalize()
	switch kind {
	case TalkFrameKindText:
		return p.Text || p.Data
	case TalkFrameKindData:
		return p.Data || p.Text
	case TalkFrameKindMedia:
		return p.Media
	case TalkFrameKindStream:
		return p.Streaming
	case TalkFrameKindSensor:
		return p.Sensors
	case TalkFrameKindControl:
		return true
	default:
		return p.Data
	}
}

// TalkFrame is the universal transport unit for Go TALK compatibility.
type TalkFrame struct {
	ID        string            `json:"id,omitempty"`
	Session   string            `json:"session,omitempty"`
	Topic     string            `json:"topic,omitempty"`
	Stream    string            `json:"stream,omitempty"`
	Kind      TalkFrameKind     `json:"kind,omitempty"`
	Direction TalkDirection     `json:"direction,omitempty"`
	Payload   interface{}       `json:"payload,omitempty"`
	Metadata  map[string]string `json:"metadata,omitempty"`
	Sequence  int64             `json:"sequence,omitempty"`
	Timestamp time.Time         `json:"timestamp,omitempty"`
}

// Normalize fills safe defaults for a frame.
func (f TalkFrame) Normalize() TalkFrame {
	if strings.TrimSpace(f.Kind) == "" {
		f.Kind = TalkFrameKindData
	}
	if strings.TrimSpace(f.Direction) == "" {
		f.Direction = TalkDirectionBidirectional
	}
	if f.Timestamp.IsZero() {
		f.Timestamp = time.Now().UTC()
	}
	if f.Metadata == nil {
		f.Metadata = map[string]string{}
	}
	return f
}

// Validate ensures the frame can be routed and observed.
func (f TalkFrame) Validate() error {
	if strings.TrimSpace(f.Topic) == "" && strings.TrimSpace(f.Session) == "" {
		return fmt.Errorf("frame requires at least a session or topic")
	}
	return nil
}

// TalkCodec encodes and decodes frames for external transports.
type TalkCodec interface {
	ContentType() string
	Encode(TalkFrame) ([]byte, error)
	Decode([]byte) (TalkFrame, error)
}

// TalkJSONCodec is the default compatibility codec.
type TalkJSONCodec struct{}

// ContentType returns the codec MIME type.
func (TalkJSONCodec) ContentType() string { return "application/json" }

// Encode serializes a frame to JSON.
func (TalkJSONCodec) Encode(frame TalkFrame) ([]byte, error) {
	return json.Marshal(frame.Normalize())
}

// Decode deserializes a frame from JSON.
func (TalkJSONCodec) Decode(data []byte) (TalkFrame, error) {
	var frame TalkFrame
	if err := json.Unmarshal(data, &frame); err != nil {
		return TalkFrame{}, err
	}
	return frame.Normalize(), nil
}

// TalkContract describes a compatible service or protocol endpoint.
type TalkContract struct {
	Name          string            `json:"name"`
	Path          string            `json:"path"`
	Method        string            `json:"method,omitempty"`
	Version       string            `json:"version,omitempty"`
	Description   string            `json:"description,omitempty"`
	Profile       TalkProfile       `json:"profile,omitempty"`
	RequestSchema FormSchema        `json:"requestSchema,omitempty"`
	ResponseSchema FormSchema       `json:"responseSchema,omitempty"`
	Metadata      map[string]string `json:"metadata,omitempty"`
}

// Normalize fills default contract values.
func (c TalkContract) Normalize() TalkContract {
	if strings.TrimSpace(c.Name) == "" {
		c.Name = sanitizeModuleName(c.Path)
	}
	if strings.TrimSpace(c.Name) == "" {
		c.Name = "talk"
	}
	c.Path = normalizePath(c.Path)
	if strings.TrimSpace(c.Method) == "" {
		c.Method = http.MethodPost
	}
	if strings.TrimSpace(c.Version) == "" {
		c.Version = "1"
	}
	c.Profile = c.Profile.Normalize()
	if c.Metadata == nil {
		c.Metadata = map[string]string{}
	}
	return c
}

// Validate ensures the contract is structurally usable.
func (c TalkContract) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("contract name is required")
	}
	if strings.TrimSpace(c.Path) == "" {
		return fmt.Errorf("contract path is required")
	}
	if strings.TrimSpace(c.Method) == "" {
		return fmt.Errorf("contract method is required")
	}
	return nil
}

// TalkRequest is the runtime envelope handed to a TALK handler.
type TalkRequest struct {
	Contract string
	Frame    TalkFrame
	Params   map[string]string
	Query    map[string]string
	Headers  map[string]string
}

// TalkResponse is the runtime envelope returned by a TALK handler.
type TalkResponse struct {
	Status  int
	Headers map[string]string
	Frame   TalkFrame
}

// TalkHandler executes a TALK contract.
type TalkHandler func(context.Context, TalkRequest) (TalkResponse, error)

// TalkEndpoint binds a contract to a handler.
type TalkEndpoint struct {
	Contract TalkContract
	Handler  TalkHandler
}

// Normalize fills safe defaults on the endpoint contract.
func (e TalkEndpoint) Normalize() TalkEndpoint {
	e.Contract = e.Contract.Normalize()
	return e
}

// Validate ensures the endpoint can be executed.
func (e TalkEndpoint) Validate() error {
	e = e.Normalize()
	if err := e.Contract.Validate(); err != nil {
		return err
	}
	if e.Handler == nil {
		return fmt.Errorf("talk handler is required")
	}
	return nil
}

// RouteHandler adapts the endpoint into the existing GoScript router contract.
func (e TalkEndpoint) RouteHandler(codec TalkCodec) RouteHandler {
	e = e.Normalize()
	if codec == nil {
		codec = TalkJSONCodec{}
	}

	return func(w http.ResponseWriter, r *http.Request, params map[string]string) {
		if strings.ToUpper(r.Method) != strings.ToUpper(e.Contract.Method) {
			http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
			return
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

		responsePayload := response.Frame.Payload
		if err := e.validatePayload(responsePayload, e.Contract.ResponseSchema); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		payload := responsePayload
		if payload == nil {
			payload = response.Frame
		}

		writeTalkBody(w, payload, response.Status, codec)
	}
}

func (e TalkEndpoint) validatePayload(payload interface{}, schema FormSchema) error {
	if len(schema) == 0 {
		return nil
	}

	values, ok := payload.(map[string]interface{})
	if !ok {
		if payload == nil {
			values = map[string]interface{}{}
		} else {
			return fmt.Errorf("payload must be an object for schema validation")
		}
	}

	if errs := ValidateForm(schema, values); len(errs) > 0 {
		parts := make([]string, 0, len(errs))
		for _, err := range errs {
			parts = append(parts, fmt.Sprintf("%s: %s", err.Field, err.Message))
		}
		sort.Strings(parts)
		return fmt.Errorf(strings.Join(parts, "; "))
	}

	return nil
}

// TalkSession represents a bidirectional runtime channel for one topic or peer.
type TalkSession struct {
	mu         sync.RWMutex
	ID         string
	Topic      string
	Profile    TalkProfile
	subscribers map[string]chan TalkFrame
	history    []TalkFrame
	sequence   int64
	closed     bool
}

// NewTalkSession creates a session with the supplied profile.
func NewTalkSession(id, topic string, profile TalkProfile) *TalkSession {
	id = strings.TrimSpace(id)
	if id == "" {
		id = sanitizeModuleName(topic)
	}
	if id == "" {
		id = "talk"
	}

	return &TalkSession{
		ID:          id,
		Topic:       normalizePath(topic),
		Profile:     profile.Normalize(),
		subscribers: make(map[string]chan TalkFrame),
	}
}

// Publish stores and broadcasts a frame on the session.
func (s *TalkSession) Publish(frame TalkFrame) (TalkFrame, error) {
	if s == nil {
		return TalkFrame{}, fmt.Errorf("talk session is nil")
	}

	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return TalkFrame{}, fmt.Errorf("talk session is closed")
	}

	frame = frame.Normalize()
	if frame.Session == "" {
		frame.Session = s.ID
	}
	if frame.Topic == "" {
		frame.Topic = s.Topic
	}
	if frame.Direction == "" {
		frame.Direction = TalkDirectionBidirectional
	}
	if !s.Profile.Supports(frame.Kind) {
		s.mu.Unlock()
		return TalkFrame{}, fmt.Errorf("profile %q does not support frame kind %q", s.Profile.Name, frame.Kind)
	}

	s.sequence++
	frame.Sequence = s.sequence
	if frame.ID == "" {
		frame.ID = fmt.Sprintf("%s-%d", frame.Session, frame.Sequence)
	}

	s.history = append(s.history, frame)
	subscribers := make([]chan TalkFrame, 0, len(s.subscribers))
	for _, ch := range s.subscribers {
		subscribers = append(subscribers, ch)
	}
	s.mu.Unlock()

	for _, ch := range subscribers {
		safeSendTalkFrame(ch, frame)
	}

	return frame, nil
}

// Subscribe registers a subscriber to the session stream.
func (s *TalkSession) Subscribe(subscriberID string, buffer int) (<-chan TalkFrame, error) {
	if s == nil {
		return nil, fmt.Errorf("talk session is nil")
	}
	if strings.TrimSpace(subscriberID) == "" {
		return nil, fmt.Errorf("subscriber id is required")
	}
	if buffer < 1 {
		buffer = 16
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return nil, fmt.Errorf("talk session is closed")
	}

	if existing, ok := s.subscribers[subscriberID]; ok {
		return existing, nil
	}

	ch := make(chan TalkFrame, buffer)
	s.subscribers[subscriberID] = ch
	return ch, nil
}

// Unsubscribe removes a subscriber and closes its channel.
func (s *TalkSession) Unsubscribe(subscriberID string) {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if ch, ok := s.subscribers[subscriberID]; ok {
		close(ch)
		delete(s.subscribers, subscriberID)
	}
}

// Poll returns a copy of recent frames.
func (s *TalkSession) Poll(limit int) []TalkFrame {
	if s == nil {
		return nil
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	if limit <= 0 || limit >= len(s.history) {
		out := make([]TalkFrame, len(s.history))
		copy(out, s.history)
		return out
	}

	out := make([]TalkFrame, limit)
	copy(out, s.history[len(s.history)-limit:])
	return out
}

// Snapshot returns a lightweight view of the session state.
func (s *TalkSession) Snapshot() TalkSessionSnapshot {
	if s == nil {
		return TalkSessionSnapshot{}
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	return TalkSessionSnapshot{
		ID:          s.ID,
		Topic:       s.Topic,
		Profile:     s.Profile,
		Subscribers: len(s.subscribers),
		Frames:      len(s.history),
		Closed:      s.closed,
	}
}

// Close shuts down the session and all subscribers.
func (s *TalkSession) Close() {
	if s == nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.closed {
		return
	}

	for id, ch := range s.subscribers {
		close(ch)
		delete(s.subscribers, id)
	}
	s.closed = true
}

// TalkSessionSnapshot summarizes the current session state.
type TalkSessionSnapshot struct {
	ID          string
	Topic       string
	Profile     TalkProfile
	Subscribers int
	Frames      int
	Closed      bool
}

// TalkBus manages multiple sessions and their routed frames.
type TalkBus struct {
	mu       sync.RWMutex
	sessions map[string]*TalkSession
}

// NewTalkBus creates a new transport bus.
func NewTalkBus() *TalkBus {
	return &TalkBus{
		sessions: make(map[string]*TalkSession),
	}
}

// OpenSession creates or replaces a session.
func (b *TalkBus) OpenSession(id, topic string, profile TalkProfile) *TalkSession {
	if b == nil {
		return nil
	}

	session := NewTalkSession(id, topic, profile)

	b.mu.Lock()
	defer b.mu.Unlock()
	b.sessions[session.ID] = session
	return session
}

// Session returns a session by id.
func (b *TalkBus) Session(id string) (*TalkSession, bool) {
	if b == nil {
		return nil, false
	}

	b.mu.RLock()
	defer b.mu.RUnlock()

	session, ok := b.sessions[strings.TrimSpace(id)]
	return session, ok
}

// Publish forwards a frame through the named session.
func (b *TalkBus) Publish(sessionID string, frame TalkFrame) (TalkFrame, error) {
	session, ok := b.Session(sessionID)
	if !ok {
		return TalkFrame{}, fmt.Errorf("session %q not found", sessionID)
	}
	return session.Publish(frame)
}

// Subscribe attaches a subscriber to the named session.
func (b *TalkBus) Subscribe(sessionID, subscriberID string, buffer int) (<-chan TalkFrame, error) {
	session, ok := b.Session(sessionID)
	if !ok {
		return nil, fmt.Errorf("session %q not found", sessionID)
	}
	return session.Subscribe(subscriberID, buffer)
}

// Unsubscribe removes a subscriber from the named session.
func (b *TalkBus) Unsubscribe(sessionID, subscriberID string) {
	session, ok := b.Session(sessionID)
	if !ok {
		return
	}
	session.Unsubscribe(subscriberID)
}

// CloseSession closes and removes a session.
func (b *TalkBus) CloseSession(sessionID string) {
	if b == nil {
		return
	}

	b.mu.Lock()
	session, ok := b.sessions[strings.TrimSpace(sessionID)]
	if ok {
		delete(b.sessions, strings.TrimSpace(sessionID))
	}
	b.mu.Unlock()

	if ok {
		session.Close()
	}
}

// Sessions returns snapshots for all known sessions.
func (b *TalkBus) Sessions() []TalkSessionSnapshot {
	if b == nil {
		return nil
	}

	b.mu.RLock()
	sessions := make([]*TalkSession, 0, len(b.sessions))
	for _, session := range b.sessions {
		sessions = append(sessions, session)
	}
	b.mu.RUnlock()

	out := make([]TalkSessionSnapshot, 0, len(sessions))
	for _, session := range sessions {
		out = append(out, session.Snapshot())
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].ID < out[j].ID
	})
	return out
}

func readTalkBody(r *http.Request, codec TalkCodec) (interface{}, error) {
	if r == nil || r.Body == nil {
		return nil, nil
	}

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}

	if codec != nil && !isJSONTalkCodec(codec) {
		if frame, err := codec.Decode(data); err == nil {
			if frame.Payload != nil {
				return frame.Payload, nil
			}
			return frame, nil
		}
	}

	var payload interface{}
	if err := json.Unmarshal(data, &payload); err != nil {
		return string(data), nil
	}

	if envelope, ok := payload.(map[string]interface{}); ok && looksLikeTalkEnvelope(envelope) {
		if codec != nil {
			if frame, err := codec.Decode(data); err == nil {
				if frame.Payload != nil {
					return frame.Payload, nil
				}
				return frame, nil
			}
		}

		var frame TalkFrame
		if err := json.Unmarshal(data, &frame); err == nil {
			frame = frame.Normalize()
			if frame.Payload != nil {
				return frame.Payload, nil
			}
			return frame, nil
		}
	}

	return payload, nil
}

func looksLikeTalkEnvelope(payload map[string]interface{}) bool {
	if len(payload) == 0 {
		return false
	}

	_, hasPayload := payload["payload"]
	if !hasPayload {
		return false
	}

	for _, key := range []string{"kind", "session", "topic", "stream", "direction", "sequence"} {
		if _, ok := payload[key]; ok {
			return true
		}
	}

	return false
}

func queryValuesToMap(r *http.Request) map[string]string {
	if r == nil || r.URL == nil {
		return map[string]string{}
	}

	values := r.URL.Query()
	out := make(map[string]string, len(values))
	for key, list := range values {
		if len(list) == 0 {
			continue
		}
		out[key] = list[len(list)-1]
	}
	return out
}

func headersToMap(r *http.Request) map[string]string {
	if r == nil {
		return map[string]string{}
	}

	out := make(map[string]string, len(r.Header))
	for key, list := range r.Header {
		if len(list) == 0 {
			continue
		}
		out[key] = list[len(list)-1]
	}
	return out
}

func writeTalkBody(w http.ResponseWriter, payload interface{}, status int, codec TalkCodec) {
	if w == nil {
		return
	}
	if status <= 0 {
		status = http.StatusOK
	}

	switch v := payload.(type) {
	case nil:
		w.WriteHeader(status)
		return
	case []byte:
		w.WriteHeader(status)
		_, _ = w.Write(v)
		return
	case string:
		w.WriteHeader(status)
		_, _ = w.Write([]byte(v))
		return
	}

	if codec != nil {
		if frame, ok := payload.(TalkFrame); ok {
			if data, err := codec.Encode(frame); err == nil {
				w.Header().Set("Content-Type", codec.ContentType())
				w.WriteHeader(status)
				_, _ = w.Write(data)
				return
			}
		}
	}

	data, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

func safeSendTalkFrame(ch chan TalkFrame, frame TalkFrame) (ok bool) {
	defer func() {
		if recover() != nil {
			ok = false
		}
	}()

	select {
	case ch <- frame:
		return true
	default:
		return false
	}
}
