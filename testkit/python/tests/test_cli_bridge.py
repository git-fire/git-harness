from __future__ import annotations

import json
import os
import os.path
import shutil
import subprocess
from pathlib import Path

from git_harness import GitHarnessClient, ScanOptions


def _run_git(repo: Path, *args: str) -> None:
    subprocess.run(
        ["git", *args],
        cwd=repo,
        check=True,
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
    )


def test_analyze_repository_finds_git_dir(tmp_path: Path) -> None:
    repo = tmp_path / "r"
    repo.mkdir()
    _run_git(repo, "init")
    (repo / "a.txt").write_text("x\n")
    _run_git(repo, "add", "a.txt")
    _run_git(repo, "commit", "-m", "init")

    client = GitHarnessClient()
    meta = client.analyze_repository(repo)

    assert os.path.realpath(meta["path"]) == os.path.realpath(str(repo))
    assert meta["name"] == "r"
    assert meta["isDirty"] is False


def test_is_dirty_detects_untracked(tmp_path: Path) -> None:
    repo = tmp_path / "d"
    repo.mkdir()
    _run_git(repo, "init")
    (repo / "b.txt").write_text("y\n")

    client = GitHarnessClient()
    assert client.is_dirty(str(repo)) is True


def test_safety_sanitize_text_masks_token() -> None:
    client = GitHarnessClient()
    # SanitizeText matches GitHub PATs with ghp_ + at least 36 alphanumerics.
    token = "ghp_" + ("a" * 36)
    out = client.safety_sanitize_text(f"pat {token}")
    assert "ghp_" not in out
    assert "[REDACTED]" in out


def test_safety_sanitize_text_empty_string_round_trip() -> None:
    client = GitHarnessClient()
    assert client.safety_sanitize_text("") == ""


def test_safety_format_warning_empty_list() -> None:
    client = GitHarnessClient()
    assert client.safety_format_warning([]) == ""


def test_scan_repositories_finds_nested_repo(tmp_path: Path) -> None:
    outer = tmp_path / "outer"
    inner = outer / "nested" / "proj"
    inner.mkdir(parents=True)
    _run_git(inner, "init")
    (inner / "f").write_text("1\n")
    _run_git(inner, "add", "f")
    _run_git(inner, "commit", "-m", "c")

    client = GitHarnessClient()

    repos = client.scan_repositories(
        ScanOptions(root_path=str(outer), use_cache=False, max_depth=20)
    )
    paths = {os.path.realpath(r["path"]) for r in repos}
    assert os.path.realpath(str(inner)) in paths


def test_subprocess_json_contract_smoke() -> None:
    """Guardrail: stdin JSON shape accepted by the Go CLI."""
    root = Path(__file__).resolve().parents[3]
    cli = os.environ.get("GIT_HARNESS_CLI", "").strip()
    cmd = [cli] if cli else ["go", "run", "./cmd/git-harness-cli"]
    if cli and not Path(cli).is_file() and shutil.which(cli) is None:
        # Relative path from repo root (typical in CI)
        cmd = [str((root / cli).resolve())]
    proc = subprocess.run(
        cmd,
        cwd=root,
        input=json.dumps({"op": "safety_security_notice"}),
        text=True,
        encoding="utf-8",
        errors="replace",
        capture_output=True,
        check=False,
        timeout=120,
    )
    assert proc.returncode == 0, proc.stderr
    stdout = (proc.stdout or "").strip()
    body = json.loads(stdout)
    assert body["ok"] is True
    assert "notice" in body and len(body["notice"]) > 0
