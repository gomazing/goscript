# GoScript Architecture Principles

GoScript is a language and runtime for the next compute era.
It is not designed around the legacy web stack as the center of gravity.

## Core Positioning

GoScript sits at **8/10** on the language maturity scale:

- `0` = primitive language
- `5` = advanced language
- `8` = framework-ready language
- `10` = full-fledged framework

At `8`, GoScript should feel batteries included for app creation, while still staying a language and runtime rather than absorbing framework policy into the core.

GoScript should be able to run across:

- browsers
- desktop apps
- operating-system modules
- IoT devices
- VR and spatial apps
- agentic AI systems

The browser is one target, not the target.

The core design should prioritize modern hardware and execution models:

- DGX Spark-class systems
- Apple M5 and above
- NPU-oriented workloads
- TPU-oriented workloads
- multi-threaded accelerator-backed execution

## API vs TALK

API and TALK are not the same thing.

### API

API is discipline.

It defines how communication should happen when the contract must be predictable:

- request and response shape
- typed boundaries
- serialized payloads
- compatibility rules
- stable interfaces

API is the right model when the system needs clarity, governance, and compatibility.

### TALK

TALK is freedom.

It defines how communication can happen when the connection itself should stay alive, parallel, and expressive:

- per-thread communication
- per-session communication
- per-connection communication
- multi-stream transport
- bidirectional exchange
- swarm-style interaction
- sensor, media, text, and data flow

TALK is the right model when the system needs freedom, parallelism, and adaptive transport behavior.

## Principle Summary

- API is the contract.
- TALK is the transport freedom.
- GoScript should support both natively.
- GoScript should ship batteries-included primitives so vibe coders can build apps quickly.
- GoScript should not be shaped around the legacy ecosystem first.
- GoScript should be built for the hardware and execution models of the AI era.

## What This Means for the Language

GoScript should optimize for:

- native multi-threaded execution
- low-latency transport paths
- parallel streams by default
- sensor and media handling
- agentic workflows
- hardware acceleration
- runtime primitives that work across browser, desktop, and system modules

GoScript should not optimize itself around:

- framework lock-in
- legacy JavaScript-style assumptions
- browser-only thinking
- monolithic app shells

## Practical Rule

If something is about correctness, compatibility, or a stable contract, it belongs in API space.

If something is about live, parallel, adaptive communication, it belongs in TALK space.

If something is about execution speed, layout, or transport latency on modern hardware, the implementation should be tuned for the target platform first.
