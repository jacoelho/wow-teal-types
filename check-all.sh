#!/usr/bin/env sh
set -eu
project_root=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
"$project_root/bin/wow-teal" verify
"$project_root/bin/wow-teal" check-types
"$project_root/bin/wow-teal" check "$project_root/examples/WowTealExample/src"
