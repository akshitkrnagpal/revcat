package publish

import "testing"

func TestHashJSONStable(t *testing.T) {
	a := map[string]any{"a": 1, "b": 2, "c": []any{1, 2, 3}}
	b := map[string]any{"c": []any{1, 2, 3}, "b": 2, "a": 1}
	ah, _ := hashJSON(a)
	bh, _ := hashJSON(b)
	if ah != bh {
		t.Fatalf("want equal hashes for reordered maps, got %s vs %s", ah, bh)
	}
}

func TestHashJSONDifferent(t *testing.T) {
	a := map[string]any{"a": 1}
	b := map[string]any{"a": 2}
	ah, _ := hashJSON(a)
	bh, _ := hashJSON(b)
	if ah == bh {
		t.Fatal("want different hashes for different values")
	}
}

func TestPlanStepsCurrentOnly(t *testing.T) {
	steps := planSteps(false, true, nil, func() (map[string]any, error) {
		t.Fatal("paywall fetcher should not be called when newPaywall is nil")
		return nil, nil
	})
	if len(steps) != 1 {
		t.Fatalf("want 1 step, got %d", len(steps))
	}
	if steps[0].kind != stepSetCurrent {
		t.Fatalf("want stepSetCurrent, got %v", steps[0].kind)
	}
}

func TestPlanStepsAlreadyCurrent(t *testing.T) {
	steps := planSteps(true, true, nil, func() (map[string]any, error) {
		return nil, nil
	})
	if len(steps) != 0 {
		t.Fatalf("want 0 steps, got %d", len(steps))
	}
}

func TestPlanStepsPaywallSkipWhenIdentical(t *testing.T) {
	body := map[string]any{"hello": "world"}
	steps := planSteps(true, true, body, func() (map[string]any, error) {
		return body, nil
	})
	if len(steps) != 0 {
		t.Fatalf("want 0 steps when paywall matches, got %d", len(steps))
	}
}

func TestPlanStepsPaywallChange(t *testing.T) {
	steps := planSteps(true, true, map[string]any{"v": 2}, func() (map[string]any, error) {
		return map[string]any{"v": 1}, nil
	})
	if len(steps) != 1 || steps[0].kind != stepPaywall {
		t.Fatalf("want one paywall step, got %+v", steps)
	}
}
