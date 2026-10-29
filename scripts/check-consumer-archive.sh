#!/usr/bin/env sh
set -eu
project_root=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
"$project_root/bin/wow-teal" package >/dev/null
package_version=$("$project_root/bin/wow-teal" version)
wow_version=$(cat "$project_root/.wow-version")
archive="$project_root/dist/wow-teal-types-$package_version-retail-$wow_version.zip"
workspace=$(mktemp -d)
trap 'rm -rf "$workspace"' EXIT HUP INT TERM
unzip -q "$archive" -d "$workspace"
cd "$workspace/wow-teal-types"
# The compiler is external (stock Teal pinned by exact version), so tools/teal
# must not reappear alongside the producer-only trees.
for omitted in scripts internal/release internal/wowapi tools/cdn tools/wow-types tools/teal .github; do
    if test -e "$omitted"; then
        echo "consumer archive contains excluded files: $omitted" >&2
        exit 1
    fi
done
make test check example package
bin/wow-teal version
bin/wow-teal rockspecs --check
bin/wow-teal rockspecs --repository example/wow-teal-types
bin/wow-teal rockspecs --repository example/wow-teal-types --check
