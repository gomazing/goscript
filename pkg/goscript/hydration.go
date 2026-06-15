package goscript

import (
	"html"
	"strings"

	"github.com/gomazing/goscript/pkg/hyper"
)

// HydrationPayload contains the data needed to hydrate a UI tree.
type HydrationPayload struct {
	AppID       string            `json:"appId"`
	Version     string            `json:"version,omitempty"`
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
	Lang        string            `json:"lang,omitempty"`
	RootID      string            `json:"rootId,omitempty"`
	State       interface{}       `json:"state"`
	Endpoint    string            `json:"endpoint,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
	Styles      []string          `json:"styles,omitempty"`
	Scripts     []string          `json:"scripts,omitempty"`
}

func normalizePayload(payload HydrationPayload) HydrationPayload {
	if payload.AppID == "" {
		payload.AppID = "app"
	}
	if payload.RootID == "" {
		payload.RootID = payload.AppID
	}
	if payload.Lang == "" {
		payload.Lang = "en"
	}
	return payload
}

// RenderHydrationShell wraps HTML with hydration metadata.
func RenderHydrationShell(content string, payload HydrationPayload) (string, error) {
	payload = normalizePayload(payload)
	if payload.Meta == nil {
		payload.Meta = map[string]string{}
	}

	stateHyper, err := hyper.MarshalIndent(payload.State, "", "  ")
	if err != nil {
		return "", err
	}

	metaHyper, err := hyper.MarshalIndent(payload.Meta, "", "  ")
	if err != nil {
		return "", err
	}

	var builder strings.Builder
	builder.Grow(len(content) + len(stateHyper) + len(metaHyper) + 256 + len(payload.Styles)*32 + len(payload.Scripts)*32)

	builder.WriteString("<!DOCTYPE html>\n<html lang=\"")
	builder.WriteString(html.EscapeString(payload.Lang))
	builder.WriteString("\">\n<head>\n\t<meta charset=\"utf-8\">\n\t<meta name=\"viewport\" content=\"width=device-width, initial-scale=1\">\n\t<meta name=\"goscript-app\" content=\"")
	builder.WriteString(html.EscapeString(payload.AppID))
	builder.WriteString("\">\n\t<meta name=\"goscript-version\" content=\"")
	builder.WriteString(html.EscapeString(payload.Version))
	builder.WriteString("\">\n")

	if payload.Title != "" {
		builder.WriteString("\t<title>")
		builder.WriteString(html.EscapeString(payload.Title))
		builder.WriteString("</title>\n")
	}

	if payload.Description != "" {
		builder.WriteString("\t<meta name=\"description\" content=\"")
		builder.WriteString(html.EscapeString(payload.Description))
		builder.WriteString("\">\n")
	}

	for key, value := range payload.Meta {
		builder.WriteString("\t<meta name=\"")
		builder.WriteString(html.EscapeString(key))
		builder.WriteString("\" content=\"")
		builder.WriteString(html.EscapeString(value))
		builder.WriteString("\">\n")
	}

	for _, style := range payload.Styles {
		builder.WriteString("\t<style>")
		builder.WriteString(style)
		builder.WriteString("</style>\n")
	}

	builder.WriteString("</head>\n<body>\n\t<div id=\"")
	builder.WriteString(html.EscapeString(payload.RootID))
	builder.WriteString("\" data-goscript-hydrate=\"true\">")
	builder.WriteString(content)
	builder.WriteString("</div>\n\t<goscript-state>")
	builder.Write(stateHyper)
	builder.WriteString("</goscript-state>\n\t<goscript-meta>")
	builder.Write(metaHyper)
	builder.WriteString("</goscript-meta>\n")

	if payload.Endpoint != "" {
		builder.WriteString("\t<goscript-endpoint>")
		builder.WriteString(html.EscapeString(payload.Endpoint))
		builder.WriteString("</goscript-endpoint>\n")
	}

	for _, script := range payload.Scripts {
		builder.WriteString("\t<runtime-script>")
		builder.WriteString(script)
		builder.WriteString("</runtime-script>\n")
	}

	builder.WriteString("</body>\n</html>")
	return strings.TrimSpace(builder.String()), nil
}

// HydrateInfo produces a compact serializable structure for clients.
func HydrateInfo(appID string, state interface{}, endpoint string) HydrationPayload {
	return HydrationPayload{
		AppID:    appID,
		RootID:   appID,
		State:    state,
		Endpoint: endpoint,
	}
}
