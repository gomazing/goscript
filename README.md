# GoScript

Modern web language built for developing websites in vibe coding.

## The Legacy Web Stack for the AI Era

GoScript is a Go-native web language and application platform for building websites, dashboards, ecommerce, swarm systems, and massive modular ERP software without handing the language layer to the legacy browser stack.

It is multi-threaded by nature, performance-first, and increasingly batteries included: styling, UI primitives, motion, routing, API and database tooling, build-out exports, and AI-guided project structure.

GoScript is not a replacement for Next.js.

GoScript is a replacement for the legacy browser scripting layer.

GoScript is also being shaped as an agentic-first language: intelligent systems should be able to build, inspect, and extend GoScript programs as naturally as humans.

Its longer-term architecture target is broader than today's digital-only languages. GoScript is being designed with a path toward ternary logic, multi-valued logic (MVL), and future quantum-capable execution models, while staying backward compatible with conventional digital runtimes.

[![GitHub stars](https://img.shields.io/github/stars/gomazing/goscript?style=social)](https://github.com/gomazing/goscript/stargazers)
[![GitHub license](https://img.shields.io/github/license/gomazing/goscript)](https://github.com/gomazing/goscript/blob/main/LICENSE)
[![Go Report Card](https://goreportcard.com/badge/github.com/gomazing/goscript)](https://goreportcard.com/report/github.com/gomazing/goscript)
[![GoDoc](https://godoc.org/github.com/gomazing/goscript?status.svg)](https://godoc.org/github.com/gomazing/goscript)

## Why GoScript

- **Built for vibe coding**: develop modern websites and products with a Go-native language instead of stitching Go and the legacy browser stack together.
- **Multi-threaded by nature**: GoScript inherits Go's concurrency-first mindset instead of growing out of a single-threaded web language.
- **Performance first**: built for unparalleled performance with Go-native concurrency, strong server-side execution, and a path toward serious web and edge workloads.
- **All batteries included**: styling, components, routing, API tooling, database tooling, motion primitives, build-out exports, and AI guidance layers.
- **AI-era friendly**: `base/` and `agents/` let teams guide AI coders and build agentic runtime systems in a uniform way. GoScript is meant to be loved by autonomous builders, not only understood by humans.
- **Built for future architectures**: use `cs` for clean client-server apps and `sw` for modular distributed server architecture with Swarm.

## Old Way vs Future Way

GoScript is built to replace the stitched-together legacy web stack with native building blocks for the AI era.

[Read the feature-by-feature guide](./docs/FUTURE_WAY_VS_OLD_WAY.md)
[Open the proposal page for other AIs](./docs/proposed-updates.html)

## Core Protocols

- **Agentic-first**: AI systems are the primary builders and consumers of GoScript, with humans supported second.
- **Machine-legible**: favor explicit semantics, clear contracts, and deterministic behavior that intelligent systems can reason about safely.
- **Architecture-forward**: keep the language ready for digital, ternary, MVL, and future quantum-capable targets.
- **Backward compatible**: digital execution remains the baseline path so GoScript can ship practical software today.

## What You Can Build

1. Simple yet elegant websites using **Go Vibe**
2. Robust ecommerce marketplaces with frontend and middleware
3. Serverless and multi-server applications using **Swarm**
4. Massive modular ERP systems

## Why You Are Going To Love BO

`BO` means **Build Out**.

It lets you export a selected module, tool, or application slice into a deployable artifact without treating the whole product like one indivisible monolith.

- Export a tool as `exe`
- Package a portable `goe`
- Prepare app outputs for `apk`, `ipa`, and `dmg`
- Build from packs so AI agents and humans can inspect what is being shipped
- Export inspectable slices so modular apps can ship only the pieces they need
- Turn one ERP into many focused binaries when needed

[Learn more about BO](./README_BO.md)

## The Stack

- **Go FAST**: compiler/runtime performance, lower-allocation rendering, routing, SSR, and hot-path efficiency
- **Go PAINT**: canvas-first spatial UI, 2D/3D composition, pixel plotting, and hybrid surfaces
- **Go IRT (In Real Time)**: realtime event hubs, streaming sync, scheduler-driven background work, and Swarm communication
- **Go Jetpack**: profiling, observability, benchmarks, diagnostics, and verification workflows
- **GoScript language layer**: components, pages, routing, SSR, hydration, state, hooks, and language/runtime primitives
- **Go Vibe**: motion foundations inspired by Motion / Framer Motion, expressed in a Go-native way
- **Gocsx**: utility-first styling for Go-native UI work
- **GoScale**: API, database, and edge-oriented service foundations
- **Swarm**: modular distributed server architecture for multi-node apps
- **GOPM**: setup, tooling, package workflows, project manifests, lockfiles, and manifest-aware project scaffolding
- **Go Jetpack**: performance and observability tooling
- **Workspace foundations**: early models for docs, inline tables, file browsing, and Notion-style editing surfaces
- **AI guidance layers**: `base/` for build-time instructions and `agents/` for runtime autonomous workers

[Learn more about GOPM](./README_GOPM.md)
[Learn more about Gocsx](./README_GOCSX.md)
[Learn more about GoScale](./README_GOSCALE.md)
[Learn more about Go Jetpack](./README_JETPACK.md)
[Learn more about Vibe](./README_VIBE.md)
[Roadmap pillars](./docs/ROADMAP_PILLARS.md)
[Future use cases](./docs/FUTURE_USE_CASES.md)
[Advanced capabilities checklist](./docs/ADVANCED_CAPABILITIES.md)
[AI adoption checklist](./docs/AI_ADOPTION_CHECKLIST.md)
[Base guidance](./base/README.md)
[Runtime agents](./agents/README.md)

## Try It Now

Well, try now!

Ask Claude to build a website, admin app, dashboards, mobile app, or modular ERP using this repo. The new `goscript` CLI also gives you `fmt`, `check`, `index`, and `watch` workflows for day-to-day development.

Give your feedback.

## Developer CLI

The `goscript` command now gives you a few language-native workflows:

- `go install github.com/gomazing/goscript/cmd/goscript@latest`
- `goscript fmt [path ...]` to format source and markup files
- `goscript check [path ...]` to scan for diagnostics and stub markers
- `goscript index [path ...]` to build a lightweight symbol index
- `goscript watch [path ...]` to poll for file changes and drive hot-reload workflows
- `goscript add <component-name>` to scaffold a new component starter

This is the first layer of the DX pass for the language itself, not just the package manager.

## Quick Start

### Installation

```bash
# Install GOPM
go install github.com/gomazing/goscript/cmd/gopm@latest

# Create a website project
gopm setup --cs --type website my-site
cd my-site
gopm get
```

### Quick Start: Website

```bash
# Scaffold a website in client-server mode
gopm setup --cs --type website my-app
cd my-app

# Start the development workflow
gopm run dev
```

### Quick Start: Swarm ERP

```bash
# Scaffold a modular ERP in swarm mode
gopm setup --sw --type erp my-erp
cd my-erp

# Review the generated pack
cat packs/my-erp.pack
```

### Creating Web Components

#### Class-based Component

```go
type MyComponent struct {
    goscript.LifecycleComponentBase
}

func NewMyComponent(props goscript.Props) *MyComponent {
    base := goscript.NewBaseComponent(props, nil)
    component := &MyComponent{}
    component.LifecycleComponentBase.BaseComponent = *base
    return component
}

func (c *MyComponent) Render() string {
    return goscript.CreateElement("div", nil, 
        goscript.CreateElement("h1", nil, "Hello, World!"),
    )
}
```

#### Functional Component

```go
func MyComponent(props goscript.Props) string {
    return goscript.CreateElement("div", nil, 
        goscript.CreateElement("h1", nil, "Hello, World!"),
    )
}
```

### Creating a 2D Canvas Application

```go
package main

import (
        "log"
        "net/http"

        "github.com/gomazing/goscript/pkg/gocsx"
        "github.com/gomazing/goscript/pkg/gocsx/engine"
)

func main() {
        // Create a new Gocsx instance
        g := gocsx.New()

        // Create a new engine with 2D context
        e := engine.NewEngine(&engine.EngineConfig{
                Context: engine.Context2D,
        })

        // Create a new Canvas2D
        canvas := engine.NewCanvas2D("main-canvas", 800, 600, e)

        // Set render callback
        canvas.SetRenderCallback(func(ctx *engine.Canvas2DContext, deltaTime float64) {
                // Clear the canvas
                ctx.ClearRect(0, 0, 800, 600)

                // Draw a rectangle
                ctx.FillStyle = "#ff0000"
                ctx.FillRect(100, 100, 200, 150)
        })

        // Start the engine
        e.Start()

        // Start the server
        log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Creating a 3D WebGPU Application

```go
package main

import (
        "log"
        "net/http"

        "github.com/gomazing/goscript/pkg/gocsx"
        "github.com/gomazing/goscript/pkg/gocsx/engine"
)

func main() {
        // Create a new Gocsx instance
        g := gocsx.New()

        // Create a new engine with 3D context
        e := engine.NewEngine(&engine.EngineConfig{
                Context: engine.Context3D,
        })

        // Create a new WebGPU instance
        webgpu := engine.NewWebGPU()

        // Create a new Three.js scene
        scene := engine.NewThreeJSScene(e, webgpu)

        // Create a camera
        scene.CreateCamera("main-camera", "Main Camera", [3]float64{0, 0, 5}, [3]float64{0, 0, 0})

        // Create a cube
        scene.CreateCube("cube1", "Cube 1", [3]float64{0, 0, 0}, 1.0, [3]float64{1, 0, 0})

        // Start the engine
        e.Start()

        // Start the server
        log.Fatal(http.ListenAndServe(":8080", nil))
}
```

### Using GoScale API and Database

```go
package main

import (
        "log"

        "github.com/gomazing/goscript/pkg/goscale/api"
        "github.com/gomazing/goscript/pkg/goscale/db"
)

func main() {
        // Initialize database
        database, err := db.NewGoScaleDB(&db.Config{
                ConnectionString: "postgres://user:password@localhost:5432/mydb",
                TimeSeriesEnabled: true,
        })
        if err != nil {
                log.Fatalf("Failed to initialize database: %v", err)
        }

        // Define schema
        schema := api.NewSchema()
        schema.AddType("User", map[string]string{
                "id":    "ID!",
                "name":  "String!",
                "email": "String!",
                "posts": "[Post]",
        })
        schema.AddType("Post", map[string]string{
                "id":      "ID!",
                "title":   "String!",
                "content": "String!",
                "author":  "User!",
        })

        // Initialize API
        apiServer := api.NewGoScaleAPI(&api.Config{
                Schema:  schema,
                DB:      database,
                Port:    8080,
                EdgeEnabled: true,
        })

        // Start API server
        log.Fatal(apiServer.Start())
}
```

### Using Jetpack Performance Monitoring

```go
package main

import (
        "log"
        "net/http"

        "github.com/gomazing/goscript/pkg/jetpack/core"
        "github.com/gomazing/goscript/pkg/jetpack/frontend"
)

func main() {
        // Initialize Jetpack
        jp := core.NewJetpack()
        jp.EnableDevMode()

        // Create performance panel
        panel := frontend.NewPerformancePanel(jp)
        panel.Show()

        // Register metrics
        fps := 60.0
        jp.RegisterMetric(
                core.MetricFPS,
                "fps",
                "Frames per second",
                "fps",
                &fps,
                []string{"performance"},
        )

        // Initialize Lighthouse monitor
        lighthouse := frontend.NewLighthouseMonitor(jp)
        
        // Run Lighthouse audit
        _, err := lighthouse.RunAudit("http://localhost:8080")
        if err != nil {
                log.Printf("Failed to run Lighthouse audit: %v", err)
        }

        // Start exporting metrics
        jp.ExportEnabled = true
        jp.ExportEndpoint = "http://metrics.example.com"
        jp.StartExporting()

        // Start HTTP server
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                // Record FPS metric
                jp.RecordMetric("fps", 58.5)
                
                // Serve HTML with performance panel
                html := `<!DOCTYPE html><html><body><h1>Hello World</h1></body></html>`
                htmlWithPanel, _ := panel.InjectIntoHTML(html)
                w.Header().Set("Content-Type", "text/html")
                w.Write([]byte(htmlWithPanel))
        })

        log.Fatal(http.ListenAndServe(":8080", nil))
}
```

## 📚 Documentation

### Component System
- [Component System](docs/component-system.md)

### Syntax and Abbreviations
- [GoScript Syntax Glossary](./docs/SYNTAX.md)

### Hyper Data Format
- [Hyper Format](./docs/HYPER.md)

### Gocsx Styling Layer
- [Gocsx Documentation](./README_GOCSX.md)

### GoScale API and Database
- [GoScale Documentation](./README_GOSCALE.md)

### GOPM Package Manager
- [GOPM Documentation](./README_GOPM.md)

### Jetpack Performance Monitoring
- [Jetpack Documentation](./README_JETPACK.md)

## 🏗️ Architecture

GoScript follows a layered language architecture that lets the language core, UI runtime, styling layer, build tools, and service layers work independently or together:

```
GoScript
├── Gocsx (Styling Layer)
│   ├── Core
│   │   ├── Configuration
│   │   ├── CSS Generator
│   │   └── Component System
│   ├── Platforms
│   │   ├── Web
│   │   ├── Mobile
│   │   └── Desktop
│   └── Components
│       ├── Button
│       ├── Card
│       └── ...
├── GoEngine (2D/3D Rendering)
│   ├── Core
│   │   ├── Engine
│   │   └── Scene Graph
│   ├── WebGPU
│   │   ├── Renderer
│   │   └── Shaders
│   └── Canvas2D
│       ├── Renderer
│       └── Sprites
├── GoScale (API and Database)
│   ├── API
│   │   ├── Schema
│   │   ├── Resolvers
│   │   └── Edge Computing
│   └── Database
│       ├── PostgreSQL
│       ├── NoSQL
│       └── TimeSeries
├── GOPM (Package Manager)
│   ├── Core
│   │   ├── Package Management
│   │   └── Dependency Resolution
│   └── Commands
│       ├── CSS Commands
│       ├── WebGPU Commands
│       ├── API Commands
│       └── DB Commands
└── Jetpack (Performance Monitoring)
    ├── Core
    │   ├── Metrics
    │   └── Panel
    ├── Frontend
    │   ├── Lighthouse
    │   └── Web Vitals
    ├── Backend
    │   ├── API Monitoring
    │   └── System Metrics
    └── Security
        ├── Vulnerability Scanning
        └── Anomaly Detection
```

## 🔧 Configuration

GoScript uses a unified Hyper configuration approach across all components:

```hyper
<goscript-config>
  <gocsx>
    <theme>default</theme>
    <breakpoints>
      <item key="sm">640px</item>
      <item key="md">768px</item>
      <item key="lg">1024px</item>
      <item key="xl">1280px</item>
    </breakpoints>
  </gocsx>
  <engine>
    <webgpu>
      <enabled>true</enabled>
      <shaders>./shaders</shaders>
    </webgpu>
    <canvas2d>
      <enabled>true</enabled>
      <sprites>./sprites</sprites>
    </canvas2d>
  </engine>
  <goscale>
    <api>
      <port>8080</port>
      <edge-enabled>true</edge-enabled>
    </api>
    <db>
      <connection-string>localhost:5432</connection-string>
      <time-series-enabled>true</time-series-enabled>
    </db>
  </goscale>
  <jetpack>
    <monitoring>
      <enabled>true</enabled>
      <metrics>
        <item>fps</item>
        <item>memory_usage</item>
        <item>api_latency</item>
      </metrics>
    </monitoring>
    <panel>
      <enabled>true</enabled>
      <position>bottom-right</position>
      <opacity>0.8</opacity>
    </panel>
  </jetpack>
</goscript-config>
```

## 📋 Feature Comparison

### GoScript vs the Legacy Browser Stack

- **Language**: GoScript is Go-native, the legacy browser stack is the browser's current native scripting layer.
- **Team workflow**: GoScript keeps product logic and UI logic in one Go mental model.
- **Performance**: GoScript inherits Go's compiled, predictable runtime characteristics.
- **Type safety**: GoScript benefits from Go's compile-time guarantees and explicitness.
- **Deployment**: GoScript fits a Go-shaped delivery story instead of a JS package graph.
- **Ownership**: Go teams stay in one language instead of splitting the product across Go and JS/TS.
- **Best fit**: GoScript is for teams that want the web to feel native to Go.

### GoScript vs React

- **Language**: GoScript uses Go, React uses the legacy browser stack.
- **Performance**: GoScript offers better performance due to Go's efficiency.
- **Type Safety**: GoScript has stronger type safety through Go's type system.
- **Learning Curve**: Familiar app structure for React developers, but grounded in Go.
- **Ecosystem**: React has a larger ecosystem, while GoScript integrates with Go libraries.
- **CSS Layer**: GoScript includes Gocsx, React requires external libraries.
- **3D Rendering**: GoScript includes WebGPU integration, React requires external libraries.
- **API System**: GoScript includes GoScale, React requires external libraries.
- **Performance Monitoring**: GoScript includes Jetpack, React requires external libraries.

### Gocsx vs Tailwind CSS

- **Language**: Gocsx uses Go, Tailwind uses the legacy browser stack/CSS
- **Platforms**: Gocsx supports web, mobile, and AR/VR, Tailwind is web-only
- **Type Safety**: Gocsx has type safety, Tailwind does not
- **Components**: Gocsx has built-in components, Tailwind requires additional libraries
- **Customization**: Both have powerful customization options

### GoScript WebGPU vs Three.js

- **Language**: GoScript uses Go, Three.js uses the legacy browser stack
- **Integration**: GoScript offers tighter integration with the application
- **Performance**: GoScript can achieve better performance through Go
- **Type Safety**: GoScript has stronger type safety
- **Features**: Three.js has more features currently, but GoScript is rapidly evolving

## Why GoScript Matters

GoScript matters because it gives the Go community a native blessing: a way to build modern web apps without asking every team member to learn a second language just to ship the frontend.

That means less context switching, fewer seams between backend and UI, and more of the product living in the language Go developers already trust. The point is not to replace the legacy browser scripting layer in the browser. The point is to let Go teams own more of the stack in Go, with the same clarity and discipline that made them choose Go in the first place.

## 🔄 Roadmap

- **Mobile Platform Adapter**: Native mobile support for iOS and Android
- **AR/VR Platform Adapter**: Support for AR and VR applications
- **Advanced Component Library**: Expanded set of UI components
- **Testing Infrastructure**: Comprehensive testing tools
- **IDE Integration**: Integration with popular IDEs
- **Animation System**: Advanced animation and transition system
- **Machine Learning Integration**: Integration with ML toolkits and inference runtimes
- **Serverless Deployment**: Support for serverless deployment
- **Multi-tenant Support**: Built-in multi-tenant capabilities
- **Internationalization**: Built-in i18n support

## 📦 Examples

Check out the examples in the repository:

### Web Components
- `pkg/components/counter.go`: Demonstrates class-based components with state
- `pkg/components/home.go`: Shows how to use the context API and functional components

### Styling Layer
- `cmd/gocsx_demo`: Basic styling layer demo

### 2D and 3D Applications
- `cmd/gocsx_2d_demo`: 2D canvas application demo
- `cmd/gocsx_3d_demo`: 3D WebGPU application demo

### API and Database
- `cmd/goscale_demo`: API and database demo

### Performance Monitoring
- `cmd/jetpack_demo`: Performance monitoring demo

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

Apache License, Version 2.0
