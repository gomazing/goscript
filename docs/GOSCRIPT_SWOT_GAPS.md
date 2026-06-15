# GoScript SWOT Gaps Checklist

This checklist is refreshed from the report input and current repo state.
Timelines are ignored here on purpose. The goal is to separate what is already landed from what still needs hardening or invention.

## Already Landed

- [x] Canonical language core contracts exist: `SourceUnit`, `LanguageCore`, `LoweredUnit`, `RuntimeSemantics`, and module-boundary rules
- [x] Go FAST compile-time surface lowering for UI structures exists
- [x] GSX transpile cache now avoids recomputing repeated lowerings on identical source
- [x] Go Graft / Go RUN structured UI graph and lowered runtime node primitives exist
- [x] Go PAINT native geometry, hit-testing, and spatial index primitives exist
- [x] Go IRT safer publish path for realtime topics exists
- [x] RealtimeHub batch publish support exists for dense collaborative updates
- [x] SyncQueue batch enqueue support exists for dense state bursts
- [x] Go TALK runtime codec negotiation and bidirectional frame compatibility exists
- [x] Router param segments are compiled instead of being re-parsed on every request
- [x] Scheduler submission now respects context cancellation
- [x] SyncQueue enqueue/drain now avoid full-queue snapshot copying on every notification
- [x] LSP completion now uses cached sorted names and prefix narrowing
- [x] Hydration and SSR now use a direct builder path to reduce template overhead and allocations
- [x] Batteries-included starter scaffold now emits `go.mod`, `README.md`, `cmd/server/main.go`, `app/pages/home.gsx`, and `app/pages/home.go`
- [x] Glossary and architecture docs exist to keep the vocabulary aligned for humans and agents

## Partially Landed

- [ ] First-pass tagged variant / exhaustive matching primitive exists, but it still needs language-syntax integration
- [ ] Go FAST still needs lower-allocation hot paths beyond the current compile-time lowering
- [ ] Go PAINT still needs a fuller retained scene graph and spatial render pipeline
- [ ] Go IRT still needs stronger live sync, presence, and swarm coordination semantics
- [ ] Go TALK still needs the external protocol story and transport guarantees finalized
- [ ] BO packaging still needs the pack format, slice rules, and export contract locked
- [ ] Go Jetpack still needs tighter proof loops for render, trace, profile, and regression evidence
- [ ] Scenario scorecards still need repeatable benchmarks for scenario-based comparisons

## Core Gaps Still Open

### Language Core

- [ ] True sum types with exhaustive matching
- [ ] Union and intersection types
- [ ] Null safety at the type level
- [ ] Async / await ergonomics
- [ ] Macro and code-generation system
- [ ] Effect system
- [ ] Type-level string literals
- [ ] Const generics
- [ ] Pattern-matching exhaustiveness checks
- [ ] Linear or ownership-aware resource tracking

### Runtime And Compilation

- [ ] Clear primary compile target story for browser, edge, and server
- [ ] WASM Component Model support
- [ ] WASI support for server and edge
- [ ] Tree shaking and dead-code elimination
- [ ] Incremental compilation
- [ ] Hot module replacement with state preservation
- [ ] Source maps and debug symbols
- [ ] Bundle splitting and module federation
- [ ] Runtime size reduction for practical distribution
- [ ] JS interop bridge for gradual adoption

### UI And Reactivity

- [ ] Fine-grained reactivity or signals
- [ ] Portal / teleport support
- [ ] Fragment support
- [ ] Ref and forward-ref support
- [ ] Event system for efficient dispatch
- [ ] Suspense and error boundaries
- [ ] Concurrent rendering
- [ ] Hydration mismatch detection
- [ ] SSE and WebSocket hooks
- [ ] Observer hooks for intersection, resize, and mutation

### Component System

- [ ] Compound component pattern
- [ ] Render props / slot pattern
- [ ] Lazy and dynamic components
- [ ] Memoization helpers
- [ ] Imperative handle support
- [ ] Layout effect support
- [ ] Deferred values and transitions
- [ ] Optimistic updates
- [ ] Action API
- [ ] Context selectors
- [ ] DevTools integration

### Motion And Spatial UI

- [ ] Spring physics engine
- [ ] Layout animations
- [ ] Shared element transitions
- [ ] Gesture recognition
- [ ] Scroll-linked animations
- [ ] Reduced-motion support
- [ ] Animation controls and sequencing
- [ ] GPU acceleration hints
- [ ] 3D transforms
- [ ] Canvas and WebGPU animation integration

### Accessibility

- [ ] Focus management
- [ ] Roving tab index
- [ ] ARIA helpers
- [ ] Keyboard navigation primitives
- [ ] Screen-reader announcers
- [ ] Skip links
- [ ] Focus-visible support
- [ ] Contrast checking
- [ ] High-contrast mode support
- [ ] Accessibility-tree inspection
- [ ] Automated a11y testing

### State, Forms, And Data

- [ ] Signals-based state management
- [ ] Derived/computed signals
- [ ] Batch updates and effect scopes
- [ ] Deep store reactivity
- [ ] Form handling and schema validation
- [ ] Field arrays and multi-step forms
- [ ] Data fetching and cache invalidation
- [ ] Realtime collaboration state sync
- [ ] Undo/redo
- [ ] State machines

### Styling, Testing, And DX

- [ ] Styling and theming system
- [ ] Design tokens and theme switching
- [ ] Component testing story
- [ ] End-to-end testing story
- [ ] Storybook-style isolated development
- [ ] VS Code extension maturity
- [ ] Rename refactoring and code intelligence
- [ ] Coverage and regression proofing

### Package Management And Distribution

- [ ] Component registry
- [ ] Component installation flow
- [ ] Version resolution and lock files
- [ ] Monorepo support
- [ ] Private registries
- [ ] Dependency deduplication
- [ ] Peer dependency rules
- [ ] Overrides and resolutions
- [ ] License and vulnerability scanning
- [ ] Bundle analysis

### Performance, Security, And Ops

- [ ] Image optimization
- [ ] Font optimization
- [ ] Script loading control
- [ ] Prefetching and preloading
- [ ] Tree-shaking proof
- [ ] Compression pipeline
- [ ] PWA and offline support
- [ ] Core Web Vitals optimization
- [ ] Memory-leak detection
- [ ] FPS monitoring
- [ ] Long-task detection
- [ ] CSP and secure headers
- [ ] XSS and CSRF protection
- [ ] Input sanitization
- [ ] Authentication and authorization primitives
- [ ] Rate limiting

### Interoperability And Adoption

- [ ] GoScript integration stories for existing ecosystems
- [ ] Web Components export
- [ ] JS library wrappers
- [ ] Type-definition generation
- [ ] API gateway and queue integration
- [ ] Database integration story
- [ ] CMS and e-commerce integration
- [ ] Analytics and auth-provider integration
- [ ] Message bus integration
- [ ] Search integration

### Composable Architecture Support

- [ ] Module system that enables packaged business capabilities in ecosystem code
- [ ] WIT-style interfaces for ecosystem modules and boundaries
- [ ] Discovery and registry search for reusable pieces
- [ ] Build-time orchestration support for frameworks and apps built on GoScript
- [ ] Isolation so each module can fail independently
- [ ] Interchangeability across modules that share the same contract
- [ ] Event-driven communication between components
- [ ] Explicit state ownership for every module
- [ ] Test isolation with mocked dependencies
- [ ] Auto-generated documentation from interfaces and package metadata

### Enterprise Adoption

- [ ] Enterprise support contracts and SLAs
- [ ] Training and certification path
- [ ] Migration path from existing stacks
- [ ] Case studies and production proof
- [ ] Compliance posture for regulated buyers
- [ ] Security review and procurement readiness
- [ ] Internal enablement docs for large teams

### Hardware And Future Targets

- [ ] DGX Spark support path
- [ ] Apple M5+ support path
- [ ] NPU-aware execution path
- [ ] TPU-aware execution path
- [ ] Binary / ternary / MVL logic support
- [ ] Protocol story for future multi-stream transport

## What To Keep Out Of Core

- [ ] Do not turn GoScript into a Next.js clone
- [ ] Do not turn GoScript into a Django clone
- [ ] Do not turn GoScript into an Angular-style opinionated framework
- [ ] Do not collapse the language into a frontend-only toy
- [ ] Do not collapse the language into a backend-only tool

## Next Build Order

- [ ] Finalize the syntax glossary for Go FAST, Go PAINT, Go IRT, Go TALK, and Go Jetpack
- [ ] Define the compile pipeline from `source.gsx` to runtime artifacts
- [ ] Lock the BO manifest format and package rules
- [ ] Expand scenario scorecards for web, PWA, ERP, SaaS, gaming, and streaming workloads
- [ ] Specify the slice isolation contract for crash recovery and partial reloads
- [ ] Define the default transport story for Go TALK and Go IRT
