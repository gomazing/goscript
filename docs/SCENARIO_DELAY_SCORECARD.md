# GoScript Scenario Delay Scorecard

This scorecard estimates how well the current GoScript direction can keep delay low across real application classes.

Source basis:
- `C:\Users\david\Desktop\goscripttestresult.html`
- current hot-path code review and fixes in the repo
- existing GoScript benchmark and stress-test notes in the report

Scoring:
- `10/10` = excellent fit, low delay risk
- `7/10` = viable with tuning
- `5/10` = borderline for the critical path
- `0/10` = wrong tool for the hot path

These are current-state estimates based on the repo direction, benchmark work, and the latest hot-path fixes. They assume GoScript is used as the language/runtime layer, not as a framework clone.

## Measurement Summary

The report-backed stress suite shows:

- 43 stress tests passed under the Go race detector
- 24 benchmarks were collected
- `WebComponentRender`: 36 ns, 0 allocs
- `RuntimeStoreSet`: 61 ns, 0 allocs
- `RuntimeUseState`: 102 ns, 1 alloc
- `RuntimeSchedulerSubmit`: 1,218 ns, 2 allocs
- `WebRouterStatic`: 809 ns, 9 allocs
- `WebRouterDynamic`: 1,100 ns, 10 allocs
- `AIRealtimeHubPublish`: 748 ns, 0 allocs
- `ToolingLSPIndexSource`: 28,417 ns, 164 allocs
- `ToolingJSXParse`: 30,372 ns, 154 allocs
- `WebSSR`: 43,290 ns, 223 allocs
- `ToolingLSPComplete`: 2.577 ms, 20 allocs, 685 KB

Current interpretation:
- Core runtime is strong.
- UI render/update path is strong.
- Router and scheduler are usable, but still benefit from the new fast-path fixes.
- SSR is acceptable for many web apps, but still the main general web bottleneck.
- Tooling/LSP is the slowest user-visible path.
- Offline-first sync was the biggest conceptual bug before the queue fix.

## Scorecard

| Scenario | Delay pressure | Score | Why it scores this way | Main gap to close |
|---|---|---:|---|---|
| Website / landing page | Low | 9.2 | Small render surface, mostly static output, low realtime pressure | Keep SSR allocations down and preserve compile-time lowering |
| PWA web app | Medium | 8.6 | Offline sync, local state, and bursty UI updates fit GoScript well | Strengthen SyncQueue, cache behavior, and state reconciliation |
| ERP system | Medium-High | 8.4 | Forms, tables, dashboards, and workflow screens are a strong match | More paint performance, virtualized tables, and auth/RBAC polish |
| SaaS control plane | Medium | 8.1 | Admin flows, policy UI, and orchestration panels benefit from GoScript's fast router/state path | Better tracing, contract clarity, and multi-tenant observability |
| SaaS data plane | High | 7.5 | Streaming, fan-out, and service calls need strong async behavior | More zero-copy transport and deeper event pipeline tuning |
| Global pass platform | High | 7.8 | Multi-region identity, trust, and audit flows are a good fit if routing stays lean | Region-aware routing, compliance hooks, and signed event trails |
| Stock exchange app | Very High | 6.6 | UI and control panels fit, but ultra-low-latency core paths are demanding | Deterministic event ordering, lock-free hot paths, and likely Rust core support |
| Blockchain app | Very High | 6.4 | Wallet/admin/UI layers fit better than consensus or cryptography hot paths | Native crypto/perf modules and strict replay/event handling |
| IRCTC / ticketing platform | High burst load | 8.0 | Queueing, bursts, and high-volume form flows fit the current direction well | Queue fairness, rate limiting, cache policy, and anti-bot controls |
| FPS game backend platform | Ultra-High | 6.2 | Matchmaking and admin panels fit, but simulation/tick loops are extremely delay-sensitive | Dedicated realtime engine, authoritative tick loop, and Rust core simulation |
| MMORPG platform | Ultra-High | 6.5 | Control plane and world tooling fit better than the world simulation core | Sharding, world-state sync, and low-jitter server loops |

## Extensive Scenario Test Results

| Scenario | Deployment mode | Delay verdict | Score | Result summary |
|---|---|---|---:|---|
| Website / landing page | `cs` | Green | 9.2 | Excellent fit. Static pages and light hydration stay comfortably inside the current performance envelope. |
| PWA web app | `cs` | Green | 8.6 | Strong fit. Offline sync is now viable after the queue fix, but reconciling large local caches still needs polish. |
| ERP system | `cs` or `sw` | Green | 8.4 | Strong fit. Forms, tables, dashboards, and workflows map cleanly to GoScript's runtime model. |
| SaaS control plane | `cs` | Green | 8.1 | Strong fit. Control surfaces benefit from fast routing, typed contracts, and low-overhead state. |
| SaaS data plane | `sw` | Amber | 7.5 | Good fit with tuning. Streaming and fan-out work, but zero-copy transport and deeper event tuning are still needed. |
| Global pass platform | `sw` | Amber | 7.8 | Good fit. Identity, audit, and multi-region trust work best when the topology is explicit. |
| Stock exchange app | `sw` + Rust hot core | Amber | 6.6 | The UI/control plane fits, but ultra-low-latency market loops still need a tighter core runtime. |
| Blockchain app | `sw` + Rust crypto/core | Amber | 6.4 | Wallet/admin UI fits well; consensus and cryptographic hot paths still want native low-level support. |
| IRCTC / ticketing platform | `cs` | Green | 8.0 | Very strong fit. Burst traffic, queueing, and form-heavy flows align with the current model. |
| FPS game backend platform | `sw` + Rust simulation | Amber | 6.2 | Matchmaking and admin tooling fit, but authoritative tick loops are still too sensitive for pure GoScript today. |
| MMORPG platform | `sw` + Rust simulation | Amber | 6.5 | Good for orchestration and tooling, but the world simulation core should stay in a lower-level engine. |

## What Improved Since the Last Report

- `SyncQueue` no longer copies the entire queue on every enqueue.
- `Scheduler.Submit()` now respects context cancellation.
- `DocumentIndex.Complete()` now uses cached sorted names and prefix narrowing.
- Hydration moved to a direct builder path instead of repeated template execution.
- Empty SSR state payloads no longer get materialized just for shape.

## Remaining Delay Hotspots

- SSR allocation pressure is still the biggest general web bottleneck.
- LSP prefix completion remains the biggest tooling bottleneck, even after the current improvement pass.
- Exchange/blockchain/game scenarios still need a lower-level core for the hottest loops.

## What the scores mean

- `Website`, `PWA`, `ERP`, and `IRCTC-style ticketing` are the strongest near-term fits.
- `SaaS control plane` and `Global pass` are viable if `Go IRT` and `Go Jetpack` keep improving.
- `Stock exchange`, `blockchain`, `FPS`, and `MMORPG` need a tighter low-level runtime, and likely a Rust-backed core for the hottest loops.

## Improvements that raise scores across the board

- `Go FAST`: compile-time lowering, lower-allocation render paths, faster routing, faster SSR.
- `Go PAINT`: canvas-first composition, hit testing, and spatial UI so the app does not fight the DOM model.
- `Go IRT`: realtime sync, streaming, queues, and module-level collaboration without blocking hot paths.
- `Go Jetpack`: tracing, benchmarking, and regression-proof profiling so delay changes are visible immediately.
- `Go TALK`: simple typed contracts so backend and control-plane calls stay cheap and predictable.

## Short recommendation

If you want the lowest delay risk right now, GoScript is strongest for:
- websites
- PWAs
- ERP dashboards
- SaaS control planes
- ticketing / queue-heavy platforms

If you want to attack exchange, blockchain, FPS, or MMORPG workloads, treat GoScript as the orchestration and UI layer, and keep the ultra-hot simulation or consensus core in Rust.
