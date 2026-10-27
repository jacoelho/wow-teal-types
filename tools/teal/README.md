# Paired Teal compiler

`./tools/teal/tl` runs Teal `0.24.8-wow.1`. Its public `unknown` type carries an
arbitrary value without permitting inspection or conversion to a concrete type
or legacy `any`. This compiler and `tools/check_secrets.lua` form one checking
boundary: ordinary type checking cannot establish WoW secret-value accessibility.

The offline runtime consists of `tl`, `cli.lua`, `tl.lua`, `argparse.lua`,
`contract.lua`, `check_capabilities.lua`, and `LICENSE`. It requires LuaJIT (or a
compatible executable selected with `LUA`). Generated addon code targets Lua 5.1.
The consumer package binds the compiler, checker, and native call policy by hash.

## Source and reproduction

`upstream/tl.lua` is the unmodified generated compiler at official Teal tag
[v0.24.8](https://github.com/teal-language/tl/tree/v0.24.8), commit
`2b6b3c2e53249f3b910e07fc9e7cc03f6b14f3ea`. `cli.lua` is that tag's unmodified `tl`
entry point. `argparse.lua` is unmodified
[argparse 0.7.1](https://github.com/luarocks/argparse/blob/0.7.1/src/argparse.lua).
`provenance.json` records their SHA-256 digests. Teal's MIT license is retained in
`LICENSE`; argparse's MIT notice remains at the beginning of its source.

`unknown.patch` is the sole extension source. The vendored generated Lua source
is patched directly, so bootstrapping does not require an installed Teal compiler
or network. Existing implementation `any` types belong to this pinned upstream
dependency, not the application's declaration contracts.

```sh
python3 tools/teal/rebuild.py          # verify upstream digests and runtime bytes
python3 tools/teal/rebuild.py --write  # regenerate runtime after editing the patch
luajit tools/teal/tests/compiler.lua
go test ./internal/consumer -run '^TestUnknown' -count=1
tools/teal/tl run tools/teal/check_capabilities.lua
```

The capability check prints `unknown=1;type_report=1;lua_target=5.1` only after
validating the paired version, capability table, public unknown type code, and
report API. The checker additionally requires `report.wow_schema == 1` and every
used expression and expanded tuple member's type identity. Unsupported report
shapes stop checking; they do not imply a public value.

## Type contract

Public unknown uses its own `unknown_value` internal identity and
`PUBLIC_UNKNOWN` report code (`0x04000000`), separate from Teal's lax inference
`UNKNOWN`. Aliases, union members, generic substitution, and tuple positions
retain this distinction. Assignment to unknown is allowed; arbitrary calls,
indexing, arithmetic, truthiness, iteration, casts, and legacy-any erasure are
rejected. Tag-only `type` and equality observation are permitted only with the
paired checker proofs described below. Runtime `is` checks may narrow unknown to `number`, `string`, `boolean`,
or `nil`. Integer and structural narrowing are deliberately unsupported: Lua 5.1
cannot establish those contracts with the existing runtime checks.

Function parameters are contravariant and results covariant. Mutable arrays,
maps, tuples, and writable record fields are invariant to prevent an alias from
inserting an invalid value. A known function may transport unknown parameters and
results; an unknown value itself is not callable. Fresh result-only generic
containers remain supported. An unconstrained generic body is checked with an
unknown upper bound while retaining its type-parameter identity; constraints
determine which operations the body can perform. Type reports retain both opaque
bounds and declared constraints for the secret checker. Teal's existing interface-only generic constraint
syntax is unchanged; `T is unknown` is not new syntax in this extension.

`select`, `table.pack`, `table.unpack`, and protected calls retain typed transport
where their contracts permit it. This is not permission to pass unknown to an
arbitrary builtin. The compiler reports canonical builtin identity across
aliases and overloads so the paired checker can distinguish transport from
inspection. Pack types and exact heterogeneous variadic forwarding are outside
this extension; signatures must retain known fixed positions and use opaque
positions only where the source contract lacks that precision.

## Native call policy

The finalized declaration projection owns `types/wow/call_policy.lua`. The checker
consumes it; it does not infer permissions from rendered signature text. Schema 1
is a returned Lua table:

```lua
return {
   schema = 1,
   compiler = "0.24.8-wow.1",
   declarations = {
      ["types/wow/wow_env.d.tl"] = {
         ["12:23"] = {
            symbol = "C_Example.Accept",
            secret_parameters = { 1 },
            secret_vararg_from = 3, -- optional
            secret_contents_parameters = { 1 },
            secret_contents_vararg_from = 3, -- optional
         },
      },
   },
}
```

Paths are package-relative and limited to `wow.d.tl`, `wow_env.d.tl`, and
`wow_secure_env.d.tl` under `types/wow/`. Each key is the one-based line and column
of the originating **function keyword**, with one entry per overload. For example
`global check: function(unknown):boolean` has origin `1:15`. A record field
` f: function(unknown):unknown` begins its function at column 5. Aliases retain
that origin. Argument permissions use one-based positions in the full ABI tuple,
including an explicit method receiver at position 1. A vararg permission applies
to every position from its starting index; a single listed position does not
authorize an unbounded tail.

`secret_parameters` permits an opaque root value. `secret_contents_parameters`
separately permits potentially secret nested members. The two optional vararg
fields have the corresponding meanings. Empty or missing permissions deny the
operation. An unchanged unknown value may hide a table with secret contents:
root accessibility and publicness guards alone cannot establish member safety.
Primitive narrowing after the same-value guards removes that ambiguity.

Only the canonical native guard entries may carry `guard`: `canaccessvalue`
uses `"value"`, `canaccessallvalues` uses `"values"`, `canaccesstable` uses `"table"`,
and `issecretvalue` uses `"secret"`. These functions are permitted to inspect the
opaque boundary for their documented purpose. A shadowing function with the same
name does not acquire guard authority. Policy provenance is its canonical source
coordinate plus the package hash, not its display name.

A local implementation may receive opaque values because its body is checked.
External callbacks and native sinks require their policy or a same-value
accessibility/publicness proof for a narrowed primitive. The checker follows
aliases, invalidates proofs after relevant writes and closure calls, and checks
all expanded argument positions. It fails closed on missing/incompatible policy,
compiler capabilities, or required report identities.

## Runtime observation

`type(value)` returns only a string tag. It does not make an unknown value a
record, table, or callable. The checker requires same-value accessibility and
nonsecret proofs before observing a potentially secret value. `==` and `~=` require accessibility for opaque operands and additionally require
publicness for unknown operands, including unconstrained generics. A known
scalar marker can be compared by an accessible caller without asserting that it
is public. Equality on unknown adds no narrowing facts. Its normal
Lua metamethod and error behavior remains intact, so callers can protect an
observation with `pcall`.

`__wow_observe_field(receiver, key)` returns unknown without asserting a member
contract. Its receiver must already have a static table, record, interface, or
userdata type, and its key must be a string. Unknown and unconstrained generic
receivers are rejected. The intrinsic permits neither writes nor callable or
structural narrowing. The compiler emits `(receiver)[key]`, evaluating each
operand once and preserving ordinary `__index` behavior and errors.

The intrinsic requires a direct call: storing it in a variable or passing it as a
callback is rejected because no runtime intrinsic exists. A shadowing local
function follows ordinary function rules and receives no intrinsic privilege.
The checker permits field observation only after any receiver/key secrecy
obligations are established; hidden contents remain unknown in the result.

Callable record invocations carry the selected originating function coordinate
and an implicit receiver flag. The checker includes that receiver at tuple
position 1 before applying the native policy, just as it does for colon calls.

Extracting an overloaded method binds every overload's receiver before treating
it as a plain function. Calls through the stored value still validate the explicit
receiver and each overload's arguments/results; they do not reread the member.

A scalar union's alternatives describe the root value, not nested contents.
Container fields and tuple entries retain independent contents obligations. Fresh
table literals can establish public contents from their actual keys and values,
including nested literals and expanded result tuples, even when a contextual
parameter type permits opaque members. This proof does not apply to mutable
aliases or opaque values hidden inside a literal.

Explicit inheritance admits a mutable view only when redeclared fields preserve
their storage contract. This includes writable method slots: a substitutable
function override is insufficient when a wider alias can replace the function.
Function storage equality includes optional and variadic positions, and callback
subtyping checks the explicit receiver when a method becomes a plain function.

Unknown-erasure checks traverse the semantic type graph, including generic
wrappers, type arguments, constraints, inherited interfaces, and inferred table
members. Generic wrappers cannot hide opaque results from casts or legacy-any
conversions. These checks retain distinct generic identities and lawful fresh
instantiations.
