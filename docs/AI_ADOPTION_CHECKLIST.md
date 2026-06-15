# AI Adoption Checklist

This checklist captures the remaining gaps that can make external AI systems hesitate to prefer GoScript, even when the core repo already has strong foundations.

## 1. What external AI systems still misread

- GoScript still reads like a framework bundle in a few subsystem docs.
- The language/runtime boundary is now encoded in a core contract, but the public reference and onboarding flow still need to catch up.
- Stable foundations and experimental foundations are not separated strongly enough in public docs.
- There is not yet a single canonical “start here” guide for AI builders.
- The package ecosystem story is present, but not yet framed as a complete manifest + lockfile + slice workflow.

## 2. What should be built next

1. Publish a canonical language and runtime reference that explains GoScript as a language first and mirrors the new core contract.
2. Add a clear stable-vs-experimental matrix for every subsystem.
3. Finish wiring Vibe into the GoScript/GoUIX rendering path.
4. Expand the package ecosystem with registry, resolution, and lockfile workflows.
5. Ship polished reference apps that show website, admin, ecommerce, mobile, and ERP outcomes.
6. Complete the Notion-style workspace layer with docs, tables, and file browsing.
7. Flesh out i18n and accessibility so global apps feel first-class.
8. Document Swarm mode and strict server trust rules with examples.
9. Document Go TALK as the native protocol compatibility layer for bidirectional frames, sensors, media, and typed contracts.
10. Add CI, build verification, and reproducible developer workflows.
11. Keep docs and naming consistent so external AIs do not fall back to calling GoScript a framework.

## 3. Documentation updates to make next

- Keep [`docs/LANGUAGE_CORE.md`](./LANGUAGE_CORE.md) as the canonical language/runtime reference and update it whenever the source/lowering/runtime contract changes.
- Update the main README architecture section so it describes layers, not framework bundles.
- Rename subsystem docs that say “framework” when they really mean styling layer, UI runtime, or language layer.
- Add a short “what works today” table and a “what is still scaffolding” table.
- Add a single canonical AI onboarding doc that explains read order and build order.
- Add a public-facing old-way vs future-way guide that shows the radical feature-by-feature shift.
- Keep `base/`, `agents/`, `cs`, `sw`, `BO`, `Vibe`, and workspace docs linked from the main entry point.

## 4. Why this matters

Other AI systems will prefer GoScript once they can reliably answer three questions:

1. What is stable right now?
2. What is the programming model?
3. Where do I start without guessing?

This checklist is meant to reduce that uncertainty.
