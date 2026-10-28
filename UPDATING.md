# Updating for a new WoW release

Generation and publication are manual maintainer operations from a source checkout.
[`.wow-version`](.wow-version) is the only WoW version setting. Change its
`major.minor.patch.build` value, then run:

```sh
make generate
```

Use Linux or macOS with Go 1.27+, `zstd`, Git and stock Teal 0.24.8
(`luarocks install tl 0.24.8`), found through `TL` or `tl` on `PATH`. Any other
`tl --version` output fails closed. Acquisition and generation use Go without
external module dependencies.

Generation acquires the selected client, prepares an isolated candidate, renders
declarations and native call permissions, runs consumer checks, and repeats
generation to verify determinism. Only a verified candidate replaces maintained
outputs. Source locks, package hashes, README target, example TOC and recipe follow
the selected version. Package and compiler versions remain separate.

Annotation revisions stay pinned until explicitly changed. Concurrent updates and
edits during a run are rejected. Exceptions and interrupts roll back changed files;
a journal supports recovery of interrupted publication. This does not guarantee
power-loss atomicity.

## Failed generation

The command reports the failed operation and retains its candidate and logs under
`build/`. Existing declarations remain usable. Resolve the source or contract issue
and rerun generation. Unsupported native fields, malformed metadata, missing files
and incompatible declarations fail before publication. The generator does not use
cross-source reports, runtime captures or approval ledgers.

## Source validation

[`client_sources.json`](internal/wowapi/inputs/client_sources.json) records the client UI
inventory, configuration identities, content keys, SHA-256 hashes and sizes.
Acquisition verifies every file before publishing a snapshot. Original Lua, XML,
TOC and listed asset bytes remain in that snapshot; Blizzard inputs have no
source-mirror fallback. Snapshots recorded with an `extractor` field predate the
current lock format and are rejected; remove the `extractor` block, or delete the
snapshot directory to extract again. [`sources.json`](internal/wowapi/inputs/sources.json) owns the
selected generator inputs and pinned annotation revisions.

Literal parsing rejects unsupported statements and duplicate keys. Native decoding
rejects unsupported fields and value shapes, including permission metadata.
Malformed recognized annotations fail with their source location. Bootstrap stores
native projections under `build/`, keyed by the client lock and selected source
representation; it does not copy them into a second native cache.
The [declaration contract](docs/api/teal-catalogue.md) describes projection limits
and the secret-value model.

## XML validation boundary

The extracted `Interface/AddOns/Blizzard_SharedXML/UI.xsd` is versioned with the
client. For build 70575, libxml2 validates all 1,085 UI XML documents against it;
eight `Bindings` documents require a separate contract. XML schema validation
does not establish Lua API signatures or secret-value permissions.

XML validation is not currently a generation gate. Keep the CDN extractor
independent of the XML validator.

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

`bin/wow-types bootstrap` reacquires locked current inputs.
`bin/wow-types verify-client /path/to/snapshot` verifies an existing snapshot.
`bin/wow-types stage-release` prepares a candidate for explicit annotation revision
updates; inspect its source and declaration changes before publishing with
`generate`. It does not publish the active release.

## Verify and package

After `make generate` succeeds, run the additional maintainer checks above, then:

```sh
bin/wow-teal rockspecs --check
bin/wow-teal package
```

A package/compiler update starts in `manifest.json` and requires its corresponding
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
tag and declare the MIT license from `manifest.json`. The source checkout and
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

Run generation, the additional maintainer checks and recipe lint locally before
publication. Publish from a clean checkout of the matching package version tag
after that tag is available in the public repository:

```sh
version="$(bin/wow-teal version)" &&
test -z "$(git status --porcelain)" &&
test "$(git rev-parse HEAD)" = "$(git rev-parse "refs/tags/v$version^{commit}")" &&
luarocks upload "build/rockspecs/wow-tl-type-$version-1.rockspec" \
  --api-key "${LUAROCKS_API_KEY:?Set your LuaRocks API key}"
```

[Consumer CI](.github/workflows/check.yml) validates the committed snapshot with
checks, the Go test suites (generator parity is skipped without a client
snapshot), the example build, temporary archive creation and recipe lint.
Catalogue generation and registry publication remain manual maintainer commands.

## Maintenance boundaries

The [architecture entry point](docs/architecture.md) records capability ownership,
input capture, declaration generation and publication recovery. Only `generate`
and `recover` publish the active release cohort. `check-generator` compares regenerated artifacts and metadata, then runs the
consumer declaration check without changing active outputs. Removed standalone metadata,
catalogue, coverage and template-refresh commands have no replacement writer.

API reports should include client/package/compiler versions, expected and
observed behavior, and a minimal public-input example. Keep private SavedVariables
and inaccessible values out of reports.
