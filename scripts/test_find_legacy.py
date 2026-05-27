#!/usr/bin/env python3
"""Tests for find-legacy.py."""

from __future__ import annotations

import subprocess
import sys
import tempfile
import unittest
from pathlib import Path

SCRIPT = Path(__file__).resolve().parent / "find-legacy.py"


def run_find_legacy(*args: str, root: Path) -> subprocess.CompletedProcess[str]:
    return subprocess.run(
        [sys.executable, str(SCRIPT), *args, "--root", str(root)],
        capture_output=True,
        text=True,
        check=False,
    )


class FindLegacyTests(unittest.TestCase):
    def setUp(self) -> None:
        self.tempdir = tempfile.TemporaryDirectory()
        self.root = Path(self.tempdir.name)

    def tearDown(self) -> None:
        self.tempdir.cleanup()

    def write(self, relative: str, content: str) -> Path:
        path = self.root / relative
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(content, encoding="utf-8")
        return path.resolve()

    def test_exact_whole_word_only(self) -> None:
        path = self.write("exact.txt", "workspace\nworkspaces\nMONMS_WORKSPACE\n")
        result = run_find_legacy("--search", "workspace", root=self.root)
        self.assertEqual(result.returncode, 0)
        self.assertIn(f"{path} (1 mention): lines 1", result.stdout)
        self.assertNotIn("lines 2", result.stdout)
        self.assertNotIn("lines 3", result.stdout)

    def test_partial_matches_substrings(self) -> None:
        path = self.write("partial.txt", "workspace\nworkspaces\nmy-workspace-dir\n")
        result = run_find_legacy("--search", "workspace", "--partial", root=self.root)
        self.assertEqual(result.returncode, 0)
        self.assertIn(f"{path} (3 mentions): lines 1, 2, 3", result.stdout)

    def test_no_matches_returns_empty(self) -> None:
        self.write("clean.txt", "site directory only\n")
        result = run_find_legacy("--search", "wprkspace", root=self.root)
        self.assertEqual(result.returncode, 0)
        self.assertEqual(result.stdout.strip(), "")

    def test_multiple_lines_reported(self) -> None:
        path = self.write("multi.txt", "line1 workspace\nline2 ok\nline3 workspace\n")
        result = run_find_legacy("--search", "workspace", root=self.root)
        self.assertEqual(result.returncode, 0)
        self.assertIn(f"{path} (2 mentions): lines 1, 3", result.stdout)

    def test_requires_search(self) -> None:
        result = subprocess.run(
            [sys.executable, str(SCRIPT), "--root", str(self.root)],
            capture_output=True,
            text=True,
            check=False,
        )
        self.assertEqual(result.returncode, 2)
        self.assertIn("--search", result.stderr)

    def test_quoted_search_with_spaces(self) -> None:
        path = self.write("phrase.txt", "hello old workspace path\n")
        result = run_find_legacy("--search", "old workspace", root=self.root)
        self.assertEqual(result.returncode, 0)
        self.assertIn(f"{path} (1 mention): lines 1", result.stdout)


if __name__ == "__main__":
    raise SystemExit(unittest.main())
