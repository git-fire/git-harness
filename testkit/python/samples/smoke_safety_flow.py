from __future__ import annotations

from git_harness import GitHarnessClient


def main() -> int:
    client = GitHarnessClient()
    token = "ghp_" + ("a" * 36)
    out = client.safety_sanitize_text(f"export TOKEN={token}")
    if token in out:
        raise RuntimeError("expected full token to be redacted")
    if "[REDACTED]" not in out:
        raise RuntimeError("expected redaction marker in sanitized output")
    notice = client.safety_security_notice()
    if len(notice) < 10:
        raise RuntimeError("expected security notice body")
    print("python sample safety flow: OK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
