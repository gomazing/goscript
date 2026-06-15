# GoScript Syntax and Abbreviation Glossary

This glossary uses the format:

`Legacy / Common term` - `GoScript shorthand` - `What it is` - `Example Syntax`

Examples are conceptual and may evolve. The goal is to help humans and AI agents map old web ideas onto GoScript ideas quickly.

Naming collisions to watch:
- `Go FAST` is both a pillar name and a compiler-tree family name, so we use `Go FAST tree` when we want the AST meaning to stay clear.
- `manifest` is overloaded, so BO uses `pack` for export descriptors while GOPM can still use project-level manifests.
- `Go Surface` is the styling/runtime layer, while `Go WRT` is the human-facing write syntax.
- `pack` is BO-specific and should not be used as a generic project config word.
- `hyper` is the rich data/document format and should stay separate from `pack`.

## 1. Roadmap Pillars

| Legacy / Common term | GoScript shorthand | What it is | Example Syntax |
| --- | --- | --- | --- |
| Performance work | Go FAST | The performance and compiler pillar | `fast build ./app --opt` |
| Visual rendering and layout | Go PAINT | The rendering and spatial UI pillar | `paint stage { ... }` |
| Realtime sync and collaboration | Go IRT | The In Real Time pillar | `irt subscribe room="sales"` |
| Profiling and observability | Go Jetpack | The performance and verification pillar | `jetpack profile ./app` |

## 2. Language and Compiler Terms

| Legacy / Common term | GoScript shorthand | What it is | Example Syntax |
| --- | --- | --- | --- |
| JSX | Go WRT | Human-facing component syntax and compatibility input | `wrt <Hero title="Hello" />` |
| GOCSX / `gocsx` | Go Surface | The Go-native styling and surface layer | `surface Button { ... }` |
| GoUIX / `gouix` | GoScript UI runtime | The reactive Go-centric UI runtime | `ui := gouix.New()` |
| AST | Go FAST tree | The compiler tree used for parsing and transforms inside the Go FAST pipeline | `fast parse source.gsx` |
| IR | Go RUN | The machine-facing intermediate representation | `run emit component` |
| Component Graph | Go Graft | The structured UI graph the runtime can render or slice | `graft Card { title: "Sales" }` |
| Sum type / ADT | Variant | A tagged value with exhaustive matching support | `variant Maybe.Some("hello")` |
| Surface syntax | Go WRT | The readable syntax people write before compilation | `wrt { ... }` |

## 3. Rendering and Spatial UI

| Legacy / Common term | GoScript shorthand | What it is | Example Syntax |
| --- | --- | --- | --- |
| DOM | DOM bypass / Go PAINT | The legacy browser tree model GoScript wants to sidestep for rich surfaces | `paint.use(canvas)` |
| Canvas | Canvas | A direct drawing surface for pixels, charts, and spatial interfaces | `canvas.draw(...)` |
| GPU | GPU | The hardware used for rendering and compute acceleration | `gpu.dispatch(...)` |
| WebGPU | WebGPU | The browser GPU API used for direct rendering and compute | `paint.backend("webgpu")` |
| 2D | 2D | Flat spatial UI and document-like composition | `stage.mode("2d")` |
| 3D | 3D | Scene-based rendering and animation-heavy surfaces | `stage.mode("3d")` |
| Vibe | Go Vibe | The motion and interaction layer | `vibe.spring(...)` |
| Pixel plotting | Go PAINT | Explicit coordinate rendering | `paint.at(120, 48, "Button")` |
| Coordinate vectors | Go PAINT | Numeric positions and sizes for rendering and hit testing | `rect := vec(20, 40, 300, 200)` |

## 4. Build, Package, and Export

| Legacy / Common term | GoScript shorthand | What it is | Example Syntax |
| --- | --- | --- | --- |
| Build Out | BO | The transcompiling exporter for pack-driven deployable artifacts | `bo export erp.pack -exe` |
| Go Package Manager | GOPM | Project setup, manifests, lockfiles, and package workflows | `gopm init` |
| Project manifest | Hyper | The project-level description file used by GOPM | `gopm.hyper` |
| Pack file | Pack | The BO export descriptor that contains output, slices, assets, and build metadata | `erp.pack` |
| Lockfile | Lockfile | The file that pins exact build and dependency resolution | `gopm.lock` |
| Slice | Slice | An internal export unit declared inside a pack file | `slice /calc` |
| Bundle | Bundle | The compiled output for a target environment | `bundle build` |

## 6. Rich Data and Documents

| Legacy / Common term | GoScript shorthand | What it is | Example Syntax |
| --- | --- | --- | --- |
| Legacy structured data | Hyper | A rich document and data exchange format for text, images, graphs, tables, and embedded blocks | `report.hyper` |
| Document block | Hyper block | A typed block inside a Hyper file | `text "Sales up 17%"` |
| Image / graph / table | Hyper elements | Structured visual blocks that live inside a Hyper document | `graph line { ... }` |

## 7. Service, Data, and Trust

| Legacy / Common term | GoScript shorthand | What it is | Example Syntax |
| --- | --- | --- | --- |
| Application Programming Interface | Go TALK | The protocol compatibility and service interface layer | `talk connect room="sales"` |
| Database | DB | Persistent storage | `db.query(...)` |
| Structured data | Hyper | The rich text and data exchange format used for documents and payloads | `hyper.parse(...)` |
| RPC | RPC | Remote function calls across process or network boundaries | `rpc call` |
| gRPC | gRPC | Typed high-performance RPC transport | `grpc.stub(...)` |
| GraphQL | GraphQL | Schema-driven querying for flexible data fetching | `graphql.query(...)` |
| Role-Based Access Control | RBAC | Permission checks based on roles | `allow role="admin"` |
| GoScale | GoScale | The API, database, and edge service layer | `goscale serve` |

## 8. Realtime and Topology

| Legacy / Common term | GoScript shorthand | What it is | Example Syntax |
| --- | --- | --- | --- |
| Client Server Mode | cs | Centralized topology mode | `mode := cs` |
| Swarm mode | sw | Distributed topology mode | `mode := sw` |
| Swarm | Swarm | Distributed execution model | `swarm deploy` |
| Realtime updates | Go IRT | Low-latency sync and collaboration behavior | `irt subscribe room="alpha"` |

## 9. Observability and Verification

| Legacy / Common term | GoScript shorthand | What it is | Example Syntax |
| --- | --- | --- | --- |
| Jetpack | Go Jetpack | The observability, profiling, and proof layer | `jetpack watch` |
| Lighthouse | Jet Fix | Browser-style audit and scoring tool | `jetfix run lighthouse` |
| Benchmark | Benchmark | A repeatable performance test | `bench run` |
| Diagnostics | Diagnostics | Tools that reveal bottlenecks, regressions, or anomalies | `diag status` |
| Tracing | Jet Trace | Span tracking and flow inspection | `jet trace request` |
| Profiling | Jet Run | CPU, memory, and latency measurement | `jet run cpu` |

## 10. Directory Shorthands

| Legacy / Common term | GoScript shorthand | What it is | Example Syntax |
| --- | --- | --- | --- |
| `base/` | base | Foundation docs and configuration | `use base/` |
| `agents/` | agents | Agent roles and automation docs | `load agents/` |
| `core/` | core | Core runtime internals | `core/renderer` |
| `app/` | app | App-level code | `app/pages` |
| `packs/` | packs | Pack files | `packs/admin.pack` |

## 11. How To Read The Names

- Use `Go FAST` when you are talking about speed, allocations, compiler work, or the hot path.
- Use `Go PAINT` when you are talking about canvas, spatial UI, 2D/3D composition, or DOM bypass.
- Use `Go IRT` when you are talking about live sync, collaboration, distributed modules, or streaming state.
- Use `Go Jetpack` when you are talking about profiling, observability, benchmarks, or verification.
- Use `Go WRT` for the human-facing component syntax.
- Use `Go Graft` for the machine-facing UI structure.
- Use `Go TALK` for protocol compatibility, APIs, and service interfaces.
- Use `Go RUN` for the intermediate representation.
- Use `Variant` for tagged values and sum-type style matching.
- Use `Jet Fix`, `Jet Trace`, and `Jet Run` for the Jetpack tools.

When in doubt, write the full phrase first. That keeps the project readable for both humans and AI agents.
