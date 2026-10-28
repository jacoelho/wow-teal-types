# Architecture

This is the architecture entry point. [README](../README.md) describes consumer
usage, [UPDATING](../UPDATING.md) describes release operations, and the
[declaration contract](api/teal-catalogue.md) owns projection semantics and limits.
Maintainer commands and their source inputs require the source checkout; the
consumer archive contains only consumer commands and tests.

## Capability ownership

| Capability | Owner | Contract |
| --- | --- | --- |
| Acquire and publish a release cohort | `internal/release` | `Update` prepares and verifies a candidate before publishing; `Recover` resolves interrupted publication. These are the only active cohort writers. |
| Project declarations | `internal/wowapi` | Validate native metadata, apply explicit corrections, finalize a typed declaration graph, and render declarations and call permissions. |
| Check and build an addon | `internal/consumer` | Capture bounded inputs, verify package bytes, admit the compiler by exact version, run all compiler phases against that capture, then publish checked Lua. |
| Extract the selected client | `tools/cdn` | Independent Go module; verified acquisition feeds the release boundary. |
| Validate locked source identities | `internal/sourcelock` | Concrete transport schemas and path validation shared by acquisition and interpretation. |
| Enforce secret operations | `tools/check_secrets.lua` | Runs inside stock Teal 0.24.8; applies the `Secret` marker rules and canonical native call permissions from the finalized projection. |
| Reject native runtime imports | `tools/check_runtime_imports.lua` | Inspect generated Lua before publication. |

CLI packages translate arguments and pass explicit dependencies. Release
metadata preparation returns artifact bytes; it cannot publish them. Candidate
writing and read-only comparison are separate operations. Recipe rendering is
consumer-owned, while its maintained template belongs to the release cohort.
Standalone metadata, catalogue, coverage and template-refresh writers are removed.

The CDN command creates output and cache directories, resolves symlinks, and
compares directory identities and ancestors before acquisition. This rejects
overlap under the filesystem's own case and alias rules. Rejected requests may
leave empty directories; extraction and cache file writes have not started.
Release owns the UI manifest selection (path to file data ID) and passes it to
the extractor in `config.json` alongside the client identity; the extractor
validates and follows it and holds no copy. The CDN is the only acquisition
source. Extraction receipts carry client identity and verified files. Archive
checksums and size-hint validation belong to extraction, without a separate
reporting model. Release bounds extraction time around the extractor process.

## Captured compiler inputs

A consumer invocation owns a private directory until completion. It captures the
package manifest and verifies each copied package file against that manifest.
The capture includes the checker and generated native call policy. The compiler
is stock Teal from `TL` or `PATH`, admitted after package verification only when
`tl --version` prints exactly `0.24.8`; it is pinned by version, not by hash.

Bundling an unmodified Teal was rejected: it duplicates the LuaRocks
distribution and grows the archive without adding a guarantee the version pin
lacks. Treating every `any` as opaque was rejected because most untyped inputs
are not secret, so it would be untrue to the sources.
Selected sources and admitted `.tl`/`.lua` dependencies share one mirrored physical
namespace, preserving identity for overlapping include roots. Teal retains control
of module suffix and include-root precedence. Native declarations precede caller
roots. Generated files live outside every compiler search root.

Direct source/include symlinks and symlinks encountered in admitted trees are
rejected. Existing ancestor aliases, such as macOS `/var`, are resolved once at
the input boundary. Excluded subtrees are absent from the capture unless explicitly
selected or supplied as a physical `-I` root. No symlinks are needed to implement
resolution; copying makes the validated bytes the bytes used by every phase.

The capture bounds caller inputs to 512 selected sources, 2 MiB per source or
dependency, 16 MiB total and 10,000 dependency traversal entries. Package capture
has a separate 16 MiB bound. Each compiler process has a 60-second deadline and
inherits cancellation. Temporary inputs are removed on success or failure.

Edits after capture affect the next invocation. Capture is not a filesystem-wide
atomic snapshot or protection against hostile concurrent mutation. The build
receipt records the captured manifest hash. Compiler diagnostics map captured
source paths back to the caller's paths.

## Publication and recovery

Consumer output belongs to its receipt. A build rejects foreign or edited output,
stages a complete replacement, and restores the previous directory on install
failure. If restoration also fails, it retains the backup and reports its path
and both errors.
Callers serialize publication to the same destination.

Release publication owns a larger journaled cohort transaction and its recovery
command. Its journal and input fingerprints remain distinct from consumer
directory replacement. Neither operation claims power-loss atomicity.
Recovery and extractor building do not require `.wow-version`. Explicit snapshot
verification uses the supplied snapshot; the default snapshot path,
`build/wow-update/snapshots/<version>/<extractor source sha256>`, is keyed by
the client version and the current `tools/cdn` source, so an extractor change
extracts afresh instead of reusing older output.

Snapshot records and the client lock identify the client and its verified files
only: identity, manifests, and each file's content key, SHA-256 and size.
Recording the extractor source hash in the lock was rejected: it added no
integrity beyond the per-file hashes, and because lock equality compared it,
any `tools/cdn` edit made re-extraction of an unchanged build fail.

## Declaration generation

Maintainers run generation and registry publication manually from the source
checkout, following [UPDATING](../UPDATING.md). GitHub Actions validates the
committed consumer snapshot and its packaging contracts.

Locked client and annotation sources feed `internal/wowapi`. Its closed literal
parser validates source syntax without arbitrary Go payloads. Native decoding
rejects unsupported fields, shapes and permissions before constructing concrete
models. Recognized malformed annotations fail with source locations; unknown tags
and unrelated Lua statements are ignored. Explicit overrides own the corrections
missing from those sources.

Release bootstrap supplies native input roots from a verified client snapshot.
Its immutable projection cache is keyed by both the snapshot lock and selected
source representation. No ordinary native mirror is maintained. Only consumed
native documents, version metadata and Questie annotations are selected; the full
client inventory remains verified independently.

Native policy owns presence-aware Boolean flags under one closed classification.
Absent and explicitly false flags remain distinct, and cloned contracts own their
flag maps. String and reference-valued permissions retain concrete types.

Interpretation builds a mutable catalogue; finalization resolves graph closure
into concrete fields, methods and signatures. Rendering consumes that finalized
representation without changing it and emits three declaration files plus the
native call policy. Typed passes preserve declaration ordering and record each
function's compiler origin while emitting its signature. Embedded CR/LF in a
declaration is rejected so physical line counts remain exact. Secure aliases retain
the originating declaration's policy. Full generation runs twice to verify
determinism. Read-only generator checks compare declarations and package metadata
before invoking the same consumer declaration check used by release candidates.

Ingestion retains declaration facts and counts, discarding source documents after
validation and predicate extraction. Duplicate detection and inheritance expansion
own their bookkeeping locally. Finalized declarations own their slices independently
of the mutable catalogue. Closure compares deterministic JSON state for at most 32
passes; historical encoding is reserved for generated names and emitted metadata.

Source locks own acquisition identities; `manifest.json` owns consumer file hashes.
There is no separate release provenance file, report manifest, cross-source audit,
runtime probe or approval ledger. Generation verifies its inputs and compiler
contracts; it does not independently corroborate every API against a running client.

A mutable map renderer was rejected because rendering must not change the graph.
Reparsing rendered declarations to recover policy locations was removed because
emission already owns their final positions. A second native cache and separate
compiler runner were removed to avoid competing source and compiler contracts.
A shared filesystem framework was rejected because local rename injection suffices
to test the actual recovery boundary. Audit reports and approval fingerprints were
removed rather than retained as a second source of generation authority.
