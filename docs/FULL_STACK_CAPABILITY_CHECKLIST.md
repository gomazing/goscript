# GoScript Full-Stack Capability Checklist

GoScript should stay a language and runtime, not a framework clone.
This checklist splits the work into frontend and backend pillars so we keep both planes first-class.
The next milestone is to finish the programming-language capabilities first, then let frameworks grow on top of those primitives.
Several core pieces are already landed in code; the remaining work is called out below so we keep the checklist honest.
Composable architecture belongs in the ecosystem layer, not in the language core.
GoScript's position is 8/10: framework-ready language with batteries included, not a full framework.

## Progress Snapshot

This is the current working order for the language/runtime track:

1. Programming language core - parser, syntax, lowering, runtime semantics, type model, and module boundaries.
2. Go FAST - remove hot-path overhead and make the compiler/runtime pipeline cheaper.
3. Go PAINT - improve the spatial rendering model and direct visual composition.
4. Go IRT - strengthen realtime sync, streaming, and distributed module behavior.
5. Go Jetpack - expand measurement, profiling, and proof so changes stay trustworthy.
6. Go TALK - keep transport frames, service contracts, and protocol compatibility simple, typed, and easy to integrate.
7. Shared foundations - keep pack/export, slice boundaries, and AI-readable tooling aligned.

Related scorecard:

- [Scenario delay scorecard](SCENARIO_DELAY_SCORECARD.md) for websites, PWAs, ERP, SaaS control/data plane, global pass, stock exchange, blockchain, IRCTC/ticketing, FPS, and MMORPG scenarios.
- [Deployment modes](DEPLOYMENT_MODES.md) for `cs` and `sw` deployment shape, repo support, and open gaps.

Current state:

- The language core now has explicit source, lowering, runtime semantics, and module-boundary contracts in code, but the public reference still needs to be finalized.
- `SourceUnit`, `LanguageCore`, `LoweredUnit`, and `RuntimeSemantics` now exist as the canonical core types.
- Frontend foundations exist, but the render path still needs deeper compile-time lowering and paint-first execution.
- Go FAST now has a compile-time surface-lowering path, but more hot-path work is still needed.
- GSX transpilation now caches normalized inputs so repeated lowering can stay off the hot path.
- Backend foundations exist, but API, auth, and realtime behavior still need stronger first-class runtime support.
- Go TALK now exists as a native compatibility layer for bidirectional frames, codec negotiation, service contracts, and future external protocol bridges.
- Observability exists, but it should become the default way we prove every performance claim.
- The language surface should remain open for ecosystem frameworks, not absorb framework behavior into the core.
- Go Graft / Go RUN now exist as the first structured UI graph and lowered runtime node primitives.
- Router param routes are now compiled into segments instead of being re-parsed on every request.
- Go PAINT now has native geometry, hit-testing, and spatial index primitives in the core package.
- Go IRT now has a safer publish path for realtime topics after unsubscribe events.
- RealtimeHub now has batch publish support so dense collaborative updates can amortize lock overhead.
- SyncQueue enqueue/drain now avoid full-queue snapshot copying on every notification, and batch enqueue now amortizes lock overhead.
- Scheduler submission now respects context cancellation so a full queue cannot deadlock submitters after shutdown.
- LSP completion now uses cached sorted names and prefix narrowing instead of scanning every symbol on each request.
- Hydration/SSR now uses a direct builder path to reduce template overhead and allocation pressure.
- `gopm setup` now emits a batteries-included starter scaffold with `go.mod`, `README.md`, `cmd/server/main.go`, `app/pages/home.gsx`, and `app/pages/home.go`.
- The current open work is concentrated around type-system depth, runtime portability, UI reactivity, package distribution, interoperability, and proving the scenario scores with repeatable benchmarks.
- The first tagged-variant / exhaustive-match primitive is now in code, but the syntax-level language integration is still pending.

## Next Build Order

1. Programming language core - syntax, lowering, type model, runtime semantics, and module rules.
2. Go FAST - compile-time lowering, lower allocations, and lean hot-path execution.
3. Go PAINT - spatial rendering, hit testing, and direct 2D/3D composition.
4. Go IRT - realtime sync, event flow, and module-level collaboration.
5. Go Jetpack - profiling, benchmark, trace, and verification loops.
6. Go TALK - protocol compatibility, bidirectional frames, and typed service clarity.
7. Go Graft / Go RUN - structured UI graph and lowered runtime nodes.
8. Shared foundations - pack/export, slices, docs, and AI-readable tooling.

## Frontend Pillar

The frontend side must feel native, fast, and spatial.

### Go PAINT

- [ ] Canvas-first rendering pipeline for rich surfaces
- [ ] Direct 2D composition for documents, dashboards, and editor UIs
- [ ] 3D scene support on the same surface when needed
- [ ] Pixel plotting and coordinate-vector layout
- [ ] Click-target collision detection and hit maps
- [ ] DOM-bypass rendering for complex visual apps

### Go FAST

- [x] Compile-time lowering for UI structures
- [ ] Remove runtime parsing from the hot render path
- [ ] Lower-allocation component render loops
- [ ] Faster hydration and view update paths
- [x] Faster route dispatch for frontend navigation
- [ ] Faster SSR and page assembly

### Go IRT

- [ ] Live state sync for collaborative views
- [ ] Realtime cursor, presence, and editing events
- [ ] Streaming updates for dashboards and docs
- [ ] Shared module state across screens
- [ ] Realtime module coordination for multi-surface apps

### Go Jetpack

- [ ] Render timing and frame diagnostics
- [ ] Input latency profiling
- [ ] Memory and allocation inspection for UI paths
- [ ] Visual regression and fidelity checks
- [ ] Benchmark reports for render, layout, and interaction

## Backend Pillar

The backend side must scale cleanly, stay predictable, and stay easy for AI and humans to reason about.

### Go TALK

- [ ] Clear protocol compatibility layer for APIs and future TALK transport
- [ ] Stable request, response, and frame contracts
- [ ] Typed boundaries for internal and external calls
- [ ] Bidirectional text, media, data-stream, and sensor transport
- [ ] Stream-friendly endpoints for realtime clients
- [ ] Simple integration points for storage, auth, and external codecs

### Go FAST

- [ ] Efficient routing and middleware chains
- [ ] Low-overhead request dispatch
- [ ] Allocation-aware request processing
- [ ] Fast serialization and parsing paths
- [ ] Hot-path tuning for high-concurrency services

### Go IRT

- [ ] Realtime server events and fan-out
- [ ] Swarm-style distributed module execution
- [ ] Background jobs and scheduler-driven work
- [ ] Live collaboration backends
- [ ] Module-to-module communication with trust boundaries

### Go Jetpack

- [ ] Tracing across API, DB, and event paths
- [ ] Profiling for server hot spots
- [ ] Security and policy diagnostics
- [ ] Benchmark harnesses for backend subsystems
- [ ] Regression proof for performance changes

## Shared Full-Stack Foundations

- [ ] One language for frontend and backend primitives
- [ ] Shared types across UI, service, and module layers
- [ ] Explicit parser, lowering, runtime, and transport-contract definitions for the language core
- [ ] Pack-driven export flow for deployable artifacts
- [ ] Clean slice boundaries for modular apps
- [ ] Tooling that helps AI coders inspect and modify the system safely
- [ ] First-class docs and glossary support so humans and agents share the same vocabulary
- [ ] Future-ready architecture for spatial, realtime, and multi-surface applications
- [x] Batteries-included starter paths for fast vibe-coded app generation
- [ ] Composable framework support primitives that ecosystem authors can build on without changing the core language design

## Programming Language Core

This is the part that must exist before any serious composable framework ecosystem can grow on top of GoScript.

- [ ] Stable syntax and grammar for the language surface
- [ ] Parser that lowers into a typed internal form
- [ ] Explicit type model for values, components, modules, transport contracts, and runtime frames
- [ ] Runtime semantics for execution, memory, and concurrency
- [ ] Module and package rules that support pack-based distribution
- [ ] Shared lowering path for UI, backend, and realtime features
- [ ] Error model that is predictable for humans and agents
- [ ] Tooling hooks for Jetpack-style measurement and regression checks
- [ ] Clear bridge from source syntax to Go FAST / Go PAINT / Go IRT / Go TALK / Go Jetpack primitives

## What We Are Not Building

- [ ] A Next.js clone
- [ ] A Django clone
- [ ] An Angular-style opinionated framework inside the core repo
- [ ] A frontend-only language
- [ ] A backend-only language

GoScript should stay a full-stack language with first-class frontend and backend capability, while leaving frameworks to the ecosystem built on top of it.
