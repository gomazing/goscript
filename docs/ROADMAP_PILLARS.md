# GoScript Roadmap Pillars

These are the names to use when discussing GoScript's next major performance and architecture moves.

## Go FAST

Go FAST is the performance and compiler pillar.

It covers:

- Compile-time component lowering
- Faster routing and middleware paths
- Arena-backed render/runtime hot paths
- Lower-allocation SSR and state updates
- Structural optimizations that reduce parse-time work

Use this name when talking about:

- speed
- efficiency
- compiler work
- render throughput
- runtime memory pressure

## Go PAINT

Go PAINT is the rendering and spatial UI pillar.

It covers:

- 2D canvas rendering
- 3D scene composition
- Pixel-position plotting
- Coordinate vectors and collision maps
- Hybrid UI surfaces that mix widgets and scenes

Use this name when talking about:

- canvas-first interfaces
- WebGPU-style composition
- Figma-like layout systems
- DOM bypass ideas
- visual apps and spatial products

## Go IRT

Go IRT means In Real Time.

It is the realtime, streaming, sync, and modular event pillar.

It covers:

- live streams and subscriptions
- event hubs
- scheduler-driven background work
- collaborative state sync
- Swarm-style module communication

Use this name when talking about:

- realtime dashboards
- collaborative apps
- modular distributed systems
- event-driven ERP flows
- low-latency application behavior

## Go TALK

Go TALK is the protocol compatibility and service-contract pillar.

It covers:

- bidirectional frames
- text, data, media, stream, and sensor payloads
- typed request and response contracts
- codec negotiation for text, media, sensor, and stream payloads
- protocol bridges for HTTP, external transports, and future TALK protocol layers
- compatibility for hyper-performant multi-threaded data exchange

Use this name when talking about:

- external protocol compatibility
- service interfaces
- bidirectional transport
- agentic multi-modal exchanges
- sensor and stream ingress/egress

## Go Jetpack

Go Jetpack is the performance, observability, and verification pillar.

It covers:

- profiling
- benchmarking
- tracing
- memory diagnostics
- security and performance panels
- developer feedback loops

Use this name when talking about:

- measurement
- observability
- regressions
- tuning
- confidence in the platform

## How the pillars fit together

- Go FAST makes GoScript fast.
- Go PAINT makes GoScript visually radical.
- Go IRT makes GoScript feel alive in real time.
- Go TALK makes GoScript compatible with protocol-grade transport and service contracts.
- Go Jetpack makes the platform measurable and trustworthy.

If a feature improves speed, call it Go FAST.
If it improves the visual model, call it Go PAINT.
If it improves realtime behavior, call it Go IRT.
If it improves protocol compatibility or service contracts, call it Go TALK.
If it improves diagnostics or proof, call it Go Jetpack.
