package bootstrap

import (
	"strings"
	"testing"
)

func TestResolvePolicy_FlagOverridesEnv(t *testing.T) {
	res, err := ResolvePolicy("keep", "overwrite", true)
	if err != nil {
		t.Fatalf("ResolvePolicy error: %v", err)
	}
	if res.Value != PolicyKeep {
		t.Fatalf("expected keep policy, got %s", res.Value)
	}
	if res.Source != PolicySourceFlag {
		t.Fatalf("expected source flag, got %s", res.Source)
	}
}

func TestResolvePolicy_EnvUsedWhenFlagEmpty(t *testing.T) {
	res, err := ResolvePolicy("", "overwrite", true)
	if err != nil {
		t.Fatalf("ResolvePolicy error: %v", err)
	}
	if res.Value != PolicyOverwrite {
		t.Fatalf("expected overwrite policy, got %s", res.Value)
	}
	if res.Source != PolicySourceEnv {
		t.Fatalf("expected env source, got %s", res.Source)
	}
}

func TestResolvePolicy_DefaultPromptInteractive(t *testing.T) {
	res, err := ResolvePolicy("", "", true)
	if err != nil {
		t.Fatalf("ResolvePolicy error: %v", err)
	}
	if res.Value != PolicyPrompt {
		t.Fatalf("expected prompt policy, got %s", res.Value)
	}
	if res.Source != PolicySourceDefault {
		t.Fatalf("expected default source, got %s", res.Source)
	}
}

func TestResolvePolicy_InvalidValue(t *testing.T) {
	if _, err := ResolvePolicy("invalid", "", true); err == nil {
		t.Fatalf("expected error for invalid policy value")
	}
}

func TestResolvePolicy_NonInteractiveDefaultsToKeep(t *testing.T) {
	res, err := ResolvePolicy("", "", false)
	if err != nil {
		t.Fatalf("non-interactive resolution must not fail, got %v", err)
	}
	if res.Value != PolicyKeep {
		t.Fatalf("expected keep without a terminal, got %s", res.Value)
	}
}

func TestParsePolicy_ErrorNamesValue(t *testing.T) {
	_, err := ParsePolicy("replace")
	if err == nil || !strings.Contains(err.Error(), "replace") {
		t.Fatalf("expected error naming the bad value, got %v", err)
	}
}
