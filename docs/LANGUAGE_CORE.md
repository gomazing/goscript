# GoScript Language Core

GoScript is a language and runtime first.
Frameworks, app shells, and domain platforms are meant to be built on top of these primitives, not embedded into the core.

The canonical architecture direction is documented in [ARCHITECTURE_PRINCIPLES.md](ARCHITECTURE_PRINCIPLES.md).

This document is the canonical read for the language contract:

- source units
- module boundaries
- lowering
- runtime semantics
- structural UI graphs
- machine-facing runtime graphs

## Core Objects

### `SourceUnit`

One source unit is the smallest language-aware building block.
It represents a file or logical unit of GoScript code.

It carries:

- `Path`
- `Kind`
- `Module`
- `Content`
- `EntryPoint`
- `Experimental`
- `Semantics`

### `ModuleSpec`

One module spec describes a unit of buildable GoScript capability.
It is used for:

- module registration
- dependency tracking
- build ordering
- package boundaries
- pack-driven export rules

### `LanguageCore`

The `LanguageCore` is the language-level contract object.
It stores:

- registered modules
- registered source units
- build order
- diagnostics
- lowering hooks

## Runtime Semantics

GoScript does not treat every source unit the same way.
Each unit has a runtime shape.

The canonical runtime planes are:

- `frontend`
- `backend`
- `realtime`
- `hybrid`

The runtime semantics tell the toolchain whether a unit is more like:

- a visual surface
- a backend module
- a realtime event path
- a mixed deployment unit

## Lowering Model

GoScript should always be able to move through this chain:

1. Source unit
2. Normalized module context
3. Validation
4. Lowering
5. Structural graph
6. Machine graph
7. Renderable output or runtime execution

The structural graph is the human-facing layer.
The machine graph is the runtime-facing layer.
The transport-compatibility layer is Go TALK, which is where external protocol bridges, codec negotiation, bidirectional frames, and service contracts stay native to the language.

In GoScript naming:

- `Go Graft` is the structural graph
- `Go RUN` is the lowered runtime graph
- `Go TALK` is the protocol compatibility layer

## What the Core Must Do

The language core must stay responsible for:

- syntax boundaries
- lowering boundaries
- runtime semantics
- transport compatibility and frame negotiation
- module resolution
- dependency ordering
- source-level diagnostics
- deterministic snapshots

It should not become responsible for:

- Next.js-style app shells
- Django-style framework behavior
- opinionated business workflows
- product-specific routing conventions
- domain frameworks that belong in the ecosystem

## Minimal Example

```go
core := goscript.NewLanguageCore()

_ = core.RegisterModule(goscript.ModuleSpec{
	Name: "pages",
	Path: "/modules/pages",
})

_ = core.AddSource(goscript.SourceUnit{
	Path:    "pages/home.gsx",
	Content: "TODO build the home surface",
})

lowered, err := core.LowerSource("pages/home.gsx", func(unit goscript.SourceUnit) (goscript.GraftNode, []goscript.Diagnostic, error) {
	return goscript.Graft("div", goscript.Props{"id": "shell"}, goscript.GraftText("hello")), nil, nil
})
```

The key point is not the exact syntax.
The key point is that the language core owns the contract, and everything else builds on top of it.

## Stable Direction

The language should continue to grow in this order:

1. Language core
2. `Go FAST`
3. `Go PAINT`
4. `Go IRT`
5. `Go Jetpack`
6. `Go TALK`
7. Ecosystem frameworks built by developers
