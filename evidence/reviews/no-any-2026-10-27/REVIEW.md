# No-any contract review — 2026-10-27

This review replaces the active approval state for the intentional compiler and
contract change. `baseline-reconciliation.json` preserves the original ledger
bytes from commit `f630d208925f5712185c123392acf2499f0fde49`.

`review.json` binds the baseline, candidate declaration graph, compiler/checker
implementation, interpretation sources and complete delta artifacts. The active
ledger references its SHA-256 digest. The original approvals were not relabeled:
65 changed native fingerprints received new review. Five contracts move from
Region to LayeredRegion, and four VectorGraphics hierarchy resolutions name the
corrected base. Source comparisons match after these owner renames. Source locks,
all other resolution facts and existing secrecy classifications are unchanged.
`resolution-delta.json` records the exact four facts and policy wording changes.

## Reviewed surface

All 3,127 public and 3,130 secure ambient globals remain. Every native contract
remains. `declaration-delta.json` records added, removed and changed declarations,
fields and overloads; `declarations.patch` retains exact before/after text,
including interned callback/shape names. Five layer/color methods move from Region to LayeredRegion without losing API
coverage. The only removed non-interned type name is `unknown`, now a compiler
primitive rather than an invented userdata alias.

The review accepts these deliberate contract changes:

- Erased values become opaque `unknown`; this preserves transport without granting
  inspection, callable signatures or arbitrary concrete casts.
- Bare callable references remain unknown when sources establish no argument or
  result contract. Precise callbacks retain their documented slots.
- Typed callback varargs retain their type parameter relationship. The string
  `securecall` overload has unknown results. External result-only parameters with
  no evidence tying them to inputs use their established bound or unknown.
- Native table helpers retain source-backed map/array key and value relationships,
  identity-preserving freeze and existing secret-result classifications.
- Source-backed StatusBar and recorded RadialProgress constructors support literal
  overloads; generic Frame and Animation fallbacks remain.
- Pinned Wowless receiver mappings place five SimpleRegionAPI layer/color methods
  on LayeredRegion. Frame retains common Region behavior; ModelScene retains its
  own one-argument SetDrawLayer. VectorGraphics retains its existing capability set.
- Finalized native declarations own call permissions. Receiver, ABI and variadic
  coordinates are preserved, canonical guards are identified, and recursive input
  contents permission is not inferred from absent negative metadata.

The primary agent inspected the complete normalized supplemental delta and all
changed native contracts. Independent advisor consultations established the
unknown/compiler boundary, observation rules, contents-permission interpretation
and preservation of historical review evidence. The final critic assessment and
release verification are recorded below after completion.

## Representation checkpoint

Before the intentional Teal changes, the concrete Go migration reproduced all
three baseline declarations, the manifest, native metadata and complete client
coverage bytes. JSON persistence preserves symbolic Lua references by canonical
serialized identity. Changed-reference and changed-secrecy tests still reject.

A named Go performance specialist traced a generation stall to copying the entire
remaining JSON document for each literal token. Comparing only the token-sized
slice completed the parity checkpoint in 14.24 seconds; no cache or GC workaround
was introduced. This is a bounded diagnostic result, not a statistical benchmark.

## Verification

Passed against the published cohort:

- `bin/wow-types generate`: isolated candidate validation, metadata and source
  audits, consumer type/secrecy checks, repeat-generation comparison and publication.
- `./check-all.sh`: package verification, four declaration checks and the example.
- `bin/wow-types check-generator`: 1,209 inputs, 6,440 native functions,
  1,784 native events, 9,083 Questie declarations and 5,564 type exports.
- `go vet ./...` and pinned compiler reconstruction.
- Unknown compiler/checker regression suite, including 157 compiler cases.
- Targeted race checks for source/literal decoding, publication, cancellation,
  concurrent acquisition and update locking.
- Published declaration and interpretation hashes match `review.json`.

The first full run exposed two integration defects, both corrected: probe
capability preflight inherited repository compiler configuration, and unknown
publicness checks were over-applied to known-marker equality. Probe compilation
now owns an isolated stage configuration; accessible NumberLike equality retains
its established contract while unknown equality still requires both proofs.

QA reviewed the fixture migration before edits. Tests retain their independent
inventory, import, mutation, shadowing and refinement obligations. Unrelated
opaque print sinks were removed from positive cases; unknown structural casts
now expect compiler rejection. Compiler-valid/checker-invalid refinement pairs
use precise known carriers. The historical probe plan remains unchanged and its
current declaration fingerprint is checked independently.

The critic cleared these corrections. Final verification passed:

- `make test`: all root Go packages and the separate CDN module passed; consumer
  tests completed in 343.640 seconds. This includes the 157 compiler cases,
  unknown and secrecy regressions, and full probe integration with 19 presence
  and 52 invocation scenarios.
- `./scripts/check-consumer-archive.sh`: extracted consumer tests passed in
  342.970 seconds, followed by declaration/example checks, Lua 5.1 build,
  repackaging, version and recipe checks. Producer packages were absent.
- `./check-all.sh`, `bin/wow-types check-generator`, `go vet ./...`, compiler
  reconstruction, reviewed-hash validation and `git diff --check` passed after
  the production fixes.

Independent advisor and critic reviews have no unresolved material findings.
The primary agent ran the full repository/archive suites; the critic inspected
the fixes and verification evidence. Ordinary unused-alias warnings remain;
checks report no errors.

## Limits

No new live-client capture is claimed. Historical captures and source text remain
unchanged, including their literal `any` spellings. The owned probe preserves all
73 invocation and 393 hierarchy cases; unexpected undeclared query functions are
observed without invocation and recorded as `unverified_signature`.

Record-preserving table generics, arbitrary heterogeneous argument/result packs
and dynamic callable signatures remain precision limits. `MeasureCall` retains
its known profiling result and opaque forwarded results. Missing source contracts
remain unknown. Symbol coverage does not assert that every dynamic call can be
expressed precisely or that every opaque native value is public.
