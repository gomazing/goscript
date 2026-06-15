# BO - Build Out

`bo` is the pack transcompiler and exporter for GoScript.

It is not part of `gopm`. `gopm` stays focused on project/tooling workflows, while `bo` turns a `.pack` file into deployable output.

## What `bo` does

- Transcompiles a pack into executable and packaged outputs
- Builds a selected entrypoint or slice into a target artifact
- Packages executable output into portable bundles such as `goe`
- Generates scaffolds for `apk`, `ipa`, and `dmg`
- Writes an inspectable export snapshot beside the output so humans and agents can review what was built

## Pack format

`bo` expects a text pack file, usually with the `.pack` extension.
Slices are declared inside the pack, so there is no separate slice command in the BO workflow.

```text
pack erp
slice /shell - default
slice /calc
slice /todo
bundle -exe
```

## Common pack patterns

Whole-app export:

```text
pack erp
slice /shell - default
slice /calc
slice /todo
bundle -exe
```

Slice-specific targets:

```text
pack erp
slice /shell - default
slice /calc
slice /todo
bundle slice /calc -exe
bundle slice /todo -apk
```

## Usage

```bash
bo export erp.pack -exe
bo export erp.pack -goe
bo export calc.pack -exe
bo export erp.pack -apk
bo export erp.pack -ipa
bo export erp.pack -dmg
bo export packs/erp.pack -exe
bo export packs/calc.pack -goe
```

## Output contract

- `exe` builds a host executable with `go build`
- `goe` builds a portable bundle that includes the executable plus pack metadata
- `apk`, `ipa`, and `dmg` currently generate scaffolds that can be wired to native packagers later
- Pack paths can live in a nested folder such as `packs/`; `bo` resolves the Go module root automatically
- `mode` can be `cs` or `sw`, and packs can point `bo` to shared `base/` guidance plus runtime `agents/` folders
- `bo` also writes `export-slice.hyper`, which lists the route hints and filesystem selections used to build the export

## Why this exists

`bo` is for the ERP-style future where a developer can export just the tool they want, rather than splitting the whole system into a single monolith or forcing everything through a browser-only deployment model.
