package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gomazing/goscript/pkg/goscript"
)

func main() {
	app := goscript.NewApp("goscript-demo", "1.0.0")
	app.DefaultMeta["theme"] = "midnight"
	app.Styles = []string{
		`body { margin: 0; font-family: Inter, Arial, sans-serif; background: #0f172a; color: #e2e8f0; }`,
		`.shell { max-width: 960px; margin: 0 auto; padding: 64px 24px; }`,
		`.hero { background: linear-gradient(135deg, #1d4ed8, #8b5cf6); padding: 32px; border-radius: 24px; }`,
		`.card { background: rgba(15, 23, 42, 0.72); border: 1px solid rgba(148, 163, 184, 0.2); border-radius: 20px; padding: 24px; margin-top: 24px; }`,
	}
	app.Scripts = []string{
		`window.addEventListener('DOMContentLoaded', function () { console.log('GoScript demo ready'); });`,
	}

	home := goscript.FunctionalComponent(func(props goscript.Props, children ...interface{}) string {
		return goscript.CreateElement("main", goscript.Props{"class": "shell"},
			goscript.CreateElement("section", goscript.Props{"class": "hero"},
				goscript.CreateElement("p", nil, "GoScript"),
				goscript.CreateElement("h1", nil, "Modern web language built for vibe coding"),
				goscript.CreateElement("p", nil, "The JavaScript for the AI Era, built with Go-native runtime foundations."),
			),
			goscript.CreateElement("section", goscript.Props{"class": "card"},
				goscript.CreateElement("h2", nil, "What you can build"),
				goscript.CreateElement("ul", nil,
					goscript.CreateElement("li", nil, "Simple yet elegant websites with Go Vibe"),
					goscript.CreateElement("li", nil, "Robust ecommerce marketplaces"),
					goscript.CreateElement("li", nil, "Swarm-based multi-server applications"),
					goscript.CreateElement("li", nil, "Massive modular ERPs"),
				),
			),
			goscript.CreateElement("section", goscript.Props{"class": "card"},
				goscript.CreateElement("a", goscript.Props{"href": "/api/hello"}, "Try the API"),
			),
		)
	})

	if err := app.RegisterPage(goscript.Page{
		Path:        "/",
		Title:       "GoScript",
		Description: "Modern web language built for developing websites in vibe coding",
		Component:   home,
		Hydrate:     true,
		Meta: map[string]string{
			"section": "home",
		},
	}); err != nil {
		log.Fatal(err)
	}

	app.GET("/api/hello", helloHandler)

	port := 8080
	fmt.Printf("Server starting on http://localhost:%d\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), app))
}

func helloHandler(w http.ResponseWriter, r *http.Request, params map[string]string) {
	fmt.Fprintf(w, "Hello from GoScript API!")
}

