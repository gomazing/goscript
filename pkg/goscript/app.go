package goscript

import (
	"fmt"
	"net/http"
	pathpkg "path"
	"sort"
	"strings"
	"sync"
)

// Page describes a renderable page in a GoScript app.
type Page struct {
	Path        string            `json:"path"`
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
	Styles      []string          `json:"styles,omitempty"`
	Scripts     []string          `json:"scripts,omitempty"`
	RootID      string            `json:"rootId,omitempty"`
	Lang        string            `json:"lang,omitempty"`
	Version     string            `json:"version,omitempty"`
	Endpoint    string            `json:"endpoint,omitempty"`
	Hydrate     bool              `json:"hydrate,omitempty"`
	Component   Component         `json:"-"`
	Layout      func(string) string `json:"-"`
}

// Normalize fills defaults for page metadata.
func (p *Page) Normalize() {
	p.Path = normalizePath(p.Path)
	if p.RootID == "" {
		root := strings.TrimPrefix(p.Path, "/")
		root = strings.ReplaceAll(root, "/", "-")
		if root == "" {
			root = "app"
		}
		p.RootID = root
	}
	if p.Lang == "" {
		p.Lang = "en"
	}
}

// PageManifest summarizes a page for tooling and build exports.
type PageManifest struct {
	Path        string            `json:"path"`
	Title       string            `json:"title,omitempty"`
	Description string            `json:"description,omitempty"`
	RootID      string            `json:"rootId,omitempty"`
	Hydrate     bool              `json:"hydrate,omitempty"`
	Lang        string            `json:"lang,omitempty"`
	Endpoint    string            `json:"endpoint,omitempty"`
	Meta        map[string]string `json:"meta,omitempty"`
}

// FrontendManifest describes the page tree and runtime settings of an app.
type FrontendManifest struct {
	Name              string         `json:"name"`
	Version           string         `json:"version,omitempty"`
	DefaultLang       string         `json:"defaultLang,omitempty"`
	StaticPrefix      string         `json:"staticPrefix,omitempty"`
	HydrationEndpoint string         `json:"hydrationEndpoint,omitempty"`
	Pages             []PageManifest `json:"pages,omitempty"`
}

// App is the top-level frontend runtime for a GoScript website.
type App struct {
	Name              string
	Version           string
	DefaultLang       string
	DefaultMeta       map[string]string
	Styles            []string
	Scripts           []string
	StaticPrefix      string
	HydrationEndpoint string

	Router *Router
	Assets *AssetManager
	Store  *Store
	Talk   *TalkRuntime

	mu    sync.RWMutex
	pages map[string]Page
}

// NewApp creates a new frontend app with sane defaults.
func NewApp(name, version string) *App {
	if strings.TrimSpace(name) == "" {
		name = "app"
	}

	return &App{
		Name:              name,
		Version:           version,
		DefaultLang:       "en",
		DefaultMeta:       map[string]string{},
		StaticPrefix:      "/static/",
		HydrationEndpoint: "/api/hydrate",
		Router:            NewRouter(),
		Store:             NewStore(),
		Talk:              NewTalkRuntime(),
		pages:             map[string]Page{},
	}
}

// Use registers middleware with the internal router.
func (a *App) Use(middleware func(http.HandlerFunc) http.HandlerFunc) {
	if a == nil {
		return
	}
	if a.Router == nil {
		a.Router = NewRouter()
	}
	a.Router.Use(middleware)
}

// Handle registers a plain HTTP route with the internal router.
func (a *App) Handle(method, path string, handler RouteHandler) {
	if a == nil {
		return
	}
	if a.Router == nil {
		a.Router = NewRouter()
	}
	a.Router.Handle(method, path, handler)
}

// GET registers a GET route with the internal router.
func (a *App) GET(path string, handler RouteHandler) {
	a.Handle("GET", path, handler)
}

// POST registers a POST route with the internal router.
func (a *App) POST(path string, handler RouteHandler) {
	a.Handle("POST", path, handler)
}

// UseTalkCodec registers a TALK codec with the app runtime.
func (a *App) UseTalkCodec(codec TalkCodec) {
	if a == nil {
		return
	}
	if a.Talk == nil {
		a.Talk = NewTalkRuntime()
	}
	a.Talk.RegisterCodec(codec)
}

// RegisterTalkEndpoint registers a TALK endpoint and mounts it on the app router.
func (a *App) RegisterTalkEndpoint(endpoint TalkEndpoint) error {
	if a == nil {
		return fmt.Errorf("app is nil")
	}
	if a.Talk == nil {
		a.Talk = NewTalkRuntime()
	}
	if a.Router == nil {
		a.Router = NewRouter()
	}

	endpoint = endpoint.Normalize()
	if err := a.Talk.RegisterEndpoint(endpoint); err != nil {
		return err
	}

	a.Router.Handle(endpoint.Contract.Method, endpoint.Contract.Path, endpoint.RouteHandlerWithRegistry(a.Talk.codecs))
	return nil
}

// TalkBus returns the realtime bus backing Go TALK.
func (a *App) TalkBus() *TalkBus {
	if a == nil || a.Talk == nil {
		return nil
	}
	return a.Talk.Bus()
}

// RegisterPage stores a renderable page in the app.
func (a *App) RegisterPage(page Page) error {
	if a == nil {
		return fmt.Errorf("app is nil")
	}
	page.Normalize()
	if page.Component == nil {
		return fmt.Errorf("page component is required")
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	if a.pages == nil {
		a.pages = map[string]Page{}
	}
	a.pages[page.Path] = page
	return nil
}

// Page returns a page by path.
func (a *App) Page(path string) (Page, bool) {
	if a == nil {
		return Page{}, false
	}
	path = normalizePath(path)

	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.pages == nil {
		return Page{}, false
	}

	page, ok := a.pages[path]
	return page, ok
}

// Pages returns the registered pages in path order.
func (a *App) Pages() []Page {
	if a == nil {
		return nil
	}
	a.mu.RLock()
	defer a.mu.RUnlock()

	if a.pages == nil {
		return nil
	}

	out := make([]Page, 0, len(a.pages))
	for _, page := range a.pages {
		out = append(out, page)
	}
	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Path < out[j].Path
	})
	return out
}

// RenderPage renders a single registered page.
func (a *App) RenderPage(path string) (string, error) {
	if a == nil {
		return "", fmt.Errorf("app is nil")
	}
	page, ok := a.Page(path)
	if !ok {
		return "", fmt.Errorf("page not found: %s", path)
	}
	return page.Render(a)
}

// FrontendManifest returns a serializable description of the frontend runtime.
func (a *App) FrontendManifest() FrontendManifest {
	if a == nil {
		return FrontendManifest{}
	}
	manifest := FrontendManifest{
		Name:              a.Name,
		Version:           a.Version,
		DefaultLang:       a.DefaultLang,
		StaticPrefix:      a.StaticPrefix,
		HydrationEndpoint: a.HydrationEndpoint,
		Pages:             make([]PageManifest, 0),
	}

	for _, page := range a.Pages() {
		manifest.Pages = append(manifest.Pages, PageManifest{
			Path:        page.Path,
			Title:       page.Title,
			Description: page.Description,
			RootID:      page.RootID,
			Hydrate:     page.Hydrate,
			Lang:        page.Lang,
			Endpoint:    page.Endpoint,
			Meta:        cloneStringMap(page.Meta),
		})
	}

	return manifest
}

// ServeHTTP renders pages, static assets, and router-backed endpoints.
func (a *App) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if a == nil {
		http.NotFound(w, r)
		return
	}
	if a.Assets != nil && a.StaticPrefix != "" && strings.HasPrefix(r.URL.Path, a.StaticPrefix) {
		a.Assets.ServeAssets(strings.TrimSuffix(a.StaticPrefix, "/"))(w, r)
		return
	}

	if (r.Method == http.MethodGet || r.Method == http.MethodHead) && a.pages != nil {
		if page, ok := a.Page(r.URL.Path); ok {
			rendered, err := page.Render(a)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(rendered))
			return
		}
	}

	if a.Router != nil {
		a.Router.ServeHTTP(w, r)
		return
	}

	http.NotFound(w, r)
}

// Render builds a page into a hydrated HTML document.
func (p Page) Render(app *App) (string, error) {
	if p.Component == nil {
		return "", fmt.Errorf("page component is required")
	}

	appName := "app"
	appVersion := ""
	defaultLang := "en"
	defaultEndpoint := "/api/hydrate"
	defaultMeta := map[string]string{}
	var appState map[string]interface{}
	var appStyles []string
	var appScripts []string

	if app != nil {
		if strings.TrimSpace(app.Name) != "" {
			appName = app.Name
		}
		appVersion = app.Version
		if strings.TrimSpace(app.DefaultLang) != "" {
			defaultLang = app.DefaultLang
		}
		if strings.TrimSpace(app.HydrationEndpoint) != "" {
			defaultEndpoint = app.HydrationEndpoint
		}
		defaultMeta = app.DefaultMeta
		appStyles = app.Styles
		appScripts = app.Scripts
		if app.Store != nil {
			appState = app.Store.Snapshot()
		}
	}

	body := p.Component.Render()
	if p.Layout != nil {
		body = p.Layout(body)
	}

	payload := HydrationPayload{
		AppID:       appName,
		Version:     coalesceString(p.Version, appVersion),
		Title:       p.Title,
		Description: p.Description,
		Lang:        coalesceString(p.Lang, defaultLang),
		RootID:      p.RootID,
		State:       appState,
		Endpoint:    coalesceString(p.Endpoint, defaultEndpoint),
		Meta:        mergeStringMaps(defaultMeta, p.Meta),
		Styles:      append(append([]string{}, appStyles...), p.Styles...),
		Scripts:     append(append([]string{}, appScripts...), p.Scripts...),
	}

	if !p.Hydrate {
		return body, nil
	}

	return RenderHydrationShell(body, payload)
}

func normalizePath(rawPath string) string {
	rawPath = strings.TrimSpace(rawPath)
	if rawPath == "" {
		return "/"
	}
	if !strings.HasPrefix(rawPath, "/") {
		rawPath = "/" + rawPath
	}
	rawPath = pathpkg.Clean(rawPath)
	if rawPath == "." {
		return "/"
	}
	return rawPath
}

func mergeStringMaps(a, b map[string]string) map[string]string {
	if len(a) == 0 && len(b) == 0 {
		return map[string]string{}
	}

	out := make(map[string]string, len(a)+len(b))
	for key, value := range a {
		out[key] = value
	}
	for key, value := range b {
		out[key] = value
	}
	return out
}

func cloneStringMap(src map[string]string) map[string]string {
	if len(src) == 0 {
		return map[string]string{}
	}

	out := make(map[string]string, len(src))
	for key, value := range src {
		out[key] = value
	}
	return out
}

func coalesceString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}
