#!/usr/bin/env python3
"""Find files containing a legacy search term with mention counts and line numbers."""

from __future__ import annotations

import argparse
import re
import sys
from pathlib import Path

SKIP_DIRS = {
    ".git",
    ".cursor",
    "node_modules",
    "__pycache__",
    ".venv",
    "venv",
}


def parse_args(argv: list[str] | None = None) -> argparse.Namespace:
    epilog = """\
match modes
  exact (default)
      Whole-word match. "workspace" matches standalone uses and tokens like
      --workspace, but not "workspaces" or "MONMS_WORKSPACE".

  --partial
      Substring match anywhere on a line. "workspace" also matches
      "workspaces", "multi-workspace", etc.

output
  One line per file that contains at least one match:

    /abs/path/to/file.go (3 mentions): lines 10, 42, 88

  Files with no matches produce no output. Exit code is 0 when the search
  completes successfully (including zero hits).

skipped paths
  .git, .cursor, node_modules, __pycache__, .venv, venv

examples
  %(prog)s --search "workspace"
      Find whole-word "workspace" under the current directory.

  %(prog)s --search "workspace" --partial
      Find any line containing the substring "workspace".

  %(prog)s --search "MONMS_WORKSPACE" --root .
      Search the repo root for the legacy env var name.

  %(prog)s --search "old workspace" --root ./site
      Exact phrase match (multi-word --search values are supported).
"""

    parser = argparse.ArgumentParser(
        prog="find-legacy.py",
        description="Find legacy terminology left in a codebase after renames or migrations.",
        epilog=epilog,
        formatter_class=argparse.RawDescriptionHelpFormatter,
    )
    parser.add_argument(
        "--search",
        required=True,
        metavar="TERM",
        help='term to find (quote multi-word strings); exact whole-word match by default',
    )
    parser.add_argument(
        "--partial",
        action="store_true",
        help="substring match instead of whole-word match",
    )
    parser.add_argument(
        "--root",
        type=Path,
        default=Path.cwd(),
        metavar="DIR",
        help="directory to search recursively (default: current working directory)",
    )
    return parser.parse_args(argv)


def build_matcher(term: str, partial: bool) -> re.Pattern[str]:
    if partial:
        return re.compile(re.escape(term))
    return re.compile(r"\b" + re.escape(term) + r"\b")


def iter_files(root: Path) -> list[Path]:
    files: list[Path] = []
    root = root.resolve()
    for path in sorted(root.rglob("*")):
        if not path.is_file():
            continue
        if any(part in SKIP_DIRS for part in path.parts):
            continue
        files.append(path)
    return files


def scan_file(path: Path, matcher: re.Pattern[str]) -> list[int]:
    try:
        text = path.read_text(encoding="utf-8", errors="replace")
    except OSError:
        return []

    lines: list[int] = []
    for index, line in enumerate(text.splitlines(), start=1):
        if matcher.search(line):
            lines.append(index)
    return lines


def main(argv: list[str] | None = None) -> int:
    args = parse_args(argv)
    term = args.search
    if not term:
        print("error: --search must be a non-empty string", file=sys.stderr)
        return 2

    root = args.root.resolve()
    if not root.is_dir():
        print(f"error: root is not a directory: {root}", file=sys.stderr)
        return 2

    matcher = build_matcher(term, args.partial)
    results: list[tuple[Path, list[int]]] = []

    for path in iter_files(root):
        line_numbers = scan_file(path, matcher)
        if line_numbers:
            results.append((path.resolve(), line_numbers))

    for path, line_numbers in results:
        count = len(line_numbers)
        line_list = ", ".join(str(n) for n in line_numbers)
        print(f"{path} ({count} mention{'s' if count != 1 else ''}): lines {line_list}")

    return 0


if __name__ == "__main__":
    raise SystemExit(main())
