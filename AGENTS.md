# AGENTS.md

<!-- agentic-guardrails:start -->
## Agentic guardrails

These notes apply to humans and to automated agents (including Cloud Agents).

1. **Autonomous work comes first**  
   The default posture is to **finish useful work without unnecessary human gates**. Prefer actions that keep momentum (fetch, build, test, open PRs) over asking for permission for routine steps—while still honoring the hard safety rules below.

2. **Multi-repository work is encouraged**  
   When a task genuinely spans more than one codebase (services, libraries, docs, sibling CLIs), work across **multiple repositories or local clones** as needed. Open **one PR per repository**, keep each PR scoped, and **link related PRs** in every description so reviewers see the coordinated change.

3. **Stay on a fresh enough tip; rebase is optional**  
   Before substantive work: `git fetch origin`, check out your branch, then prefer **`git merge --ff-only origin/<branch>`** (or `git pull --ff-only` when upstream exists). If you cannot fast-forward, pause and use the repository’s normal merge or rebase workflow—**do not** silently work on a stale checkout. **`git pull --rebase`** is *ideal* when updating an active feature branch, but it is **not required**; a merge commit or team-standard flow is fine when it avoids needless churn.

4. **Force-push is never automatic and needs explicit human buy-in**  
   Do **not** run `git push --force`, `git push --force-with-lease`, or rewrite published history on your own. If you believe it might be warranted, **stop** and give the **human** explicit **reasoning**, **effects** on collaborators, CI, and open PRs, and **why** a force-push would be needed versus safer alternatives (new branch, revert, merge). Proceed **only** after they **explicitly approve** that exact repository and branch.

5. **Focused changes and verification**  
   Keep pull requests focused; run this repository’s standard build, test, and lint commands (see `README`, `Makefile`, or `CLAUDE.md`) before requesting review.

6. **Workflow shape is yours**  
   Using **git worktree** versus a single working directory is an **operator choice**; these docs do **not** require worktrees.

<!-- agentic-guardrails:end -->

---

## Cursor Cloud specific instructions

This repo is the `git-harness` Go library plus a CLI and polyglot (Python + Java) wrappers. Standard commands live in `README.md` and `.github/workflows/ci.yml`; only the non-obvious cloud caveats are noted here.

**Layers**
- Go core (`git/`, `safety/`): `go build ./...`, `go vet ./...`, `go test -race -count=1 ./...` (see `README.md`).
- CLI (`cmd/git-harness-cli`): a stdin/stdout JSON bridge. Build with `go build -o ./bin/git-harness-cli ./cmd/git-harness-cli`.
- Wrappers (`testkit/python`, `testkit/java`): both shell out to the CLI and require `GIT_HARNESS_CLI` to point at the built binary (e.g. `export GIT_HARNESS_CLI=/workspace/bin/git-harness-cli`). Build the CLI first.

**Python wrapper**: Ubuntu is PEP 668 managed, so deps live in a dedicated venv at `~/.venvs/git-harness` (created by the update script). Run tests/samples with that interpreter, from `testkit/python`: `"$HOME/.venvs/git-harness/bin/python" -m pytest tests/ -q` and `-m samples.smoke_repo_flow` / `-m samples.smoke_safety_flow`. Note CI calls bare `python`, but only `python3` exists here.

**Java wrapper**: needs Maven + JDK 21 (baked into the snapshot). From `testkit/java`: `mvn test`, and samples via `mvn -Dtest=SampleRepoFlowSmoke,SampleSafetyFlowSmoke test`. First run populates the `~/.m2` cache.

**Known cloud-only test failure**: `git/command_test.go::TestFetchRemote_UnauthenticatedHTTPSFailsWithoutPrompt` fails in this environment. The cloud agent configures a global `url.https://x-access-token:<token>@github.com/.insteadOf` rewrite, so the "unauthenticated" fetch is actually authenticated and GitHub returns `Repository not found` instead of the expected credential-prompt error. This is an environment artifact, not a code bug; it passes in clean CI. Skip it with `-skip 'TestFetchRemote_UnauthenticatedHTTPSFailsWithoutPrompt'` when validating locally.


