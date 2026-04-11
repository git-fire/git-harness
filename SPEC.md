# Spec 001 — git-harness Extraction Audit

**Branch:** `spec/001-git-harness-audit`
**Status:** Ready for implementation

## Goal
Produce a written classification of every internal symbol in `git-fire` (and overlapping code in
`git-fcuk`) that is a candidate for extraction into `github.com/git-fire/git-harness`.

## Output
`specs/001-git-harness-audit/audit.md` — a table + narrative covering every package scanned.

## Acceptance Criteria
- Every exported symbol in `internal/git`, `internal/safety`, `internal/executor`, `internal/repo`
  has an `EXTRACT` / `KEEP` / `BOUNDARY` classification with a one-line rationale.
- Any symbol with an external dependency (Cobra, Viper, Bubble Tea, Lipgloss, charmbracelet)
  is automatically `KEEP`.
- Any symbol with zero app-layer dependencies is `EXTRACT` unless it's dead code.
- `audit.md` includes a "Shared duplication" section listing symbols that exist in BOTH
  `git-fire` and `git-fcuk` — these are highest-priority extractions.
- No code is modified in this spec.

## Notes
- Use `grep -r "cobra\|viper\|bubbletea\|lipgloss\|charmbracelet"` as a quick filter.
- Cross-reference `git-fcuk/internal/git` against `git-fire/internal/git` for duplication.
