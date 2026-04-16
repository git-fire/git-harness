package git

import (
	"context"
	"errors"
	"fmt"
)

// ErrRewriteAttemptsExceeded indicates a rewrite run remained dirty after all
// configured attempts were used.
var ErrRewriteAttemptsExceeded = errors.New("rewrite scenario exceeded maximum attempts")

// RewriteScenarioOptions configures detect -> intervene -> verify orchestration.
//
// The flow is:
//  1. Detect whether rewrite intervention is required.
//  2. Intervene when required.
//  3. Verify whether the repo is now clean.
//  4. Repeat until clean or attempts are exhausted.
type RewriteScenarioOptions struct {
	// MaxAttempts bounds total detect passes. Must be >= 1.
	MaxAttempts int

	// Detect reports whether intervention is needed for the attempt.
	// Return true when the repository still needs rewrite intervention.
	Detect func(ctx context.Context, attempt int) (needsIntervention bool, err error)

	// Intervene performs one rewrite intervention pass for the attempt.
	Intervene func(ctx context.Context, attempt int) error

	// Verify checks whether the repository is clean after intervention.
	// Return true when no further rewrite intervention is needed.
	Verify func(ctx context.Context, attempt int) (clean bool, err error)
}

// RewriteAttempt captures one detect pass in the orchestration loop.
type RewriteAttempt struct {
	Attempt           int  // 1-based pass number
	NeedsIntervention bool // Detect outcome
	Intervened        bool // Whether Intervene ran
	CleanAfterVerify  bool // Verify outcome when Intervene ran
}

// RewriteScenarioResult captures pass-by-pass outcomes.
type RewriteScenarioResult struct {
	Clean    bool
	Attempts int
	Passes   []RewriteAttempt
}

// RunRewriteScenario executes bounded detect -> intervene -> verify passes until
// the repository is clean or attempts are exhausted.
func RunRewriteScenario(ctx context.Context, opts RewriteScenarioOptions) (RewriteScenarioResult, error) {
	result := RewriteScenarioResult{}

	if ctx == nil {
		ctx = context.Background()
	}
	if opts.MaxAttempts < 1 {
		return result, fmt.Errorf("max attempts must be >= 1")
	}
	if opts.Detect == nil {
		return result, fmt.Errorf("detect callback is required")
	}
	if opts.Intervene == nil {
		return result, fmt.Errorf("intervene callback is required")
	}
	if opts.Verify == nil {
		return result, fmt.Errorf("verify callback is required")
	}

	for attempt := 1; attempt <= opts.MaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			result.Attempts = attempt - 1
			return result, fmt.Errorf("rewrite scenario cancelled before attempt %d: %w", attempt, err)
		}

		pass := RewriteAttempt{Attempt: attempt}

		needsIntervention, err := opts.Detect(ctx, attempt)
		if err != nil {
			result.Attempts = attempt
			result.Passes = append(result.Passes, pass)
			return result, fmt.Errorf("detect failed on attempt %d: %w", attempt, err)
		}
		pass.NeedsIntervention = needsIntervention

		if !needsIntervention {
			result.Attempts = attempt
			result.Clean = true
			result.Passes = append(result.Passes, pass)
			return result, nil
		}

		if err := ctx.Err(); err != nil {
			result.Attempts = attempt
			result.Passes = append(result.Passes, pass)
			return result, fmt.Errorf("rewrite scenario cancelled before intervene on attempt %d: %w", attempt, err)
		}

		if err := opts.Intervene(ctx, attempt); err != nil {
			result.Attempts = attempt
			result.Passes = append(result.Passes, pass)
			return result, fmt.Errorf("intervene failed on attempt %d: %w", attempt, err)
		}
		pass.Intervened = true

		if err := ctx.Err(); err != nil {
			result.Attempts = attempt
			result.Passes = append(result.Passes, pass)
			return result, fmt.Errorf("rewrite scenario cancelled before verify on attempt %d: %w", attempt, err)
		}

		clean, err := opts.Verify(ctx, attempt)
		if err != nil {
			result.Attempts = attempt
			result.Passes = append(result.Passes, pass)
			return result, fmt.Errorf("verify failed on attempt %d: %w", attempt, err)
		}
		pass.CleanAfterVerify = clean
		result.Passes = append(result.Passes, pass)

		if clean {
			result.Attempts = attempt
			result.Clean = true
			return result, nil
		}
	}

	result.Attempts = opts.MaxAttempts
	return result, fmt.Errorf("%w: attempts=%d", ErrRewriteAttemptsExceeded, opts.MaxAttempts)
}
