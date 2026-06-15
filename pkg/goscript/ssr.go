package goscript

type SSREngine struct {
	store *Store
}

func NewSSREngine(store *Store) *SSREngine {
	return &SSREngine{store: store}
}

func (ssr *SSREngine) RenderToString(component Component) (string, error) {
	rendered := component.Render()
	var state interface{}
	if ssr != nil && ssr.store != nil {
		state = ssr.store.Snapshot()
	}

	renderedShell, err := RenderHydrationShell(rendered, HydrationPayload{
		AppID:   "app",
		Version: "go-script",
		Title:   "GoScript App",
		RootID:  "app",
		State:   state,
	})
	if err != nil {
		return "", err
	}

	return renderedShell, nil
}

