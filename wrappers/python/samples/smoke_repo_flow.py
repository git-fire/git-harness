from __future__ import annotations

import os
import subprocess
import tempfile
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


def _git_init(repo: Path) -> None:
    _run_git(repo, "init")
    _run_git(repo, "config", "user.email", "harness-sample@example.com")
    _run_git(repo, "config", "user.name", "git-harness sample")


def main() -> int:
    client = GitHarnessClient()
    with tempfile.TemporaryDirectory(prefix="git-harness-py-repo-") as tmp:
        base = Path(tmp)
        remote = base / "origin.git"
        local = base / "local"
        remote.mkdir()
        local.mkdir()

        _run_git(remote, "init", "--bare")

        _git_init(local)
        (local / "README.md").write_text("hello\n", encoding="utf-8")
        _run_git(local, "add", "README.md")
        _run_git(local, "commit", "-m", "init")
        branch = client.get_current_branch(str(local))

        _run_git(local, "remote", "add", "origin", str(remote.resolve()))
        _run_git(local, "push", "-u", "origin", branch)

        local_sha = client.get_commit_sha(str(local), branch)
        out = subprocess.run(
            ["git", "rev-parse", branch],
            cwd=remote,
            check=True,
            capture_output=True,
            text=True,
            encoding="utf-8",
            errors="replace",
        ).stdout.strip()
        if local_sha != out:
            raise RuntimeError(f"sha mismatch local={local_sha} remote={out}")

        repos = client.scan_repositories(
            ScanOptions(root_path=str(base), use_cache=False, max_depth=30)
        )
        # macOS often differs between symlinked paths (/var vs /private/var); compare real paths.
        local_key = Path(os.path.realpath(local))
        if not any(Path(os.path.realpath(r["path"])) == local_key for r in repos):
            raise RuntimeError("scan_repositories did not find local repo")

        print("python sample repo flow: OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
