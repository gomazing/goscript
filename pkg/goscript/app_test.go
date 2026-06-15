package goscript

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAppRenderPage(t *testing.T) {
	app := NewApp("portal", "1.2.3")
	app.DefaultMeta["theme"] = "midnight"
	app.Styles = []string{"body { background: #111; }"}
	app.Scripts = []string{"window.__READY__ = true;"}

	page := Page{
		Path:        "/dashboard",
		Title:       "Dashboard",
		Description: "Operations overview",
		Hydrate:     true,
		Meta:        map[string]string{"page": "dashboard"},
		Component: FunctionalComponent(func(props Props, children ...interface{}) string {
			return CreateElement("div", Props{"class": "hero"}, "Hello GoScript")
		}),
	}

	if err := app.RegisterPage(page); err != nil {
		t.Fatalf("register page: %v", err)
	}

	html, err := app.RenderPage("/dashboard")
	if err != nil {
		t.Fatalf("render page: %v", err)
	}

	for _, want := range []string{
		"<title>Dashboard</title>",
		"meta name=\"description\" content=\"Operations overview\"",
		"data-goscript-hydrate=\"true\"",
		"id=\"dashboard\"",
		"Hello GoScript",
		"window.__READY__ = true;",
		"body { background: #111; }",
		"meta name=\"theme\" content=\"midnight\"",
	} {
		if !strings.Contains(html, want) {
			t.Fatalf("expected rendered page to contain %q, got %s", want, html)
		}
	}
}

func TestAppTalkEndpoint(t *testing.T) {
	app := NewApp("portal", "1.2.3")

	if err := app.RegisterTalkEndpoint(TalkEndpoint{
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
	}); err != nil {
		t.Fatalf("register talk endpoint: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/talk/echo", strings.NewReader("hello"))
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Accept", "text/plain")
	rec := httptest.NewRecorder()

	app.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("unexpected status: %d", rec.Code)
	}

	if got := strings.TrimSpace(rec.Body.String()); got != "HELLO" {
		t.Fatalf("unexpected response body: %q", got)
	}
}
