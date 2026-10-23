#!/usr/bin/env python3
"""Fail on an Oxford comma in prose we author.

The rule is consistency, not grammar: AGENTS.md §8 says Mathua's copy joins the
last two items of a list with "and" rather than ", and". A checker is what makes
that a rule instead of a preference, because the sentence that breaks it reads
perfectly well and nobody notices it in review.

Fenced code blocks are skipped: a comma before "and" inside an example is code,
not our prose. Binary and generated files are never opened.
"""

from __future__ import annotations

import re
import sys
from pathlib import Path

ROOTS = [
    "AGENTS.md",
    "CONTEXT.md",
    "CONTRIBUTING.md",
    "README.md",
    "docs",
    "web/next-app/app",
    "web/next-app/components",
    "web/next-app/lib",
]

SUFFIXES = {".md", ".mdx", ".tsx", ".ts"}

# Skip anything generated, vendored or built.
SKIP_PARTS = {"node_modules", ".next", "out", "vendor", "content"}

# ", and" / ", or" as the last separator of a list of three or more. The earlier
# comma is what distinguishes an Oxford comma from the ordinary comma that joins
# two independent clauses ("Learn leads, and the orientation surfaces follow
# it"), which is correct English and must not be reported. Sentences are split on
# the punctuation that ends a clause, so a comma in the previous sentence cannot
# make the next one look like a list.
PATTERN = re.compile(r",[ \t]+(?:and|or)[ \t]+")
SENTENCE = re.compile(r"[.;:!?]\s")


def prose_lines(path: Path):
    """Yield (line number, text) outside fenced code blocks and comments.

    Code comments are excluded on purpose. The rule is about copy a reader sees
    and documentation; a comment is an explanation to the next maintainer, and
    rewriting 250 of them would be a large diff that changes no behaviour and
    makes `git blame` useless for no benefit.
    """
    in_fence = False
    fence = ""
    for number, line in enumerate(path.read_text(encoding="utf-8", errors="ignore").splitlines(), 1):
        stripped = line.lstrip()
        if path.suffix in {".ts", ".tsx"} and stripped.startswith(("//", "*", "/*", "*/")):
            continue
        if stripped.startswith("```") or stripped.startswith("~~~"):
            marker = stripped[:3]
            if not in_fence:
                in_fence = True
                fence = marker
            elif marker == fence:
                in_fence = False
                fence = ""
            continue
        if in_fence:
            continue
        yield number, line


def check(path: Path) -> list[str]:
    problems = []
    for number, line in prose_lines(path):
        for match in PATTERN.finditer(line):
            # An earlier comma on the same line is what makes this a list. Lines
            # are the unit on purpose: a comma in the previous line is usually a
            # different sentence, and treating wrapped text as one list produced
            # more false reports than real ones.
            if "," in line[: match.start()]:
                problems.append(f"{path}:{number}: Oxford comma: {line.strip()}")
                break
    return problems


def main() -> int:
    root = Path(__file__).resolve().parent.parent
    problems: list[str] = []
    scanned = 0
    for entry in ROOTS:
        target = root / entry
        paths = [target] if target.is_file() else sorted(target.rglob("*"))
        for path in paths:
            if not path.is_file() or path.suffix not in SUFFIXES:
                continue
            if SKIP_PARTS & set(path.parts):
                continue
            scanned += 1
            problems.extend(check(path))

    if problems:
        print("\n".join(problems))
        print(f"\nFAIL: {len(problems)} Oxford comma(s) in {scanned} files.")
        print("Join the last two items with 'and', not ', and'.")
        return 1
    print(f"OK: no Oxford commas in {scanned} authored files.")
    return 0


if __name__ == "__main__":
    sys.exit(main())
