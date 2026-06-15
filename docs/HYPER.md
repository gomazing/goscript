# Hyper Format

Hyper is GoScript's rich document and data exchange format.

It is designed for:

- text
- tables
- graphs
- images
- blocks
- embeds
- structured metadata

Hyper exists because legacy text-based config formats are great for structured data, but they are not expressive enough for the kind of rich, visual, mixed-content documents GoScript wants to move around.

## Design goals

- Human-readable and AI-readable
- Typed blocks instead of loose blobs
- Easy to diff in version control
- Safe to parse
- Rich enough for docs, dashboards, notes, reports, and workspace content
- Compatible with Go PAINT and Go WRT workflows

## Core idea

A Hyper file is a tree of typed nodes.

Example:

```hyper
doc "Q2 Review" {
  text "Revenue is up 17%."

  table "summary" {
    row "North America" "42%"
    row "Europe" "31%"
    row "APAC" "27%"
  }

  graph "growth" {
    type: "line"
    series: "revenue"
    series: "margin"
  }

  image "dashboard" {
    src: "./assets/q2-dashboard.png"
    alt: "Q2 dashboard snapshot"
  }
}
```

## Block types

Recommended built-in blocks:

- `doc`
- `text`
- `image`
- `table`
- `graph`
- `chart`
- `code`
- `callout`
- `embed`
- `link`

## How it differs from HTML

- Hyper is typed and document-first, not browser-first.
- Hyper is intended for exchange and rendering, not for arbitrary scripting.
- Hyper can render into Go PAINT surfaces, web views, document editors, or workspace tools.
- Hyper keeps structure explicit so AI agents can reason about it more easily than free-form HTML.

## How it differs from legacy text config

- Hyper is better for mixed structured content where layout and data live together.
- Hyper can describe a table, a graph, and a paragraph in one document without forcing everything into plain key/value pairs.

## Relationship to GoScript

- `Go WRT` is for human-facing write syntax in code.
- `Go Surface` is the styling/runtime layer.
- `Go PAINT` renders the visual result.
- `Hyper` is the cross-tool content format for documents and rich data.

## Naming rule

Use `.hyper` when the file is meant to carry rich content and structured blocks.
Use `.pack` when the file is meant to drive BO exports and build targets.

That separation keeps the new ecosystem readable.
