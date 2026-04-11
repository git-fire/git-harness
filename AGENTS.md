## Cursor Cloud specific instructions

This is a **documentation-only planning repository** for the `git-harness` Go library extraction project.

### Repository contents

The repo contains only markdown specification files (`CURSOR_ULTRA_PLAN.md`, `SPEC.md`, and specs under `mnt/user-data/outputs/git-harness-plan/specs/`). There is no source code, no dependency manifests, no build system, and no runnable services.

### Development environment

- No dependencies need to be installed — there are no `go.mod`, `package.json`, `requirements.txt`, or similar files.
- No build, lint, or test commands apply.
- No services need to be started.
- The only tool needed is a text editor for markdown files and `git` for version control.

### What this repo plans

The specs describe extracting reusable git primitives from `git-fire` and `git-fcuk` into a new Go module `github.com/git-fire/git-harness`, with Python and Java wrappers. The actual code lives in separate repositories (`git-fire/git-fire`, `git-fire/git-fcuk`, `git-fire/git-testkit`), none of which are present in this workspace.
