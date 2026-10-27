# Updating for a new WoW release

[`.wow-version`](.wow-version) is the only WoW version setting. Change its
`major.minor.patch.build` value, then run:

```sh
make generate
```

Use the source checkout on Linux or macOS. Install Go 1.27+, Rust/Cargo, LuaJIT,
`zstd` and Git first. The pinned compiler is bundled under `tools/teal`. Acquisition, generation and consumer tooling run in Go; they need no Python
or virtual environment. Reproducing the compiler patch separately uses Python 3
and `patch`, as described in [the compiler contract](tools/teal/README.md). The generator uses a pinned Go
YAML dependency with explicit node validation. The CDN extractor has a separate,
standard-library-only Go module. The pinned Rust reader remains an independent
TSM cross-check. No extraction repository checkout or running WoW client is needed.

The command acquires the exact selected client, prepares an isolated candidate,
builds the pinned reader, generates declarations and full metadata, checks source
coverage, runs Teal and secret checks on the consumer example, and verifies
repeatable generation. It replaces the maintained outputs only after every check
passes. Source locks, provenance, package fingerprints, README target, example TOC
and recipe follow the selected version automatically. Package and compiler
versions remain separate from the WoW version. The consumer owns the compiler
version; release metadata records it in `package.json`.

Annotation and cross-check revisions stay pinned. A WoW version change does not
silently select newer third-party code or relabel older evidence. Concurrent
updates and edits made during a run are rejected. Ordinary exceptions and
interrupts roll back changed output files; a publication journal records an
unfinished transaction. This is not a claim of power-loss atomicity.

## When generation stops

The command reports the failed gate and retains its candidate, source inventory,
metadata and logs. Existing declarations and checks remain usable. Resolve the
reported contract or source issue, then rerun the same command; do not edit hash
ledgers merely to silence a failure.

A source-equivalence receipt can carry existing source approval only when the
complete native inventory and source bytes match after the documented newline
normalization, support/annotation/cross-check inputs are unchanged, and the
original contract fingerprints, discrepancy resolutions and complete emitted
graph still match. The original review ledger is not rewritten. Runtime
observations remain attached to their actual build; loss of required build-specific
corroboration still stops generation.

Changed permissions, secret metadata, unknown fields, unsupported formats,
missing files and weakened or unsupported projections are failures requiring
attention. A new release is not automatically approved merely because its Teal
signatures happen to look unchanged.

## Source and information preservation

[`client_sources.json`](scripts/wow_api/client_sources.json) records the complete
UI manifest inventory, client/configuration identities, original content keys,
SHA-256 hashes and sizes. Acquisition verifies every file before publishing a
snapshot. Original Lua, XML, TOC and listed asset bytes remain in the snapshot;
Blizzard inputs have no source-mirror fallback.

[`wow_native_metadata.json`](evidence/generated/wow_native_metadata.json) retains
all parsed native document fields, including headers, ordered results, false
values, absence, callbacks, permissions and secrecy conditions. A frozen
[`metadata_fields.json`](scripts/wow_api/metadata_fields.json) detects unknown
fields or value shapes; those values are retained in the candidate report before
the update stops. Literal parsing rejects ignored statements and duplicate keys.

The [coverage report](evidence/generated/wow_client_coverage.json) accounts for
every maintained export and native contract. The
[declaration contract](docs/api/teal-catalogue.md) records projection limits:
Teal declarations cannot express every source distinction. Raw bytes and complete
metadata remain the evidence; an unexpressible distinction is not silently
replaced with a broader type to make generation pass.

## XML validation boundary

The extracted `Interface/AddOns/Blizzard_SharedXML/UI.xsd` is versioned with the
client. For build 70575, libxml2 validates all 1,085 UI XML documents against it;
eight `Bindings` documents require a separate contract. XML schema validation
does not establish Lua API signatures or secret-value permissions.

The published [`jacoelho/xsd` commit 478bcb1](https://github.com/jacoelho/xsd/commit/478bcb1fb01386e535082e6fabf4eb9ce8827193)
cannot compile this schema (`complex extension content does not preserve base
content`). Resolve that incompatibility before adding it as a generation gate;
keep the CDN extractor independent of the XML validator.

## Additional maintainer checks

```sh
make check-generator
make test
make check-package
GOTOOLCHAIN=local GOWORK=off go test -race ./internal/release ./tools/wow-types
GOTOOLCHAIN=local go test -C tools/cdn -race ./...
GOTOOLCHAIN=local go vet -C tools/cdn ./...
./check-all.sh
GOTOOLCHAIN=local go vet ./...
```

`bin/wow-types bootstrap` reacquires the locked current inputs.
`bin/wow-types verify-client /path/to/snapshot` verifies an existing snapshot
without replacing it. `bin/wow-types stage-release` remains available for
explicit annotation/cross-check revision updates; its independent source review
is separate from selecting a WoW version. Historical captures and approved
operation policies are never relabeled by generation.

## Native probe when needed

Build the isolated developer addon with the same strict type and secret checks:

```sh
bin/wow-types probe build
```

This writes `build/generated/wow-api-probe/WoWAPIProbe` and `probe-build.json`;
it never installs anything. Copy that addon directory into the intended client's
`Interface/AddOns`, log in and run `/run C_AddOns.LoadAddOn("WoWAPIProbe")`.
Use `/api-probe capture` for presence or `/api-probe invoke` for public-fixture
invocations. `/api-probe status` shows the latest result.
Reload or log out to flush SavedVariables, then import them without executing Lua:

```sh
bin/wow-types capture /path/to/SavedVariables/WoWAPIProbe.lua \
  --manifest build/generated/wow-api-probe/probe-build.json \
  --receipt /path/to/installed-receipt.json \
  --output references/wow-api/native-observations/BUILD/CAPTURE
```

The installation receipt must record `build_manifest_sha256` for the exact
installed manifest and a `files` map of its runtime paths to SHA-256 hashes,
verified against the installed addon bytes. Keep it and the manifest with each
capture; merely copying manifest hashes does not verify installation.
Public fixture success does not
certify secret-context permissions or complete signatures. Existing observations
and hash-bound behavior reviews remain immutable unless explicitly superseded.

## Verify and package

After `make generate` succeeds, run the additional maintainer checks above, then:

```sh
bin/wow-teal rockspecs --check
bin/wow-teal package
```

A package/compiler update starts in `package.json` and requires its corresponding
checker/CI review. Remove the superseded recipe template before a package version
bump; generation creates its replacement. Verify the consumer ZIP from a separate
directory. Packaging does not install or publish anything.

## LuaRocks and publication

Generate a recipe using the actual public repository identity:

```sh
bin/wow-teal rockspecs --repository OWNER/wow-teal-types
bin/wow-teal rockspecs --repository OWNER/wow-teal-types --check
```

Candidates go to `build/rockspecs/` and refer to the matching `v<package-version>`
tag and declare the MIT license from `package.json`. The source checkout and
consumer ZIP include `LICENSE`; LuaRocks copies the root license with its default
documentation installation. Additional registry description metadata is optional:

```sh
bin/wow-teal rockspecs --repository OWNER/wow-teal-types \
  --description-json "$ROCKSPEC_DESCRIPTION_JSON"
luarocks lint build/rockspecs/wow-tl-type-VERSION-1.rockspec
```

Replace `VERSION` with the package version. The plain recipe can also be linted
without `--description-json`. Optional metadata is a JSON object of additional
string fields, such as `maintainer`; it cannot replace summary, detailed description,
homepage or license.
`--check` detects drift without writes. Generation never uploads or bumps versions.
The rock installs declarations only; use the matching ZIP's `wow-teal` command
for the additional secret-operation checks.

[Generator CI](.github/workflows/generator.yml) acquires locked inputs in a fresh
checkout, builds the pinned reader and checks regeneration, reviews and native
probe contracts. [Consumer CI](.github/workflows/check.yml) runs checks, tests, the example build,
archive generation and recipe lint with Lua 5.1 and pinned Teal.
[Manual generation](.github/workflows/generate-rockspecs.yml) retains a candidate
artifact and runs registry lint; `ROCKSPEC_DESCRIPTION_JSON` is optional.
[Manual upload](.github/workflows/upload-types.yml) defaults to a dry run.

A real upload requires the matching version tag, `LUAROCKS_PUBLISH_ENABLED=true`,
and a `luarocks` environment with `LUAROCKS_API_KEY`.
Configure required reviewers on that environment before enabling uploads. The
upload job requires the reusable generator workflow for the same tagged commit,
consumer checks and registry lint to pass. Ordinary pushes never publish a rock.

## Maintenance boundaries

The [architecture entry point](docs/architecture.md) records capability ownership,
input capture, declaration generation and publication recovery. Only `generate`
and `recover` publish the active release cohort. `check-generator` compares an
isolated candidate without changing active outputs. Removed standalone metadata,
catalogue, coverage and template-refresh commands have no replacement writer.

API reports should include client/package/compiler versions, expected and
observed behavior, and a minimal public-input example. Keep private SavedVariables
and inaccessible values out of reports.
