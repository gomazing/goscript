package goscript

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestTalkSessionPublishSubscribe(t *testing.T) {
	profile := TalkProfile{
		Name:         "sensor-stream",
		Text:         true,
		Data:         true,
		Media:        true,
		Streaming:    true,
		Sensors:      true,
		LiveTracking: true,
		Bidirectional: true,
		MultiThreaded: true,
	}

	session := NewTalkSession("alpha", "/rooms/alpha", profile)
	sub, err := session.Subscribe("ui", 4)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	frame, err := session.Publish(TalkFrame{
		Kind:    TalkFrameKindText,
		Payload: "hello",
	})
	if err != nil {
		t.Fatalf("publish: %v", err)
	}

	if frame.Sequence != 1 {
		t.Fatalf("expected sequence 1, got %d", frame.Sequence)
	}
	if frame.Session != "alpha" {
		t.Fatalf("expected session alpha, got %q", frame.Session)
	}

	select {
	case got := <-sub:
		if got.Payload != "hello" {
			t.Fatalf("unexpected frame payload: %#v", got.Payload)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for frame")
	}

	if history := session.Poll(10); len(history) != 1 {
		t.Fatalf("expected one frame in history, got %d", len(history))
	}

	session.Close()
}

func TestTalkProfileSupportsKinds(t *testing.T) {
	profile := TalkProfile{Text: true, Data: true}.Normalize()
	if !profile.Supports(TalkFrameKindText) {
		t.Fatalf("expected text support")
	}
	if !profile.Supports(TalkFrameKindData) {
		t.Fatalf("expected data support")
	}
	if profile.Supports(TalkFrameKindMedia) {
		t.Fatalf("did not expect media support")
	}
}

func TestTalkJSONCodecRoundTrip(t *testing.T) {
	codec := TalkJSONCodec{}
	encoded, err := codec.Encode(TalkFrame{
		ID:      "frame-1",
		Topic:   "telemetry",
		Kind:    TalkFrameKindSensor,
		Payload: map[string]interface{}{"temp": 27.5},
	})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if decoded.Topic != "telemetry" || decoded.Kind != TalkFrameKindSensor {
		t.Fatalf("unexpected decoded frame: %#v", decoded)
	}
}

func TestTalkEndpointRouteHandler(t *testing.T) {
	endpoint := TalkEndpoint{
		Contract: TalkContract{
			Name:   "echo",
			Path:   "/talk/echo",
			Method: http.MethodPost,
			RequestSchema: FormSchema{
				"message": {
					Name:     "message",
					Type:     "string",
					Required: true,
				},
			},
		},
		Handler: func(ctx context.Context, req TalkRequest) (TalkResponse, error) {
			message, _ := req.Frame.Payload.(map[string]interface{})["message"].(string)
			return TalkResponse{
				Status: http.StatusCreated,
				Frame: TalkFrame{
					Kind:    TalkFrameKindData,
					Payload: map[string]interface{}{"reply": message},
				},
			}, nil
		},
	}

	router := NewRouter()
	router.POST(endpoint.Contract.Path, endpoint.RouteHandler(TalkJSONCodec{}))

	req := httptest.NewRequest(http.MethodPost, "/talk/echo", bytes.NewBufferString(`{"message":"hello"}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	if payload["reply"] != "hello" {
		t.Fatalf("unexpected response payload: %#v", payload)
	}
}

func TestTalkEndpointAcceptsEnvelope(t *testing.T) {
	endpoint := TalkEndpoint{
		Contract: TalkContract{
			Name:   "envelope",
			Path:   "/talk/envelope",
			Method: http.MethodPost,
		},
		Handler: func(ctx context.Context, req TalkRequest) (TalkResponse, error) {
			return TalkResponse{
				Status: http.StatusOK,
				Frame: TalkFrame{
					Kind:    TalkFrameKindData,
					Payload: map[string]interface{}{"reply": req.Frame.Payload},
				},
			}, nil
		},
	}

	router := NewRouter()
	router.POST(endpoint.Contract.Path, endpoint.RouteHandler(TalkJSONCodec{}))

	body, err := json.Marshal(TalkFrame{
		Session: "stream-9",
		Topic:   "/talk/envelope",
		Kind:    TalkFrameKindData,
		Payload: map[string]interface{}{"message": "hello"},
	})
	if err != nil {
		t.Fatalf("marshal envelope: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/talk/envelope", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	reply, ok := payload["reply"].(map[string]interface{})
	if !ok || reply["message"] != "hello" {
		t.Fatalf("unexpected response payload: %#v", payload)
	}
}

func TestTalkBusLifecycle(t *testing.T) {
	bus := NewTalkBus()
	session := bus.OpenSession("stream-1", "/streams/live", TalkProfile{
		Text:      true,
		Data:      true,
		Streaming: true,
	})

	sub, err := bus.Subscribe("stream-1", "observer", 2)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	if _, err := bus.Publish("stream-1", TalkFrame{
		Kind:    TalkFrameKindStream,
		Payload: map[string]interface{}{"chunk": 1},
	}); err != nil {
		t.Fatalf("publish: %v", err)
	}

	select {
	case <-sub:
	case <-time.After(2 * time.Second):
		t.Fatalf("timed out waiting for bus frame")
	}

	snapshots := bus.Sessions()
	if len(snapshots) != 1 || snapshots[0].ID != session.ID {
		t.Fatalf("unexpected snapshots: %#v", snapshots)
	}

	bus.CloseSession("stream-1")
}

func BenchmarkTalkSessionPublish(b *testing.B) {
	session := NewTalkSession("bench", "/rooms/bench", TalkProfile{
		Text:      true,
		Data:      true,
		Streaming: true,
		Bidirectional: true,
	})

	_, _ = session.Subscribe("listener", 32)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := session.Publish(TalkFrame{
			Kind:    TalkFrameKindText,
			Payload: strings.Repeat("x", 16),
		}); err != nil {
			b.Fatal(err)
		}
	}
}
