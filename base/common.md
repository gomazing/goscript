# Common Rules

GoScript is a guided language for AI-era systems. It is not a framework clone and it is not an unguided language where every agent invents a new project shape.

## AI-First Protocols

- Assume agentic systems are the primary builders and consumers of GoScript
- Prefer explicit contracts, deterministic behavior, and machine-readable structure
- Keep abstractions portable across digital, ternary, MVL, and future quantum-capable targets
- Preserve backward compatibility with digital execution as the baseline runtime
- Make human developers beneficiaries of a machine-friendly language, not the only audience

## Required Decisions

Every AI coder must decide these before generating a project:

1. Mode: `cs` or `sw`
2. Project type: `website`, `app`, or `erp`
3. Which `base/config/*.md` files apply
4. Whether runtime autonomous agents are required

## Shared Project Shape

Use this baseline tree unless a mode or project type explicitly extends it:

```text
project/
  gopm.hyper
  base/
  agents/
  app/
    modules/
    pages/
    components/
    services/
    routes/
    assets/
  core/
  tests/
  docs/
  deploy/
```

## Global Rules

- Keep naming predictable and reusable across projects
- Prefer modules over one-off feature folders
- Keep `core/` outside normal application execution boundaries
- Treat `base/` as build-time guidance and `agents/` as runtime autonomous roles
- Keep the folder tree as uniform as possible across `website`, `app`, and `erp`
