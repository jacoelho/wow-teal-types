# WoW Teal Types

World of Warcraft Retail API declarations and secret-value checks for addons
written in [Teal](https://github.com/teal-language/tl). You write `.tl` sources,
the `wow-teal` command type-checks them against the WoW API and compiles them
to plain Lua 5.1 that WoW loads. Nothing from this package runs inside WoW.

## Requirements

- Linux or macOS.
- Go 1.27 or later.
- Teal 0.24.8 exactly: `luarocks install tl 0.24.8`. The command uses `TL` when
  set, otherwise `tl` on `PATH`, and rejects any other compiler version.

## Quick start

Try the bundled example from a checkout of this repository:

```sh
make example
```

This checks [examples/WowTealExample/src](examples/WowTealExample/src) and
writes `examples/WowTealExample/build/WowTealExample.lua`. The
`examples/WowTealExample` folder is then a complete addon: link it into
`Interface/AddOns` (see [Load it in WoW](#load-it-in-wow)) and it shows a
small panel with a health bar.

## Set up your addon

### 1. Add the toolkit

Add this repository to your addon, for example as a Git submodule, which pins
it to one commit:

```sh
git submodule add https://github.com/jacoelho/wow-teal-types vendor/wow-teal-types
```

An extracted release archive at the same path works too.

### 2. Lay out the addon

```text
MyAddon/
  MyAddon.toc
  src/Main.tl
  vendor/wow-teal-types/
```

Sources live in `src`; compiled Lua goes to `build`. Keep the two apart.

### 3. Write Teal

`src/Main.tl`:

```teal
local type W = require("wow")

local panel: W.Frame = CreateFrame("Frame", nil, UIParent)
panel:SetSize(220, 32)
panel:SetPoint("CENTER")
```

`local type` imports the declarations for type checking only; the line
disappears from the generated Lua. There is no runtime `wow` module, so a plain
`require("wow")` fails in game.

### 4. List the generated Lua in your TOC

`MyAddon.toc` lists the compiled files, not the `.tl` sources:

```text
## Interface: <interface number>
## Title: My Addon
build\Main.lua
```

Copy the `## Interface` line from
[the example TOC](examples/WowTealExample/WowTealExample.toc); it matches the
client these declarations describe. You own the TOC and any assets.

### 5. Check and build

From the addon folder:

```sh
vendor/wow-teal-types/bin/wow-teal check src
vendor/wow-teal-types/bin/wow-teal build src --output build
```

`check` runs Teal and the secret-value checks. `build` runs the same checks and
then writes Lua, keeping the directory structure: `src/Main.tl` →
`build/Main.lua`, `src/ui/Panel.tl` → `build/ui/Panel.lua`. A failed build
leaves the previous output in place.

### 6. Load it in WoW

Link the addon folder into the game's AddOns directory, then `/reload` after
each build:

```sh
ln -s "$PWD" "/Applications/World of Warcraft/_retail_/Interface/AddOns/MyAddon"
```

Adjust the path to your installation. The folder name must match the TOC name.
To publish the addon, ship the TOC, `build` and your assets; `src` and `vendor`
are not needed in game.

## Build details

- **Output ownership.** `build` creates the output directory and afterwards
  replaces only output it generated and that is unchanged; it rejects a
  directory it did not create, and edited or foreign files in it. Do not write
  other files there, and do not run two builds into the same directory at once.
- **What is checked.** Select every directory you own. Imported modules are type
  checked, but secret-value checks run only on selected sources.
- **Skipped directories.** Checking a directory skips declaration files, `.git`,
  `.venv`, `__pycache__`, `build`, `dist`, `vendor` and `node_modules`. These
  are also hidden from imports unless selected or passed with `-I`.
- **Symlinks.** Symlinked sources, include roots and links inside them are
  rejected.
- **Configuration.** The command ignores `tlconfig.lua` and `TL_PATH`; pass
  options explicitly. Use `-I <dir>` for other module directories, and pass the
  same options to `check` and `build`.
- **Inputs.** Each run checks and compiles one bounded copy of the inputs, so
  edits made during a run apply to the next one. See
  [architecture](docs/architecture.md) for limits.

## Addon globals

Declare your own globals in a declaration file, such as
`types/my_addon_env.d.tl`:

```teal
global MyAddonCounter: number
```

and add it to the WoW environment:

```sh
vendor/wow-teal-types/bin/wow-teal check src -I types --global-env-def my_addon_env
vendor/wow-teal-types/bin/wow-teal build src --output build -I types --global-env-def my_addon_env
```

## Secret values

WoW returns some values as secrets that addon code may pass on but not inspect.
Results that can be secret are typed `W.MaybeSecret<T>` (or `W.Secret<T>` for
handles and tables).

You can pass them to native functions that accept them, and to your own
parameters declared `MaybeSecret`. To compare, compute with or print one, first
prove the same value is accessible: `canaccessvalue(v)` is true, or
`issecretvalue(v)` is false. Then unwrap it with `as`:

```teal
local health = UnitHealth("player")
if not issecretvalue(health) then
   local public_health = health as number
   print(public_health + 1)
end
```

Rules to keep in mind:

- The check must be in the function that uses the value. A cast, a helper that
  checked it, or an earlier result of the same getter does not count.
- An untyped secret result is `W.SecretData`. It also needs a kind check:
  `type(v) ~= "table"` unwraps it as `W.Atom`; `type(v) == "table"` plus
  `canaccesstable(v)` unwraps it as `W.SecretTable`.
- `any`, `W.Data` and your own generic type variables are plain positions, so a
  secret value cannot pass through them.
- Plain `tl check` does not enforce these rules; use `wow-teal`.

## Limits

- Some declarations cover legacy APIs or optional libraries; check they exist
  before use.
- Nilable results are declared without `| nil`, since Teal admits `nil` for
  every type, so keep runtime nil checks.
- Type checking cannot grant protected-call permission or prove runtime
  behaviour. `--secure` selects secure declarations without granting native
  privileges.

The [declaration contract](docs/api/teal-catalogue.md) describes precision
limits. Maintainers updating to a new WoW release follow
[UPDATING.md](UPDATING.md).

## License

Project-authored code and documentation use the [MIT license](LICENSE).
Upstream source material keeps its original terms.
