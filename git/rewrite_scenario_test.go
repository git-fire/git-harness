package git

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestRunRewriteScenario_CleanOnFirstDetect(t *testing.T) {
	detectCalls := 0
	interveneCalls := 0
	verifyCalls := 0

	result, err := RunRewriteScenario(context.Background(), RewriteScenarioOptions{
		MaxAttempts: 3,
		Detect: func(_ context.Context, attempt int) (bool, error) {
			detectCalls++
			if attempt != 1 {
				t.Fatalf("attempt = %d, want 1", attempt)
			}
			return false, nil
		},
		Intervene: func(_ context.Context, _ int) error {
			interveneCalls++
			return nil
		},
		Verify: func(_ context.Context, _ int) (bool, error) {
			verifyCalls++
			return true, nil
		},
	})
	if err != nil {
		t.Fatalf("RunRewriteScenario() error = %v", err)
	}

	if !result.Clean {
		t.Fatalf("result.Clean = false, want true")
	}
	if result.Attempts != 1 {
		t.Fatalf("result.Attempts = %d, want 1", result.Attempts)
	}
	if len(result.Passes) != 1 {
		t.Fatalf("len(result.Passes) = %d, want 1", len(result.Passes))
	}
	pass := result.Passes[0]
	if pass.NeedsIntervention {
		t.Fatalf("pass.NeedsIntervention = true, want false")
	}
	if pass.Intervened {
		t.Fatalf("pass.Intervened = true, want false")
	}
	if pass.CleanAfterVerify {
		t.Fatalf("pass.CleanAfterVerify = true, want false when verify not run")
	}
	if detectCalls != 1 || interveneCalls != 0 || verifyCalls != 0 {
		t.Fatalf("calls = detect:%d intervene:%d verify:%d, want 1/0/0", detectCalls, interveneCalls, verifyCalls)
	}
}

func TestRunRewriteScenario_MultiPassUntilVerifyClean(t *testing.T) {
	detectCalls := 0
	interveneCalls := 0
	verifyCalls := 0

	result, err := RunRewriteScenario(context.Background(), RewriteScenarioOptions{
		MaxAttempts: 4,
		Detect: func(_ context.Context, _ int) (bool, error) {
			detectCalls++
			return true, nil
		},
		Intervene: func(_ context.Context, _ int) error {
			interveneCalls++
			return nil
		},
		Verify: func(_ context.Context, _ int) (bool, error) {
			verifyCalls++
			return verifyCalls == 2, nil
		},
	})
	if err != nil {
		t.Fatalf("RunRewriteScenario() error = %v", err)
	}

	if !result.Clean {
		t.Fatalf("result.Clean = false, want true")
	}
	if result.Attempts != 2 {
		t.Fatalf("result.Attempts = %d, want 2", result.Attempts)
	}
	if len(result.Passes) != 2 {
		t.Fatalf("len(result.Passes) = %d, want 2", len(result.Passes))
	}
	if result.Passes[0].CleanAfterVerify {
		t.Fatalf("pass 1 clean after verify = true, want false")
	}
	if !result.Passes[1].CleanAfterVerify {
		t.Fatalf("pass 2 clean after verify = false, want true")
	}
	if detectCalls != 2 || interveneCalls != 2 || verifyCalls != 2 {
		t.Fatalf("calls = detect:%d intervene:%d verify:%d, want 2/2/2", detectCalls, interveneCalls, verifyCalls)
	}
}

func TestRunRewriteScenario_RerunDetectUntilClean(t *testing.T) {
	detectCalls := 0
	interveneCalls := 0
	verifyCalls := 0

	result, err := RunRewriteScenario(context.Background(), RewriteScenarioOptions{
		MaxAttempts: 3,
		Detect: func(_ context.Context, _ int) (bool, error) {
			detectCalls++
			return detectCalls == 1, nil
		},
		Intervene: func(_ context.Context, _ int) error {
			interveneCalls++
			return nil
		},
		Verify: func(_ context.Context, _ int) (bool, error) {
			verifyCalls++
			return false, nil
		},
	})
	if err != nil {
		t.Fatalf("RunRewriteScenario() error = %v", err)
	}

	if !result.Clean {
		t.Fatalf("result.Clean = false, want true")
	}
	if result.Attempts != 2 {
		t.Fatalf("result.Attempts = %d, want 2", result.Attempts)
	}
	if len(result.Passes) != 2 {
		t.Fatalf("len(result.Passes) = %d, want 2", len(result.Passes))
	}
	if !result.Passes[0].NeedsIntervention || !result.Passes[0].Intervened {
		t.Fatalf("pass 1 should include intervention, got %+v", result.Passes[0])
	}
	if result.Passes[1].NeedsIntervention || result.Passes[1].Intervened {
		t.Fatalf("pass 2 should be clean detect with no intervention, got %+v", result.Passes[1])
	}
	if detectCalls != 2 || interveneCalls != 1 || verifyCalls != 1 {
		t.Fatalf("calls = detect:%d intervene:%d verify:%d, want 2/1/1", detectCalls, interveneCalls, verifyCalls)
	}
}

func TestRunRewriteScenario_ExceedsMaxAttempts(t *testing.T) {
	result, err := RunRewriteScenario(context.Background(), RewriteScenarioOptions{
		MaxAttempts: 3,
		Detect: func(_ context.Context, _ int) (bool, error) {
			return true, nil
		},
		Intervene: func(_ context.Context, _ int) error {
			return nil
		},
		Verify: func(_ context.Context, _ int) (bool, error) {
			return false, nil
		},
	})
	if err == nil {
		t.Fatal("RunRewriteScenario() error = nil, want max-attempts error")
	}
	if !errors.Is(err, ErrRewriteAttemptsExceeded) {
		t.Fatalf("error = %v, want ErrRewriteAttemptsExceeded", err)
	}
	if result.Clean {
		t.Fatalf("result.Clean = true, want false")
	}
	if result.Attempts != 3 {
		t.Fatalf("result.Attempts = %d, want 3", result.Attempts)
	}
	if len(result.Passes) != 3 {
		t.Fatalf("len(result.Passes) = %d, want 3", len(result.Passes))
	}
}

func TestRunRewriteScenario_CallbackErrors(t *testing.T) {
	t.Run("detect error", func(t *testing.T) {
		result, err := RunRewriteScenario(context.Background(), RewriteScenarioOptions{
			MaxAttempts: 2,
			Detect: func(_ context.Context, _ int) (bool, error) {
				return false, errors.New("boom-detect")
			},
			Intervene: func(_ context.Context, _ int) error {
				return nil
			},
			Verify: func(_ context.Context, _ int) (bool, error) {
				return true, nil
			},
		})
		if err == nil || !strings.Contains(err.Error(), "detect failed") {
			t.Fatalf("error = %v, want detect failure", err)
		}
		if result.Attempts != 1 {
			t.Fatalf("result.Attempts = %d, want 1", result.Attempts)
		}
	})

	t.Run("intervene error", func(t *testing.T) {
		result, err := RunRewriteScenario(context.Background(), RewriteScenarioOptions{
			MaxAttempts: 2,
			Detect: func(_ context.Context, _ int) (bool, error) {
				return true, nil
			},
			Intervene: func(_ context.Context, _ int) error {
				return errors.New("boom-intervene")
			},
			Verify: func(_ context.Context, _ int) (bool, error) {
				return true, nil
			},
		})
		if err == nil || !strings.Contains(err.Error(), "intervene failed") {
			t.Fatalf("error = %v, want intervene failure", err)
		}
		if result.Attempts != 1 {
			t.Fatalf("result.Attempts = %d, want 1", result.Attempts)
		}
		if len(result.Passes) != 1 || !result.Passes[0].NeedsIntervention {
			t.Fatalf("passes = %+v, want one pass with intervention required", result.Passes)
		}
	})

	t.Run("verify error", func(t *testing.T) {
		result, err := RunRewriteScenario(context.Background(), RewriteScenarioOptions{
			MaxAttempts: 2,
			Detect: func(_ context.Context, _ int) (bool, error) {
				return true, nil
			},
			Intervene: func(_ context.Context, _ int) error {
				return nil
			},
			Verify: func(_ context.Context, _ int) (bool, error) {
				return false, errors.New("boom-verify")
			},
		})
		if err == nil || !strings.Contains(err.Error(), "verify failed") {
			t.Fatalf("error = %v, want verify failure", err)
		}
		if result.Attempts != 1 {
			t.Fatalf("result.Attempts = %d, want 1", result.Attempts)
		}
		if len(result.Passes) != 1 || !result.Passes[0].Intervened {
			t.Fatalf("passes = %+v, want one pass with intervene=true", result.Passes)
		}
	})
}

func TestRunRewriteScenario_InvalidConfiguration(t *testing.T) {
	_, err := RunRewriteScenario(context.Background(), RewriteScenarioOptions{
		MaxAttempts: 0,
		Detect: func(_ context.Context, _ int) (bool, error) {
			return false, nil
		},
		Intervene: func(_ context.Context, _ int) error {
			return nil
		},
		Verify: func(_ context.Context, _ int) (bool, error) {
			return true, nil
		},
	})
	if err == nil || !strings.Contains(err.Error(), "max attempts") {
		t.Fatalf("error = %v, want max-attempts validation error", err)
	}

	_, err = RunRewriteScenario(context.Background(), RewriteScenarioOptions{
		MaxAttempts: 1,
		Intervene: func(_ context.Context, _ int) error {
			return nil
		},
		Verify: func(_ context.Context, _ int) (bool, error) {
			return true, nil
		},
	})
	if err == nil || !strings.Contains(err.Error(), "detect callback") {
		t.Fatalf("error = %v, want detect callback validation error", err)
	}
}

func TestRunRewriteScenario_CancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result, err := RunRewriteScenario(ctx, RewriteScenarioOptions{
		MaxAttempts: 2,
		Detect: func(_ context.Context, _ int) (bool, error) {
			return true, nil
		},
		Intervene: func(_ context.Context, _ int) error {
			return nil
		},
		Verify: func(_ context.Context, _ int) (bool, error) {
			return false, nil
		},
	})
	if err == nil {
		t.Fatal("RunRewriteScenario() error = nil, want cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunRewriteScenario() error = %v, want context.Canceled", err)
	}
	if result.Attempts != 0 {
		t.Fatalf("result.Attempts = %d, want 0", result.Attempts)
	}
}

func TestRunRewriteScenario_NilContextFallback(t *testing.T) {
	detectCalls := 0
	result, err := RunRewriteScenario(nil, RewriteScenarioOptions{
		MaxAttempts: 1,
		Detect: func(_ context.Context, attempt int) (bool, error) {
			detectCalls++
			if attempt != 1 {
				t.Fatalf("attempt = %d, want 1", attempt)
			}
			return false, nil
		},
		Intervene: func(_ context.Context, _ int) error {
			t.Fatal("Intervene should not run when detect reports clean")
			return nil
		},
		Verify: func(_ context.Context, _ int) (bool, error) {
			t.Fatal("Verify should not run when detect reports clean")
			return true, nil
		},
	})
	if err != nil {
		t.Fatalf("RunRewriteScenario() error = %v", err)
	}
	if !result.Clean {
		t.Fatalf("result.Clean = false, want true")
	}
	if detectCalls != 1 {
		t.Fatalf("detectCalls = %d, want 1", detectCalls)
	}
}

func TestRunRewriteScenario_CancelledAfterDetectSkipsIntervene(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	interveneCalls := 0
	verifyCalls := 0

	result, err := RunRewriteScenario(ctx, RewriteScenarioOptions{
		MaxAttempts: 2,
		Detect: func(_ context.Context, _ int) (bool, error) {
			cancel()
			return true, nil
		},
		Intervene: func(_ context.Context, _ int) error {
			interveneCalls++
			return nil
		},
		Verify: func(_ context.Context, _ int) (bool, error) {
			verifyCalls++
			return true, nil
		},
	})
	if err == nil {
		t.Fatal("RunRewriteScenario() error = nil, want cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunRewriteScenario() error = %v, want context.Canceled", err)
	}
	if !strings.Contains(err.Error(), "cancelled before intervene") {
		t.Fatalf("error = %v, want message about intervene phase", err)
	}
	if result.Attempts != 1 {
		t.Fatalf("result.Attempts = %d, want 1", result.Attempts)
	}
	if len(result.Passes) != 1 || !result.Passes[0].NeedsIntervention {
		t.Fatalf("passes = %+v, want one pass with detect outcome recorded", result.Passes)
	}
	if interveneCalls != 0 || verifyCalls != 0 {
		t.Fatalf("calls = intervene:%d verify:%d, want 0/0", interveneCalls, verifyCalls)
	}
}

func TestRunRewriteScenario_CancelledAfterInterveneSkipsVerify(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	verifyCalls := 0

	result, err := RunRewriteScenario(ctx, RewriteScenarioOptions{
		MaxAttempts: 2,
		Detect: func(_ context.Context, _ int) (bool, error) {
			return true, nil
		},
		Intervene: func(_ context.Context, _ int) error {
			cancel()
			return nil
		},
		Verify: func(_ context.Context, _ int) (bool, error) {
			verifyCalls++
			return true, nil
		},
	})
	if err == nil {
		t.Fatal("RunRewriteScenario() error = nil, want cancellation error")
	}
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("RunRewriteScenario() error = %v, want context.Canceled", err)
	}
	if !strings.Contains(err.Error(), "cancelled before verify") {
		t.Fatalf("error = %v, want message about verify phase", err)
	}
	if result.Attempts != 1 {
		t.Fatalf("result.Attempts = %d, want 1", result.Attempts)
	}
	if len(result.Passes) != 1 || !result.Passes[0].Intervened {
		t.Fatalf("passes = %+v, want one pass with intervene=true", result.Passes)
	}
	if verifyCalls != 0 {
		t.Fatalf("verifyCalls = %d, want 0", verifyCalls)
	}
}
