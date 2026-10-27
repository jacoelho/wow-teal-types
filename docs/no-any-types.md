# Removing erased types

Status: complete. Owned Go runtime models use concrete types; generated and
active probe Teal use opaque `unknown` without erased `any`. All 6,440 native
contracts remain accounted for. Repository, compiler, probe and extracted archive
checks passed. Advisor and critic reviews have no unresolved material findings.

The [contract review](../evidence/reviews/no-any-2026-10-27/REVIEW.md) records the
exact baseline, intentional declaration changes, verification and precision limits.

## Required outcome

Remove erased dynamic values from owned Go models and Teal types while preserving
the complete API surface. This includes generated declarations, active probe code,
consumer tooling, release tooling and generator/audit models. Replacing `any` with
an empty-interface alias, a catch-all carrier, an unchecked cast, a bare erased
function or invented generic result does not satisfy the requirement.

The user selected full API coverage, including investigation of compiler/type-system
changes where the pinned compiler cannot express safe contracts. APIs must not be
removed merely because their current declaration contains `any`.

Historical source documents, captures and review receipts remain immutable evidence.
Their literal text may describe `any`; that does not permit publishing erased
current declarations. Intentional invalid-input test strings must remain separate
from accepted code. Go generic constraints that retain each caller's concrete type
are distinct from runtime erased values. Standard-library interfaces do not justify
local dynamic domain models.

## Findings

The initial audit of commit `f630d20` found 5,399 `any` tokens across generated Teal
modules and 140 in active probe Teal code. Environment files duplicate public
signatures. The generated manifest records 781 missing-parameter fallbacks and
1,331 Wiki stubs without result annotations.

Teal 0.24.8 has no source-level safe `unknown` type. Its internal lax-mode unknown
is not a usable substitute. Empty records, userdata markers and recursive Lua-value
unions do not accept every supported runtime shape. Independent result type
parameters can invent a caller-selected result and are unsound.

Some source contracts support honest generics, including table key/value
relationships. Others need compiler support: `MeasureCall` forwards arbitrary
argument/result tuples, and `dumpobject` may return a metamethod's arbitrary result.
Missing annotations still require evidence; type-system features cannot infer a
contract that is absent from the sources.

## Ownership and migration

1. Consumer owns concrete package manifests, build receipts, archive results and
   recipe descriptions. Strict boundary decoders reject malformed shapes rather
   than returning zero values. Capture continues to bind exact manifest bytes.
2. Release owns concrete acquisition, publication and command result models and
   consumes the consumer manifest contract. Shared locked-source contracts need
   one authoritative owner and acyclic dependencies.
3. Wowapi owns a closed literal representation for arbitrary raw source evidence.
   It distinguishes null, booleans, strings, integers, floating numbers, arrays,
   objects and symbolic Lua references/expressions. It permits no arbitrary Go
   payload. Raw evidence remains beside concrete interpreted models.
4. Native documents, annotations, catalogue definitions, policies and audit results
   become concrete models. A generic value map throughout domain logic would
   retain the existing ownership and validation problem.
5. Extend the pinned Teal 0.24.8 compiler with a distinct public `unknown` type,
   safe assignment and narrowing, and matching secret-checker support. Preserve
   Lua 5.1 generation. Unknown values remain opaque until evidence permits an
   operation; they do not imply publicly inspectable values. Preserve known arity
   and parameter/result positions. Heterogeneous type packs are a separate
   precision feature, not required for complete symbol coverage. Do not invent
   universal result types for undocumented external results. Compiler changes
   require their own regression evidence before adoption.
6. Replace probe reflection with typed owned resources and case-specific calls;
   preserve the observation schema and meaning.
7. Regenerate through the release owner. Changed declarations require explicit new
   review evidence; historical graph hashes and approvals cannot be relabeled.

Representation-only Go changes must preserve artifact bytes, including canonical
JSON ordering, integer/float distinctions, wide integers, null versus absence,
UTF-8 behavior and interning fingerprints. Deliberate Teal contract changes are a
separate review from that migration.

## Completion evidence

Completion requires all of the following:

- Semantic inspection of owned Go types and Teal declarations finds no erased
  `any` or equivalent escape hatches; a text search alone is insufficient.
- Concrete boundary tests reject malformed/null/trailing inputs and unsupported
  source shapes while retaining raw evidence for diagnosis.
- Compiler tests prove unknown values cannot be used as concrete values without
  evidence, including containers, callbacks, generics, narrowing and secret values.
- The full API inventory remains accounted for with no silent exclusions.
- Generator, metadata, source/permission audits, consumer and probe checks pass
  against explicitly reviewed new artifacts.
- Extracted consumer archives run independently with the supported compiler.
- Advisor and critic reviews have no unresolved material findings.

## Compiler decision and rejected alternatives

The pinned compiler extension and secret checker form one versioned contract.
The research prototype is not suitable for adoption: callback bivariance and
mutable-container covariance permit type corruption, and loop/builtin inspection
still bypasses its initial restrictions. Function parameters require contravariant
compatibility, results covariant compatibility, and mutable containers invariant
compatibility where widening would permit invalid writes. Primitive narrowing
must not erase accessibility or secrecy obligations. Integer narrowing cannot
emit `math.type` in the supported Lua 5.1 runtime.

Classifying unknown as potentially secret is necessary but insufficient: opaque
values must not escape into inspecting legacy standard-library `any` parameters
such as `print`. Preserve native `SecretArguments` policy at unknown slots and
require a checked contract for operations on opaque values.

Upstream main was inspected at `28b9e8148adf3aaec029c753c96034adcfd86c6b`.
It provides neither public unknown nor packs, and removes reporter entry points
used by the checker. Upgrading alone therefore adds integration work without
meeting the requirement. Reusing internal lax-mode unknown, replacing unresolved
values with invented userdata records, and emitting independent universal result
type parameters were rejected because they assert unsupported contracts.

Trusted declarations can legitimately contain result-only generics, for example
fresh generic containers. The prohibition concerns invented external promises;
it is not a blanket syntax ban on all result-only type parameters.

## Distribution and call permissions

The consumer package includes compiler `0.24.8-wow.1`, its CLI dependencies and
capability preflight. Checks, builds and probe generation use that compiler by
default. An explicit override must satisfy the same version and capability
contract. Package capture preserves the executable launcher and hashes every
runtime input. LuaJIT is the default host; `LUA` selects another Lua 5.1 interpreter.
The declaration-only LuaRocks recipe requires the matching toolkit and does not
claim an unpublished compiler dependency exists in the registry.

The finalized declaration projection owns native argument secrecy permissions.
Rendering emits `types/wow/call_policy.lua`, keyed by declaration path and compiler
function-origin coordinates, with one entry per overload. Argument indices include
receivers and expanded ABI slots. Canonical guard identities are recorded there.
The checker consumes this hash-bound artifact; names alone cannot confer guard or
sink privileges. Missing or unsupported policy/report capabilities fail closed.

## Probe observation boundary

Presence lookup and typed invocation are separate operations. A compiler-owned
read-only field observation accepts a statically known table or userdata receiver
and a string key, and returns `unknown`. It preserves ordinary `__index` behavior,
evaluates operands once and grants neither mutation nor invocation authority.
The probe supplies a finite candidate list and performs the lookup in a protected
closure with caller-local receiver checks. Returned values need their own checks.

A runtime type tag does not establish a record or callable shape. Tag observation
and equality on unknown values require accessibility and publicness for the same
operands. Equality retains Lua metamethod semantics and protected execution.
Generic bodies must enforce these obligations under their declared bounds;
unconstrained parameters cannot conceal unknown values from operation checks.

The recorded presence capture reports UI query methods absent on binding, clock,
locale and signal-map resources. Only radial and vector resources have typed query
contracts. An undeclared query discovered by future presence lookup is recorded as
`unverified_signature` and is not invoked. This replaces the test fixture's invented
callable contracts without dropping candidate observations.

## Native input contents

`SecretArguments = AllowedWhenTainted` permits an argument's opaque root unless
`NeverSecret` forbids it. It does not establish permission for recursively secret
contents. The current source cohort supplies no independently established input
contents permission, so generated permissions for those contents remain empty.
`ConditionalSecretContents` describes a return in this cohort and cannot authorize
input forwarding. `NeverSecretContents` is an explicit negative input constraint.

This conservative limitation does not claim the native runtime rejects all such
inputs. Recovery requires evidence for the specific symbol, argument position and
secret-input context. Existing public-fixture observations establish no such
permission. Concrete inputs without potentially secret contents remain usable.

## Native receiver placement

The pinned Wowless documentation routes five `SimpleRegionAPI` methods to
`LayeredRegion`: `GetDrawLayer`, `SetDrawLayer`, `GetVertexColor`, `SetVertexColor`
and `SetVertexColorFromBoolean`. `nativeMethodOwner` owns this source-to-emitted
placement rule. Loading, audit lookup, coverage and policy rendering use that
owner while retaining the original native document and independent TSM lookup.

`Region` contains common capabilities. `LayeredRegion` extends it with draw-layer
and vertex-color methods. `Frame` keeps common Region, animation and script
capabilities; `ModelScene` keeps its distinct one-argument layer setter. FontString,
TextureBase and VectorGraphics retain their layered capabilities. This corrects
an invalid inherited contract exposed by mutable-field invariance; widening the
native setter or relaxing compiler variance would conceal the source mismatch.

The correction is supported by pinned Wowless method placement and hierarchy,
not a new live-client observation. Historical Region-labelled probe IDs remain
unchanged, and their method inventory is obtained from the corrected layered
capability. The new review records the five owner moves explicitly; historical
source documents, captures and approval bytes are retained.
