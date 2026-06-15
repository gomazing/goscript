# GoScript Deployment Modes

GoScript uses topology modes to describe how an app is deployed and operated.

## CS Mode Deployment

`cs` means client-server mode.

Use this when one primary server owns the backend contract and clients stay thin.

### Deployment shape

- One main server process or server cluster
- One application contract source of truth
- Centralized routing, auth, state, and policy
- Best fit for websites, dashboards, internal tools, and smaller ERP deployments

### Current repo support

- `gopm setup --cs` scaffolds the CS folder layout
- `base/modes/cs.md` defines the topology rules
- `buildout.Manifest.Mode = "cs"` is validated by the pack exporter
- `packs/<name>.pack` can describe the deployable output

### Deployment focus

- Single artifact deployment
- Fast server startup and lean request dispatch
- Clean separation between UI pages, API routes, and services

## SW Mode Deployment

`sw` means swarm mode.

Use this when modules must run on different nodes with explicit trust and fallback rules.

### Deployment shape

- Module-aware deployment
- Nodes can own different parts of the app
- Each module has a trust boundary and fallback behavior
- Best fit for modular ERP, control/data plane splits, ticketing, and distributed platforms

### Current repo support

- `gopm setup --sw` scaffolds the swarm folder layout
- `base/modes/sw.md` defines node placement and trust rules
- `buildout.Manifest.Mode = "sw"` is validated by the pack exporter
- `swarm-policies/` and `trust/` folders are scaffolded for node policy work

### Deployment focus

- Module placement and node assignment
- Pack-driven export of deployable app slices
- Trust, routing, and fallback behavior
- Better observability and rollout control than a single-server app

## What Still Needs Tightening

- A deployment matrix that maps app types to `cs` or `sw`
- Better rollout and health-check stories for multi-node deployments
- Stronger pack-level metadata for node assignment and trust
- Jetpack-backed profiling for deployment bottlenecks

