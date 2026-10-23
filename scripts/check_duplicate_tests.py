#!/usr/bin/env python3
"""Fail when two files in one package declare the same top-level function.

Two pull requests can each be green and still land an uncompilable package:
each suite was written against a main branch that did not yet contain its
sibling. That happened twice on 2026-10-23 — `TestCheckPassword` was declared
three times in `internal/auth`, and `TestEngine_SettleCurrent` twice in
`internal/diagnostic` — and in both cases only the combination was broken, so
no individual pull request check could see it. `go test ./...` catches it, but
only after the merge. This check runs before it, so the failure names the
duplicate instead of printing a compiler error against the package.

It scans `_test.go` files only. Duplicate test helpers matter as much as
duplicate test entry points, because the compiler rejects both, so every
top-level function counts. Methods are ignored: they belong to a type and two
types may share a method name. Files in a directory are grouped by their
package clause, because a package's external test package (`foo_test`) may
legitimately declare the same helpers without colliding.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parent.parent
SKIP_DIRS = {"node_modules", ".git", ".next", "out", "vendor", ".understand"}

PACKAGE = re.compile(r"^\s*package\s+(\w+)", re.MULTILINE)
FUNC = re.compile(r"^func\s+([A-Za-z_]\w*)\s*\(", re.MULTILINE)


def test_files() -> list[Path]:
    found: list[Path] = []
    for path in ROOT.rglob("*_test.go"):
        if any(part in SKIP_DIRS for part in path.parts):
            continue
        found.append(path)
    return sorted(found)


def main() -> int:
    # (directory, package) -> function name -> [locations]
    seen: dict[tuple[Path, str], dict[str, list[str]]] = {}
    files = test_files()
    for path in files:
        try:
            text = path.read_text(encoding="utf-8", errors="replace")
        except OSError as exc:
            print(f"cannot read {path}: {exc}")
            return 1
        pkg_match = PACKAGE.search(text)
        pkg = pkg_match.group(1) if pkg_match else ""
        key = (path.parent, pkg)
        table = seen.setdefault(key, {})
        for match in FUNC.finditer(text):
            name = match.group(1)
            line = text.count("\n", 0, match.start()) + 1
            rel = path.relative_to(ROOT)
            table.setdefault(name, []).append(f"{rel}:{line}")

    problems = 0
    for (directory, pkg), table in sorted(seen.items(), key=lambda item: str(item[0])):
        for name, where in sorted(table.items()):
            if len(where) < 2:
                continue
            problems += 1
            rel_dir = directory.relative_to(ROOT)
            print(f"{rel_dir} [{pkg}]: {name} declared {len(where)} times")
            for location in where:
                print(f"    {location}")

    if problems:
        print(
            f"FAIL: {problems} duplicate declaration(s) in {len(files)} test files. "
            "Rename one body to say what it asserts, or delete it if it is an exact duplicate."
        )
        return 1
    print(f"OK: no duplicate function declarations in {len(files)} test files.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
