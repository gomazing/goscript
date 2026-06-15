package goscript

import (
	"net/http"
	"strings"
	"sync"
)

type RouteHandler func(w http.ResponseWriter, r *http.Request, params map[string]string)

type Route struct {
	Method  string
	Path    string
	Handler RouteHandler
}

type routeSegment struct {
	value   string
	param   bool
	paramID string
}

type compiledRoute struct {
	Route
	segments []routeSegment
}

type Router struct {
	mu          sync.RWMutex
	exactRoutes map[string]map[string]Route
	paramRoutes []compiledRoute
	middleware  []func(http.HandlerFunc) http.HandlerFunc
}

func NewRouter() *Router {
	return &Router{
		exactRoutes: make(map[string]map[string]Route),
	}
}

func (r *Router) Use(middleware func(http.HandlerFunc) http.HandlerFunc) {
	if r == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.middleware = append(r.middleware, middleware)
}

func (r *Router) Handle(method, path string, handler RouteHandler) {
	if r == nil {
		return
	}

	route := Route{Method: method, Path: path, Handler: handler}

	r.mu.Lock()
	defer r.mu.Unlock()

	if strings.Contains(path, ":") {
		r.paramRoutes = append(r.paramRoutes, compiledRoute{
			Route:    route,
			segments: compileRouteSegments(path),
		})
		return
	}

	if r.exactRoutes == nil {
		r.exactRoutes = make(map[string]map[string]Route)
	}

	methodRoutes, ok := r.exactRoutes[path]
	if !ok {
		methodRoutes = make(map[string]Route)
		r.exactRoutes[path] = methodRoutes
	}
	methodRoutes[method] = route
}

func (r *Router) GET(path string, handler RouteHandler) {
	r.Handle("GET", path, handler)
}

func (r *Router) POST(path string, handler RouteHandler) {
	r.Handle("POST", path, handler)
}

func (r *Router) PUT(path string, handler RouteHandler) {
	r.Handle("PUT", path, handler)
}

func (r *Router) DELETE(path string, handler RouteHandler) {
	r.Handle("DELETE", path, handler)
}

func (r *Router) HEAD(path string, handler RouteHandler) {
	r.Handle("HEAD", path, handler)
}

func (r *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if r == nil {
		http.NotFound(w, req)
		return
	}

	r.mu.RLock()
	middleware := append([]func(http.HandlerFunc) http.HandlerFunc(nil), r.middleware...)
	exactRoutes := r.exactRoutes
	paramRoutes := append([]compiledRoute(nil), r.paramRoutes...)
	r.mu.RUnlock()

	requestParts := splitPathSegments(req.URL.Path)
	pathMatched := false
	if methods, ok := exactRoutes[req.URL.Path]; ok {
		pathMatched = true
		for _, method := range requestMethodCandidates(req.Method) {
			if route, ok := methods[method]; ok {
				handler := func(w http.ResponseWriter, r *http.Request) {
					route.Handler(w, r, map[string]string{})
				}

				for i := len(middleware) - 1; i >= 0; i-- {
					handler = middleware[i](handler)
				}

				handler(w, req)
				return
			}
		}
	}

	for _, route := range paramRoutes {
		methodMatched := route.Method == req.Method || (req.Method == http.MethodHead && route.Method == http.MethodGet)
		if methodMatched {
			params, ok := matchCompiledRoute(route.segments, requestParts)
			if ok {
				handler := func(w http.ResponseWriter, r *http.Request) {
					route.Handler(w, r, params)
				}

				// Apply middleware
				for i := len(middleware) - 1; i >= 0; i-- {
					handler = middleware[i](handler)
				}

				handler(w, req)
				return
			}
		}

		if _, ok := matchCompiledRoute(route.segments, requestParts); ok {
			pathMatched = true
		}
	}

	if pathMatched {
		http.Error(w, http.StatusText(http.StatusMethodNotAllowed), http.StatusMethodNotAllowed)
		return
	}

	http.NotFound(w, req)
}

func compileRouteSegments(routePath string) []routeSegment {
	parts := splitPathSegments(routePath)
	if len(parts) == 0 {
		return nil
	}

	segments := make([]routeSegment, 0, len(parts))
	for _, part := range parts {
		if strings.HasPrefix(part, ":") {
			segments = append(segments, routeSegment{
				param:   true,
				paramID: strings.TrimSpace(part[1:]),
			})
			continue
		}

		segments = append(segments, routeSegment{
			value: part,
		})
	}
	return segments
}

func matchCompiledRoute(segments []routeSegment, requestParts []string) (map[string]string, bool) {
	if len(segments) != len(requestParts) {
		return nil, false
	}

	params := make(map[string]string)
	for i, segment := range segments {
		if segment.param {
			params[segment.paramID] = requestParts[i]
			continue
		}
		if segment.value != requestParts[i] {
			return nil, false
		}
	}

	return params, true
}

func splitPathSegments(path string) []string {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "/")
}

func requestMethodCandidates(method string) []string {
	if method == http.MethodHead {
		return []string{http.MethodHead, http.MethodGet}
	}
	return []string{method}
}

