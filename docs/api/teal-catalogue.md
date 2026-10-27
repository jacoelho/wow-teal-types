# Declaration contract

The generated `types/wow/wow.d.tl` exports native handles, widget interfaces, enums,
structures, aliases, callbacks and signatures. `wow_env.d.tl` supplies public
globals; `wow_secure_env.d.tl` supplies the separately selected secure environment.
`native_values.d.tl` owns the opaque scalar identities re-exported by `wow`.
The Go generator and evidence checks live in [`internal/wowapi`](../../internal/wowapi);
[`internal/release`](../../internal/release) owns verified acquisition, isolated
candidates and publication. [`internal/consumer`](../../internal/consumer) owns
package integrity, Teal/secret checks, checked addon builds and packaging.
[`scripts/wow_api`](../../scripts/wow_api) retains reviewed data locks and the
independent Rust reader. `bin/wow-types` is the maintainer command;
`bin/wow-teal` is the consumer command. Both launch Go without changing addon
path interpretation. Input roots are explicit per generation, with no shared
process-global source selection.
The full source and comparison reports live in [`evidence/generated`](../../evidence/generated);
the consumer ZIP contains declarations and checks only.

[`.wow-version`](../../.wow-version) owns the manually selected WoW release.
Acquisition, staging, generated metadata and the probe derive the version, build
and interface from that one value. Source locks remain independent acquired
evidence and must match it. `make generate` acquires and validates an isolated candidate, including full
metadata preservation and consumer type/secret checks, before refreshing derived
provenance, package metadata, documentation and the example TOC;
historical captures and contract-review decisions are never relabeled.

Blizzard generated metadata selects native arity, ordered returns, defaults,
absence and restrictions. The maintainer workflow obtains the complete client UI
inventory directly from CASC/CDN, verifies its exact build and content identities,
then projects the selected documentation and implementation files into the
existing conversion model. Missing or modified Blizzard files fail without a
source-mirror fallback. Original packaged bytes remain intact. Legacy Git source
identities survive staging only when the entire native/support/version cohort
and full native inventory match their reviewed hashes. Otherwise all three groups
record explicit CASC identity: snapshot-lock SHA-256, packaged paths and LF
normalization with exactly one final newline. The independent Gethe cross-check
retains its observed build; a lagging mirror is a review blocker, never a reason
to relabel the primary client source.

Questie annotations supplement missing surfaces;
reviewed FrameXML additions and representation corrections remain part of the
exported cohort. Exact revisions are in [provenance](../../provenance.json).
TSM, Wowless, the separate Blizzard distribution and native observations were
cross-check evidence, not permission to relax native restrictions.

The [client coverage report](../../evidence/generated/wow_client_coverage.json)
enumerates every maintained type export, native function/event contract, retained
annotation signature and third-party library source. Its native document
inventory must match the complete selected documentation directory in the client
snapshot. Export-origin categories identify recorded inputs rather than automatic
inference percentages: a type created by an annotation may also contain native
methods and inherit native capabilities.

Packaged Lua/XML does not determine every parameter type, result tuple, optional
position, generic relation or secrecy effect. Third-party libraries, legacy
supplements, reviewed FrameXML assertions and opaque-value policy remain explicit
inputs. Accounting for all current exports does not claim that every client
global or mixin has been inferred. No missing contract is converted to `any` merely
to reach a coverage count.

Opaque markers describe a static restriction rather than a runtime userdata
carrier. Inspect a value only after same-caller native accessibility checks.
Returning an ordinary scalar additionally requires negative secrecy and a valid
scalar shape. Outer-table access does not declassify its fields; public handles
can retain secret state. Forward supported opaque values to native sinks.
Never use a cast or a runtime userdata test to remove these obligations.

The checker follows canonical marker identities through aliases and unions.
It rejects inspections lacking the required proof and validates the actual loaded
`wow`, `native_values` and native environment files against the packaged paths.
The full historical operation policy remains in the checker. Values erased
through `any`, arbitrary Lua, unselected source bodies and runtime permissions
remain outside that static claim.

Teal permits nil for every type; retain runtime absence checks. Source-untyped
boundaries retain explicit `any`. Some unions, factories and repeated tuples have
representation limits. Optional libraries and legacy supplements are not a
runtime availability guarantee. Secure declarations grant no protected access.

See [addon usage](../../README.md) and the
[release update guide](../../UPDATING.md).

The repository-owned Go CDN extractor has no external module dependencies. Its
config, archive-index, encoding, root and BLTE readers validate format-defined
integrity and bounded reads; unsupported formats or encryption fail explicitly.
The original C# extraction and the Go cold acquisition matched all 4,029 files
for the recorded migration cohort. This establishes that cohort, not universal
future-format support.

The [full native metadata report](../../evidence/generated/wow_native_metadata.json)
preserves every parsed document field. Unknown fields and shapes are recorded
before publication is refused. Raw source remains authoritative for comments,
lexical spelling and registration statements. Source-equivalence receipts may
transfer only byte-equivalent source approval; every contract, emitted graph and
build-specific evidence gate remains in force.

The Python-to-Go tooling migration preserves the reviewed declarations and all
four generated reports byte for byte. Generated headers retain their historical
generator path because those bytes belong to the approved declaration graph.
The original TSM ledger and captured runtime evidence are unchanged. Canonical
review JSON, default-spaced interning JSON, float spelling and native metadata's
UTF-8 representation remain separate serialization contracts. Parser tests cover
wide integers, invalid UTF-8, duplicate keys and unsupported statements; complete
cohort and negative permission checks exercise the actual Go converter and reader.
The source-only consumer archive carries its complete Go source/module closure;
it requires neither producer packages nor Rust or acquisition caches.
