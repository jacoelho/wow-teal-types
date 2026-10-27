# Declaration contract

`types/wow/wow.d.tl` exports native handles, widget interfaces, enums, structures,
aliases, callbacks and signatures. `wow_env.d.tl` supplies public globals;
`wow_secure_env.d.tl` supplies the separately selected secure environment.
`native_values.d.tl` owns opaque scalar identities re-exported by `wow`.

The [architecture entry point](../architecture.md) defines ownership.
`internal/wowapi` interprets locked sources and renders declarations;
`internal/release` acquires, validates and publishes candidates;
`internal/consumer` owns package integrity, checks and addon builds.

## Sources and projection

[`.wow-version`](../../.wow-version) selects the client release.
[`sources.json`](../../internal/wowapi/inputs/sources.json) locks generator inputs;
[`client_sources.json`](../../internal/wowapi/inputs/client_sources.json) locks the
complete acquired client UI inventory. Original packaged bytes remain intact.
Missing or modified Blizzard files fail without a source-mirror fallback.

Blizzard metadata selects native arity, ordered returns, defaults, absence and
restrictions. Questie annotations supplement missing surfaces. Explicit
[FrameXML additions](../../internal/wowapi/inputs/framexml.json) and
[contract corrections](../../internal/wowapi/inputs/overrides.json) fill selected gaps.
Native decoding rejects unsupported fields and value shapes before interpretation.

Teal cannot express every source distinction. Packaged Lua/XML does not determine
every parameter type, optional position, generic relation or secrecy effect.
Unresolved positions use `unknown`; missing precision does not justify dropping
symbols, emitting `any`, or inventing concrete results or generic relationships.
The retained surface is not a claim to infer every client global or mixin.

The five draw-layer and vertex-color methods from `SimpleRegionAPI` belong to
`LayeredRegion`. `Frame` retains common `Region` capabilities; `ModelScene` keeps
its distinct one-argument layer setter. FontString, TextureBase and VectorGraphics
retain layered capabilities. This explicit routing avoids incompatible inherited
contracts without weakening mutable-field invariance.

## Compiler and secret values

Checks and builds require the bundled Teal `0.24.8-wow.1` compiler together with
the secrecy checker and generated `call_policy.lua`. The consumer command selects
the bundled compiler by default; an override must satisfy the same version and
capability contract. Stock Teal does not provide this public `unknown` contract.
The [compiler reference](../../tools/teal/README.md) defines the supported
operations, variance rules, Lua 5.1 target and reproduction procedure.

`unknown` transports a value without granting indexing, invocation, arithmetic,
truthiness or conversion to a concrete type. A known function can transport
unknown arguments or results; an unknown value itself is not callable. Primitive
narrowing can establish a number, string, boolean or nil, subject to native
accessibility and secrecy checks. A runtime type tag cannot establish a record or
callable shape. Casts cannot recover an unsupported contract.

Opaque markers describe a static restriction rather than a runtime userdata
carrier. Inspect a value only after same-caller native accessibility checks.
Returning an ordinary scalar additionally requires negative secrecy and a valid
scalar shape. Outer-table access does not declassify its fields; public handles
can retain secret state. Forward opaque values only where the native argument
policy permits them. Permission for an opaque root does not authorize its nested
contents; the current source cohort supplies no independent permission to forward
potentially secret input contents. Never use a cast or a runtime userdata test to
remove these obligations.

The checker follows canonical marker and unknown identities through aliases,
unions and generic bounds. It rejects inspections lacking the required proof and
validates the loaded `wow`, `native_values` and native environment files against
the packaged paths. The finalized declaration projection owns native call
permissions in `call_policy.lua`; the checker uses each originating function's
identity and argument positions, including receivers and expanded tuples. Names
alone confer no guard or sink authority. Missing policy or compiler capabilities
fail closed. Arbitrary Lua, unselected source bodies and runtime permissions
remain outside this static claim.

Teal permits nil for every type; retain runtime absence checks. Source-supported
generics preserve key, value and element relationships. Mutable containers remain
invariant: a container of unknown values is not a universal replacement for a
concrete container. Exact heterogeneous variadic forwarding and repeated-tuple
relationships remain precision limits; known fixed argument and result positions
must still be retained. Optional libraries and legacy supplements are not a
runtime availability guarantee. Secure declarations grant no protected access.

See [addon usage](../../README.md) and the
[release update guide](../../UPDATING.md).

The repository-owned Go CDN extractor has no external module dependencies. Its
config, archive-index, encoding, root and BLTE readers validate format-defined
integrity and bounded reads; unsupported formats or encryption fail explicitly.
The original C# extraction and the Go cold acquisition matched all 4,029 files
for the recorded migration cohort. This establishes that cohort, not universal
future-format support.

The consumer archive carries its Go source closure and paired compiler runtime.
It requires neither producer packages nor acquisition caches. Source acquisition,
deterministic generation and static checks remain enforced; independent API
crosschecks and historical approval enforcement have been removed.
