"""Pytest fixtures for wrapper tests."""

from __future__ import annotations

import shutil
import subprocess
import sys

import pytest


@pytest.fixture(autouse=True)
def _git_identity_for_tests(monkeypatch: pytest.MonkeyPatch, tmp_path_factory: pytest.TempPathFactory) -> None:
    """CI images often have no global git user; commits in tests must still work."""
    git_bin = shutil.which("git")
    if git_bin is None:
        pytest.fail("git executable not found on PATH")

    home = tmp_path_factory.mktemp("gh_pytest_git_home")
    monkeypatch.setenv("HOME", str(home))
    if sys.platform == "win32":
        monkeypatch.setenv("USERPROFILE", str(home))
    subprocess.run(
        [git_bin, "config", "--global", "user.email", "pytest@example.com"],
        check=True,
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
    )
    subprocess.run(
        [git_bin, "config", "--global", "user.name", "pytest"],
        check=True,
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
    )
    yield
