# Declaration contract

`types/wow/wow.d.tl` exports native handles, widget interfaces, enums, structures,
aliases, callbacks and signatures. `wow_env.d.tl` supplies public globals;
`wow_secure_env.d.tl` supplies the separately selected secure environment.
`native_values.d.tl` owns the secret-value markers re-exported by `wow`.

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
An untyped input uses `any`; an untyped output uses `W.Data` (a public value of
undocumented shape, narrowed with `is`), or `W.SecretData` when the source
marks it secret. Missing precision does not justify dropping symbols or
inventing concrete results or generic relationships.
The retained surface is not a claim to infer every client global or mixin.
FrameXML identity helpers keep their argument types through contract corrections:
`Mixin` returns its first argument (`function<T>(T, any...): T`), and
`CopyTable`, `tInvert` and `tFilter` keep or invert their map types.

The five draw-layer and vertex-color methods from `SimpleRegionAPI` belong to
`LayeredRegion`. `Frame` retains common `Region` capabilities; `ModelScene` keeps
its distinct one-argument layer setter. FontString, TextureBase and VectorGraphics
retain layered capabilities. This explicit routing avoids incompatible inherited
contracts without weakening mutable-field invariance.

## Compiler and secret values

Checks and builds run stock Teal `0.24.8` (`luarocks install tl 0.24.8`), taken
from `TL` or else `tl` on `PATH`. The consumer verifies the package bytes first,
then requires `tl --version` to print exactly `0.24.8`; the compiler is pinned
by version, not by hash. `tools/check_secrets.lua` asserts the same version
in-process because it reads compiler internals. Any other version fails closed.

`native_values.d.tl` declares the secret model, re-exported by `wow`:

```teal
record Secret<T> is userdata end            -- static marker, never a runtime carrier
type MaybeSecret<T> = T | Secret<T>
type Data = number | string | boolean | {any: any} | function(any...): any...
type Atom = number | string | boolean | function(any...): any...
type SecretData = MaybeSecret<Atom> | Secret<SecretTable>
type SecretTable = {SecretData: SecretData}
```

`Data` is a public value of undocumented shape. `SecretData` is a possibly
secret value of undocumented shape: a scalar or function, or a table whose keys
and values are again `SecretData`, so nested reads stay opaque at every depth.

Opacity comes only from these declarations. The checker takes them from the
`native_values` module that Teal loaded and compares the declaration object
that each type reference resolves to, so a record of the same name elsewhere is
an ordinary type. Renaming or removing any of the six names fails closed. Teal rejects a union of
two userdata types, so `MaybeSecret<Frame>` cannot be used; secret handles are
emitted as `Secret<Frame>` alone.

### Checker rules

Stock Teal type-checks ordinary flows. The checker adds:

- **Operations.** Truthiness, `==`/`~=`, arithmetic, indexing, `is`, `type()`,
  iteration and casts on a value of `Secret` type need a same-value proof from a
  native guard in the consuming function. `canaccessvalue` (or
  `canaccessallvalues`) proves access, `issecretvalue` returning false proves
  publicness, and `canaccesstable` proves table access. Publicness implies
  access, because `canaccessvalue` is true for every non-secret value; it never
  implies table access, because a table is not a secret value yet may hold
  secret contents. Equality needs access to both operands. `type()` and scalar
  conversion need publicness.
- **Public stores.** Assigning a value typed exactly `number`, `integer`,
  `string`, `boolean` or `nil` to a local also proves that local public, so
  replacing a secret with a default works without a second guard:

  ```teal
  local alpha = tex:GetAlpha()
  if issecretvalue(alpha) then alpha = 1 end
  local opacity = alpha as number
  ```

  Teal narrows `alpha` to `number` inside the branch, but the `if` join
  widens it back to `MaybeSecret<number>`, so the `as` is still needed. Only
  positional pairs count; `any`, unions, fields, globals, values filled from a
  multi-return, and a local named twice in one assignment prove nothing. A
  local written anywhere in the file keeps its proofs only until the next
  call, so unwrap before calling anything. Target expressions such as
  `t[alpha as number]` read the value from before the assignment.
- **Globals carry no proofs.** Teal keeps an assignment's narrowing of a
  global across calls that may rewrite it, so reading a `MaybeSecret` global
  that Teal has narrowed to a plain type is rejected. Copy it to a local and
  guard the local.
- **Guards are always declared.** `issecretvalue` and the other guards are
  typed functions, so write `issecretvalue(x)`, not the Lua feature check
  `issecretvalue and issecretvalue(x)`; the checker drops proofs from the
  branch where the guard is missing.
- **Unwrap.** `x as T` is the only unwrap. `T` must equal the payload of one
  `Secret` arm of the operand's type. Scalars and handles need publicness
  (`if not issecretvalue(x) then local n = x as number end`); `Secret<SecretTable> as SecretTable` (nilable or not) needs
  `canaccesstable`.
  `SecretData` needs a kind proof on the same value as well: `x as W.Atom` needs
  `type(x) ~= "table"` (or `type(x) ==` a non-table kind) after
  publicness, and `x as W.SecretTable` needs `type(x) == "table"` plus
  `canaccesstable(x)`. Without the kind proof a public table with secret contents
  would unwrap as `Atom`. `SecretData` unwraps only as `Atom`, not as `number` or
  `string`; narrow the `Atom` afterwards. A cast with a table or record target
  from a type with no table arm (`Atom`, scalars, functions) is rejected.
- **Wrap.** `e as Secret<T>` is allowed only from exactly `T` or `Secret<T>`,
  so `{number} as W.Secret<W.SecretTable>` and `{number} as W.SecretData` are
  rejected. A value of unknown shape (`any`, `Data`, an authored type variable,
  or that arm of a union such as `W.Data | nil`, whose other arms must still
  wrap) enters only as `W.SecretData` or `W.Secret<W.SecretTable>`; a cast from it to
  a type with a writable opaque position (container element, record value field,
  function or authored method parameter, or such a position inside a function
  result) is rejected, which closes
  `(t as any) as {MaybeSecret<number>}`. Top-level and result positions stay
  allowed (`a as W.MaybeSecret<number>`, `a as function(): W.MaybeSecret<number>`).
- **Iteration and raw access.** `pairs` and `next` over a `SecretTable` yield
  `SecretData` keys and values that need proofs before use. `rawget` is the stdlib
  generic, so a read from a `SecretTable` stays `SecretData`.
- **Lookalikes.** A cast `S as R` to an authored record `R` is compared against
  the generated handles the value could be at runtime when `R` has generated bases
  or is userdata, or when `S` is a generated type, `userdata`, or an authored
  record or interface other than a fresh table (a table constructor, bare or
  through `setmetatable`). The candidates are the generated types the value could
  be: descendants of every generated type either side names and, for an authored
  `S`, only those whose members `S`'s members all admit (an empty interface
  admits every generated type). Each member `R` declares beyond its generated
  bases must admit, under the flow rule, the same-named member of every
  candidate; `record Pane is W.Button` redeclaring `GetText(): string` is
  rejected because `Button:GetText` returns `MaybeSecret<string>`. A member is
  exempt only when the operand is syntactically `Mixin(x, ..., M, ...)` and `M`
  supplies a `function M:Name` / `function M.Name` body: either `M` is `R`
  itself, declared in any selected file, or `M` is a table in the same file whose
  member type flows into `R`'s. Assigned members (`M.Name = f`) and stored
  `Mixin` results are not bodies. Bodies count wherever they appear in the file,
  so a body defined only after the `Mixin` call has run is still exempt; the
  checker does not model statement order.
- **Flow.** At every flow point (call arguments, local declarations, assignments,
  returns, table-constructor fields and function values) a `Secret` position in
  the actual type must not meet a plain position in the expected type. The
  comparison is deep: containers in both directions, function parameters
  contravariantly and results covariantly. A table constructor is compared item
  by item.
- **Sinks.** `any` and `Data` are plain positions, so a Secret cannot reach them.
  At stdlib and generated-declaration sinks, methods and other function-typed
  record members are not inspected (see consequences).
  An authored generic type variable is also a sink, and a value of type-variable
  type cannot reach `any`/`Data` (generic records are instantiated without a
  checked call); `MaybeSecret<T>` is the generic transport. Type variables of
  stdlib and generated declarations transport values: each argument is compared
  against its declared parameter, and every value bound to a type variable must
  be admitted, under the authored rule, by every caller-owned position that
  variable flows out to (`table.insert(list, v)`, `rawset`, callback
  parameters). The call result is re-typed from the bound arguments, so its
  destination is checked where the call expression flows; the method exemption
  applies only when that destination is a trusted `any`/`Data` parameter. An
  opaque argument whose parameter shape the checker cannot match to the type
  variables (for example a record passed as a `{K:V}` map) is rejected rather
  than trusted. `error`, `coroutine.resume` and `coroutine.yield` hand their
  `any` inputs back to authored code as `any`, so they get the authored rule.
- **Builtins.** `select`, `pcall`, `xpcall`, `table.pack`, `table.unpack`,
  `ipairs`, `pairs` and the other stdlib transports are recognized by their
  stdlib declaration location. A shadowing local has a different location and
  is checked as authored code.

Proofs are invalidated by reassignment, by writes in loops and by calls to
closures that write the guarded binding. A cast, a helper's permission check or
a previous getter result does not establish access to the current value.

### Native call policy

The finalized declaration projection owns `types/wow/call_policy.lua`; the
checker does not infer permissions from signature text. Schema 1:

```lua
return {
   schema = 1,
   compiler = "0.24.8",
   declarations = {
      ["types/wow/wow_env.d.tl"] = {
         ["12:23"] = {
            symbol = "C_Example.Accept",
            secret_parameters = { 1 },
            secret_vararg_from = 3, -- optional
         },
      },
   },
}
```

Paths are limited to `wow.d.tl`, `wow_env.d.tl` and `wow_secure_env.d.tl` under
`types/wow/`. Each key is the one-based line and column of the originating
`function` keyword, one entry per overload; aliases keep that origin.
Positions count the full argument tuple, including an explicit or `__call`
receiver at position 1. A permitted position accepts a Secret where the declared
type is plain; a vararg permission covers every position from its start. Every
native-backed overload, including reviewed literal-factory overloads, must
render exactly one parameter per native slot (an ABI tuple spans its width, the
repeating tail is one vararg); generation rejects any other arity rather than
shifting or dropping a permission. Contents
secrecy is carried by `SecretTable`, so any `secret_contents_parameters` or
`secret_contents_vararg_from` fails closed.

Only the canonical guards carry `guard`: `canaccessvalue` (`"value"`),
`canaccessallvalues` (`"values"`), `canaccesstable` (`"table"`) and
`issecretvalue` (`"secret"`). A shadowing function of the same name gains no
authority. Policy provenance is the source coordinate plus the package hash.

The checker fails closed on a wrong compiler version, a missing or malformed
policy, changed marker declarations, a missing stdlib identity, a declaration
module loaded from outside the package, or an expression whose type the compiler
did not record. Arbitrary Lua, unselected source bodies and runtime permissions
remain outside this static claim.

### Consequences and lost guarantees

- `any` is a sink: an authored pass-through such as `function(v: any): any` no
  longer accepts a Secret. Declare the position `MaybeSecret<T>` or unwrap first.
- A handle is not itself secret. At a stdlib or generated-declaration sink
  (`print`, `tostring`, `Mixin`'s mixin arguments, untyped native inputs) the sink check ignores
  function-typed record members, so a frame whose methods return `MaybeSecret`
  is accepted; value fields, elements and tuple members still count. Authored
  `any`/`Data` positions and casts keep the full rule, because `v as L` from
  `any` is unchecked and would reattach a lookalike's plain methods; storing a
  frame in an authored `{any}` is therefore rejected.
- An upcast that hides an opaque value field is rejected, since a later
  downcast could re-type the field as plain. A record with a `MaybeSecret`
  field therefore cannot be passed, stored or returned as a supertype that
  lacks the field (an interface, or `W.Frame` for `record Pane is W.Frame`). Hiding a method is
  allowed so handle hierarchies keep working; the lookalike rule catches the
  later downcast.
- Casts between containers compare elements invariantly when either side is
  opaque, so `{number} as {MaybeSecret<number>}` is rejected.
- Lookalike limit: casts from `any` or `Data` to an authored record without
  generated bases are not compared, so `a as Named` with a plain `GetName` is
  accepted. Authored `any` already rejects handles; a handle reaches such a cast
  only through a native delivery (below).
- Handler payload limit: `SetScript`, `RegisterCallback` and similar natives take
  the handler as `any`, so a handler's declared parameters
  (`function(self, event, x: number)`, `function(_, f: Pane)`) receive event
  payloads and re-emitted handles unchecked. Closing this needs generated event
  payload types and typed per-script overloads; it is a follow-up.
- Laundering limit: `(plain as any) as W.Secret<W.SecretTable>` is accepted
  (the conservative entry), after which a write through the plain alias is not
  tracked.
- Secrecy is a property of a type, not of a value. Curve evaluation returns
  `colorRGBA`, so every `ColorMixin` field is `MaybeSecret<number>`, including
  static constants such as `RAID_CLASS_COLORS[class].r` and
  `GRAY_FONT_COLOR.r`. Arithmetic on them needs the access checks and an `as`
  unwrap. A curve point that is a number or a colour is
  `MaybeSecret<number> | colorRGBA`: only the primitive member is wrapped, and
  the colour carries its secrecy in its fields.
  Rejected alternative: a separate plain colour type for FrameXML constants.
  Derived types inherit their base's secrecy, so it would have to duplicate
  `ColorMixin` instead of extending it, and it could not be passed where
  `colorRGB` is expected.
- The scalar guard is `not issecretvalue(x)` alone. Rejected alternative:
  requiring `canaccessvalue(x)` as well. Patch 12.0.0 documents `canaccessvalue`
  as true for any value that is not secret, so the second check proves nothing,
  and shipping addons (EllesmereUI: about 1,400 `issecretvalue` guards, 3
  `canaccessvalue`) guard with `issecretvalue` alone from tainted code.
- Conditionally secret results stay `MaybeSecret` regardless of the caller's
  context. Unit getters are secret for some unit tokens only (`UnitClass` and
  `UnitHealth` are often public for `"player"`), and region geometry and
  visibility (`GetWidth`, `GetLeft`, `IsShown`, `GetAlpha`) are secret only when
  the object has secret anchoring or a secret aspect, including on addon-owned
  frames anchored to a secret one. Guard each value where it is used.
  Rejected alternative: a frame-level proof such as `not frame:HasSecretValues()`
  or `not frame:IsAnchoringSecret()`. Neither is documented to cover aspects or
  anchor secrecy inherited from another region, most such queries return
  `MaybeSecret<boolean>` themselves, and any setter or anchor change on the
  frame or an anchor ancestor would invalidate the proof without the checker
  seeing it.
- A `MaybeSecret<number>` cannot be stored into a `SecretTable`; Teal rejects it
  against `SecretData`. Build typed tables instead.
- Breaking: a downcast from an empty authored interface (`b as Pane` with
  `record Pane is Base`, `interface Base end`) is compared against every
  generated type, because a handle could have entered `Base`; it is rejected when
  `Pane` redeclares an opaque native member such as `GetWidth(): number`.
- Breaking: `CreateFrame("Button", nil, p, "MyTemplate") as MyButton`, where the
  XML template applies the mixin, is rejected when `MyButton` redeclares an
  opaque native method. Call `Mixin(frame, MyButtonMixin)` in Teal; re-copying is
  idempotent.
- Rejected alternatives: exempting lookalikes by record body alone (it legalises
  `GetMouseFoci()[1] as MyButton` and any lookalike with a trivial body); a
  read-only `SecretTable` rule (any same-shaped map bypasses an alias-name rule,
  and tl 0.24.8 raw types carry no alias identity); per-native re-emitter or
  allowlist policies (they break real senders such as `SetAttribute` and miss
  unlisted natives).
- Treating every `any` as opaque was rejected: most `any` inputs are untyped
  rather than secret, so it would be untrue to the sources.
- Outside secrecy, soundness is stock Teal's: arrays and maps are covariant (an
  alias can insert a wider element), function arguments are bivariant without
  arity or variadic storage checks, poly method receivers are not bound, and an
  undocumented `Data` value needs `is` narrowing before use.
- Stock Teal 0.24.8 crashes (`find_nominal_type_decl` assertion) on a record field
  whose type is a local alias of an instantiated generic, such as
  `local type A = W.MaybeSecret<number>` used as `payload: A`. The crash is a
  non-zero exit, so checks fail closed; spell the instantiated type directly.

## Absence and call-site narrowing

Generated declarations omit `nil` union members: a result Blizzard marks
nilable is `T`, not `T | nil`. Teal admits nil for every type, so the member
adds no checking, while tl 0.24.8 cannot narrow `T | nil` through `if x`,
`if not x then return end` or `x ~= nil`; only `x is T` narrows. Keep runtime
absence checks. Optional parameters keep their `? T` marker, and a lone `nil`
type is retained. Rejected alternative: keeping `| nil` and requiring `is`
narrowing. It blocked idiomatic guards at every nilable result without
detecting any additional nil use.

`CreateFrame` and `AnimationGroup:CreateAnimation` select their result from
literal arguments. A listed frame type returns that widget type, and
`"BackdropTemplate"` with Frame, Button, CheckButton, EditBox, ScrollFrame,
Slider or StatusBar returns an interface with both bases (for example
`Frame_BackdropTemplate`). A listed animation type such as `"Alpha"` or
`"Translation"` returns that animation type. Literals are case-sensitive while
the client is not, so `"BUTTON"`, template lists, other templates and
non-literal strings return the generic `Frame` or `Animation`; assert the
intended type with `as`. Each
[literal factory](../../internal/wowapi/inputs/overrides.json) admits only
types whose base chain reaches its base type, and keeps the native contract of
the factory for every narrowed overload. `CreateFrame` declares a reviewed
generic signature, because its source annotation selects the result through a
string-dependent generic that Teal cannot express. A native-backed factory
cannot declare one: its fallback is the native signature, and its overload
pattern must match the native parameters position by position, with literal
placeholders only where the native takes a string. Native secret permissions
are granted by position, so this keeps each permission on its own parameter.

`SetFormattedText` accepts `string | number` format arguments after the format
string; native metadata omits them. They share the method's secret permission:
FontString permits secret format arguments (FrameXML passes `UnitName` results,
for example in `ChatFrameEditBox.lua`), Button does not. At input positions a
secret-permitting union is projected per member
(`W.MaybeSecret<string> | W.MaybeSecret<number>`), because `Secret<T>` is
invariant and `Secret<string | number>` would reject a `Secret<string>`.

GlobalStrings such as `ACCEPT` or `RAID_TARGET_1` are loaded by the client from
outside `Interface/AddOns`, the only extracted subtree, so they are not declared;
declare the names an addon uses in its own environment fragment. `_G` is stock
Teal's record of declared globals: `_G.ChatFrame1` checks, while a computed key
such as `_G["ChatFrame" .. i]` needs `getglobal`/`setglobal` or an `as` assertion.

Exact heterogeneous variadic forwarding and repeated-tuple relationships remain
precision limits. Optional libraries and legacy supplements are not a runtime
availability guarantee. Secure declarations grant no protected access.

See [addon usage](../../README.md) and the
[release update guide](../../UPDATING.md).

The repository-owned Go CDN extractor has no external module dependencies. Its
config, archive-index, encoding, root and BLTE readers validate format-defined
integrity and bounded reads; unsupported formats or encryption fail explicitly.
The original C# extraction and the Go cold acquisition matched all 4,029 files
for the recorded migration cohort. This establishes that cohort, not universal
future-format support.

The consumer archive carries its Go source closure, declarations and checker.
It requires an installed Teal 0.24.8 but neither producer packages nor
acquisition caches. Source acquisition,
deterministic generation and static checks remain enforced; independent API
crosschecks and historical approval enforcement have been removed.
