package goscript

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTalkTextCodecRoundTrip(t *testing.T) {
	codec := TalkTextCodec{}

	encoded, err := codec.Encode(TalkFrame{
		Kind:    TalkFrameKindText,
		Payload: "hello",
	})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	if got, ok := decoded.Payload.(string); !ok || got != "hello" {
		t.Fatalf("unexpected payload: %#v", decoded.Payload)
	}
}

func TestTalkBinaryCodecRoundTrip(t *testing.T) {
	codec := TalkBinaryCodec{}

	encoded, err := codec.Encode(TalkFrame{
		Kind:    TalkFrameKindSensor,
		Payload: []byte("sensor-bytes"),
	})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	decoded, err := codec.Decode(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}

	payload, ok := decoded.Payload.([]byte)
	if !ok {
		t.Fatalf("unexpected payload type: %T", decoded.Payload)
	}
	if !bytes.Equal(payload, []byte("sensor-bytes")) {
		t.Fatalf("unexpected payload: %q", payload)
	}
}

func TestTalkCodecRegistryNegotiation(t *testing.T) {
	registry := NewTalkCodecRegistry()

	textCodec := registry.Negotiate(TalkProfile{Text: true}, "", "text/plain; charset=utf-8")
	if contentType := normalizeTalkContentType(textCodec.ContentType()); contentType != "text/plain" {
		t.Fatalf("expected text codec, got %q", contentType)
	}

	binaryCodec := registry.Negotiate(TalkProfile{Media: true, Sensors: true}, "", "")
	if contentType := normalizeTalkContentType(binaryCodec.ContentType()); contentType != "application/octet-stream" {
		t.Fatalf("expected binary codec, got %q", contentType)
	}

	jsonCodec := registry.Negotiate(TalkProfile{}, "application/json", "")
	if contentType := normalizeTalkContentType(jsonCodec.ContentType()); contentType != "application/json" {
		t.Fatalf("expected json codec, got %q", contentType)
	}
}

func TestTalkRuntimeNegotiatedRoute(t *testing.T) {
	runtime := NewTalkRuntime()
	endpoint := TalkEndpoint{
		Contract: TalkContract{
			Name:   "echo",
			Path:   "/talk/echo",
			Method: http.MethodPost,
			Profile: TalkProfile{
				Text: true,
			},
		},
		Handler: func(ctx context.Context, req TalkRequest) (TalkResponse, error) {
			message, _ := req.Frame.Payload.(string)
			return TalkResponse{
				Status: http.StatusAccepted,
				Frame: TalkFrame{
					Kind:    TalkFrameKindText,
					Payload: strings.ToUpper(message),
				},
			}, nil
		},
	}

	if err := runtime.RegisterEndpoint(endpoint); err != nil {
		t.Fatalf("register endpoint: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/talk/echo", strings.NewReader("hello"))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Accept", "text/plain")
	rec := httptest.NewRecorder()

	runtime.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	if got := strings.TrimSpace(rec.Body.String()); got != "HELLO" {
		t.Fatalf("unexpected response body: %q", got)
	}

	if contentType := rec.Header().Get("Content-Type"); !strings.HasPrefix(contentType, "text/plain") {
		t.Fatalf("expected text/plain content type, got %q", contentType)
	}
}
