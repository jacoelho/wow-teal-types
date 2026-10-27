#!/usr/bin/env python3
"""Reproduce the offline compiler from the pinned source and reviewed patch."""
import argparse
import hashlib
import json
from pathlib import Path
import subprocess
import tempfile


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--write", action="store_true", help="replace the generated runtime")
    args = parser.parse_args()
    root = Path(__file__).resolve().parent
    provenance = json.loads((root / "provenance.json").read_text())
    for name, expected in provenance["unmodified_files"].items():
        actual = hashlib.sha256((root / name).read_bytes()).hexdigest()
        if actual != expected:
            raise SystemExit(f"pinned upstream checksum mismatch: {name}")
    with tempfile.TemporaryDirectory(prefix="wow-teal-build-") as temporary:
        target = Path(temporary) / "tl.lua"
        target.write_bytes((root / "upstream/tl.lua").read_bytes())
        subprocess.run(["patch", "--batch", "--fuzz=0", "-p0", "-i", str(root / "unknown.patch")],
                       cwd=temporary, check=True, stdout=subprocess.DEVNULL)
        generated = target.read_bytes()
    if args.write:
        (root / "tl.lua").write_bytes(generated)
    elif generated != (root / "tl.lua").read_bytes():
        raise SystemExit("tl.lua differs from pinned upstream + unknown.patch; run rebuild.py --write")
    print("bundled compiler matches pinned upstream + unknown.patch")


if __name__ == "__main__":
    main()
