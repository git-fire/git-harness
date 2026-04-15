from __future__ import annotations

import json
import os
import subprocess
from dataclasses import dataclass, field
from pathlib import Path
from typing import Any

_CLI_TIMEOUT_SECONDS = 120


def _repo_root() -> Path:
    # wrappers/python/git_harness/cli.py -> repo root
    return Path(__file__).resolve().parents[3]


def _cli_cmd() -> list[str]:
    cli = os.environ.get("GIT_HARNESS_CLI", "").strip()
    if cli:
        cli_path = Path(cli)
        if not cli_path.is_absolute():
            cli_path = _repo_root() / cli_path
        return [str(cli_path)]
    return ["go", "run", "./cmd/git-harness-cli"]


def _call(op: str, **payload: Any) -> dict[str, Any]:
    request = {"op": op, **payload}
    try:
        proc = subprocess.run(
            _cli_cmd(),
            cwd=_repo_root(),
            input=json.dumps(request),
            text=True,
            capture_output=True,
            check=False,
            timeout=_CLI_TIMEOUT_SECONDS,
        )
    except subprocess.TimeoutExpired as exc:
        raise RuntimeError(
            f"git-harness-cli timed out after {_CLI_TIMEOUT_SECONDS}s (op={op})"
        ) from exc
    stdout = (proc.stdout or "").strip()
    stderr = (proc.stderr or "").strip()
    if proc.returncode != 0:
        if stdout:
            try:
                response = json.loads(stdout)
            except json.JSONDecodeError:
                response = {}
            if not response.get("ok", True) and response.get("error"):
                raise RuntimeError(str(response["error"]))
        raise RuntimeError(
            f"git-harness-cli exited {proc.returncode}: {stderr}; stdout: {stdout}"
        )

    try:
        response = json.loads(stdout)
    except json.JSONDecodeError as exc:
        raise RuntimeError(
            f"invalid JSON from git-harness-cli: {stdout!r}; stderr: {stderr}"
        ) from exc
    if not response.get("ok", False):
        raise RuntimeError(response.get("error", "unknown git-harness-cli error"))
    return response


@dataclass(slots=True)
class ScanOptions:
    root_path: str = "."
    exclude: list[str] | None = None
    max_depth: int = 0
    use_cache: bool | None = None
    cache_file: str = ""
    cache_ttl: str = ""
    workers: int = 0
    known_paths: dict[str, bool] | None = None
    disable_scan: bool = False

    def to_payload(self) -> dict[str, Any]:
        d: dict[str, Any] = {
            "rootPath": self.root_path,
            "disableScan": self.disable_scan,
        }
        if self.exclude is not None:
            d["exclude"] = self.exclude
        if self.max_depth > 0:
            d["maxDepth"] = self.max_depth
        if self.use_cache is not None:
            d["useCache"] = self.use_cache
        if self.cache_file:
            d["cacheFile"] = self.cache_file
        if self.cache_ttl:
            d["cacheTTL"] = self.cache_ttl
        if self.workers > 0:
            d["workers"] = self.workers
        if self.known_paths is not None:
            d["knownPaths"] = self.known_paths
        return d


class GitHarnessClient:
    def scan_repositories(self, options: ScanOptions | None = None) -> list[dict[str, Any]]:
        opts = options or ScanOptions()
        res = _call("scan_repositories", scanOptions=opts.to_payload())
        return list(res.get("repositories", []))

    def analyze_repository(self, repo_path: str | Path) -> dict[str, Any]:
        res = _call("analyze_repository", repoPath=str(repo_path))
        return dict(res["repository"])

    def is_dirty(self, repo_path: str) -> bool:
        res = _call("git_is_dirty", repoPath=repo_path)
        return bool(res["dirty"])

    def get_current_branch(self, repo_path: str) -> str:
        res = _call("git_get_current_branch", repoPath=repo_path)
        return str(res["branch"])

    def get_commit_sha(self, repo_path: str, ref: str) -> str:
        res = _call("git_get_commit_sha", repoPath=repo_path, ref=ref)
        return str(res["sha"])

    def list_local_branches(self, repo_path: str) -> list[str]:
        res = _call("git_list_local_branches", repoPath=repo_path)
        return [str(b) for b in res.get("branches", [])]

    def list_remote_branches(self, repo_path: str, remote: str) -> list[str]:
        res = _call("git_list_remote_branches", repoPath=repo_path, remote=remote)
        return [str(b) for b in res.get("branches", [])]

    def ref_is_ancestor(self, repo_path: str, ancestor_ref: str, descendant_ref: str) -> bool:
        res = _call(
            "git_ref_is_ancestor",
            repoPath=repo_path,
            ancestorRef=ancestor_ref,
            descendantRef=descendant_ref,
        )
        return bool(res["isAncestor"])

    def detect_conflict(self, repo_path: str, branch: str, remote: str) -> tuple[bool, str, str]:
        res = _call("git_detect_conflict", repoPath=repo_path, branch=branch, remote=remote)
        return bool(res["hasConflict"]), str(res.get("localSHA", "")), str(res.get("remoteSHA", ""))

    def has_staged_changes(self, repo_path: str) -> bool:
        res = _call("git_has_staged_changes", repoPath=repo_path)
        return bool(res["staged"])

    def has_unstaged_changes(self, repo_path: str) -> bool:
        res = _call("git_has_unstaged_changes", repoPath=repo_path)
        return bool(res["unstaged"])

    def get_uncommitted_files(self, repo_path: str) -> list[str]:
        res = _call("git_get_uncommitted_files", repoPath=repo_path)
        return [str(p) for p in res.get("paths", [])]

    def list_worktrees(self, repo_path: str) -> list[dict[str, Any]]:
        res = _call("git_list_worktrees", repoPath=repo_path)
        return list(res.get("worktrees", []))

    def auto_commit_dirty(
        self,
        repo_path: str,
        *,
        message: str = "",
        add_all: bool = False,
        use_dual_branch: bool = True,
        return_to_original: bool = True,
    ) -> None:
        _call(
            "git_auto_commit_dirty",
            repoPath=repo_path,
            message=message,
            addAll=add_all,
            useDualBranch=use_dual_branch,
            returnToOriginal=return_to_original,
        )

    def auto_commit_dirty_with_strategy(
        self,
        repo_path: str,
        *,
        message: str = "",
        add_all: bool = False,
        use_dual_branch: bool = True,
        return_to_original: bool = True,
    ) -> dict[str, Any]:
        return _call(
            "git_auto_commit_dirty_with_strategy",
            repoPath=repo_path,
            message=message,
            addAll=add_all,
            useDualBranch=use_dual_branch,
            returnToOriginal=return_to_original,
        )

    def create_fire_branch(self, repo_path: str, original_branch: str, local_sha: str) -> str:
        res = _call(
            "git_create_fire_branch",
            repoPath=repo_path,
            originalBranch=original_branch,
            localSHA=local_sha,
        )
        return str(res["fireBranch"])

    def fetch_remote(self, repo_path: str, remote: str) -> None:
        _call("git_fetch_remote", repoPath=repo_path, remote=remote)

    def push_branch(self, repo_path: str, remote: str, branch: str) -> None:
        _call("git_push_branch", repoPath=repo_path, remote=remote, branch=branch)

    def push_all_branches(self, repo_path: str, remote: str) -> None:
        _call("git_push_all_branches", repoPath=repo_path, remote=remote)

    def safety_sanitize_text(self, text: str) -> str:
        res = _call("safety_sanitize_text", text=text)
        return str(res["text"])

    def safety_recommended_gitignore_patterns(self) -> list[str]:
        res = _call("safety_recommended_gitignore_patterns")
        return [str(x) for x in res.get("lines", [])]

    def safety_security_notice(self) -> str:
        res = _call("safety_security_notice")
        return str(res["notice"])

    def safety_format_warning(self, files: list[dict[str, Any]]) -> str:
        res = _call("safety_format_warning", suspiciousFiles=files)
        return str(res["warning"])

    def safety_scan_files(self, repo_path: str, files: list[str]) -> list[dict[str, Any]]:
        res = _call("safety_scan_files", repoPath=repo_path, files=files)
        return list(res.get("suspiciousFiles", []))
