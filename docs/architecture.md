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
| Interpret evidence and project declarations | `internal/wowapi` | Preserve raw evidence, finalize a typed declaration graph, render once, and audit the same emitted bytes. |
| Check and build an addon | `internal/consumer` | Capture bounded inputs, verify copied package bytes, run all compiler phases against that capture, then publish checked Lua. |
| Extract the selected client | `tools/cdn` | Independent Go module; verified acquisition feeds the release boundary. |
| Enforce secret operations | `tools/check_secrets.lua` | Teal 0.24.8 typed checks and canonical native declaration identity. |
| Reject native runtime imports | `tools/check_runtime_imports.lua` | Inspect generated Lua before publication. |

CLI packages translate arguments and pass explicit dependencies. Release
metadata preparation returns artifact bytes; it cannot publish them. Candidate
writing and read-only comparison are separate operations. Recipe rendering is
consumer-owned, while its maintained template belongs to the release cohort.
Standalone metadata, catalogue, coverage and template-refresh writers are removed.

## Captured compiler inputs

A consumer invocation owns a private directory until completion. It captures the
package manifest and verifies each copied package file against that manifest.
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
and both errors. Probe publication follows the same local recovery rule. These
small capability-specific transactions do not need a shared filesystem framework.
Callers serialize publication to the same destination.

Release publication owns a larger journaled cohort transaction and its recovery
command. Its journal and fingerprint gates remain distinct from consumer/probe
directory replacement. None of these operations claims power-loss atomicity.

## Declaration generation

Raw source records retain metadata that Teal cannot express. Interpretation builds
a mutable catalogue, then finalization resolves graph closure into concrete
fields, methods and signatures. Rendering consumes that finalized representation
without changing it. Cross-checks inspect the exact rendered files and finalized
signatures; they do not reconstruct or render a competing graph. Full generation
is repeated to verify determinism across acquisition, interpretation and rendering.

A generic mutable map renderer was rejected because it allowed rendering to change
the graph being audited. A broad filesystem abstraction was rejected because local
rename injection is sufficient to test the actual recovery boundary.
