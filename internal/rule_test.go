package flagapi

import (
	"fmt"
	"strings"
	"testing"
)

func TestParseFlagRuleCountry(t *testing.T) {
	rule, err := ParseFlagRule("country:in")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.TypeName != RuleTypeCountry {
		t.Fatalf("unexpected type: %s", rule.TypeName)
	}
	if rule.Country != "IN" {
		t.Fatalf("unexpected country: %s", rule.Country)
	}
}

func TestParseFlagRulePercentage(t *testing.T) {
	rule, err := ParseFlagRule("pct:25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.TypeName != RuleTypePercentage {
		t.Fatalf("unexpected type: %s", rule.TypeName)
	}
	if rule.Percentage != 25 {
		t.Fatalf("unexpected percentage: %d", rule.Percentage)
	}
}

func TestParseFlagRuleCountryPercentage(t *testing.T) {
	rule, err := ParseFlagRule("country_pct:in:25")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rule.TypeName != RuleTypeCountryPercentage {
		t.Fatalf("unexpected type: %s", rule.TypeName)
	}
	if rule.Country != "IN" {
		t.Fatalf("unexpected country: %s", rule.Country)
	}
	if rule.Percentage != 25 {
		t.Fatalf("unexpected percentage: %d", rule.Percentage)
	}
}

func TestPercentageEnabledBoundaries(t *testing.T) {
	key := []byte("test-key")
	if PercentageEnabled("flag-1", "user-1", 0, key) {
		t.Fatal("0% should never be active")
	}
	if !PercentageEnabled("flag-1", "user-1", 100, key) {
		t.Fatal("100% should always be active")
	}
}

func TestPercentageEnabledDeterministic(t *testing.T) {
	key := []byte("test-key")
	first := PercentageEnabled("flag-rollout", "user-42", 37, key)
	for i := 0; i < 50; i++ {
		next := PercentageEnabled("flag-rollout", "user-42", 37, key)
		if next != first {
			t.Fatalf("non-deterministic result at iteration %d", i)
		}
	}
}

func TestPercentageEnabledMonotonic(t *testing.T) {
	key := []byte("test-key")
	for i := range 500 {
		userID := userIDForTest(i)
		at25 := PercentageEnabled("flag-x", userID, 25, key)
		at30 := PercentageEnabled("flag-x", userID, 30, key)
		if at25 && !at30 {
			t.Fatalf("monotonicity violated for %s", userID)
		}
	}
}

func userIDForTest(i int) string {
	return fmt.Sprintf("user-%d", i)
}

func TestNewUserIDPrefixAndFormat(t *testing.T) {
	id, err := newUserID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(id, "usr_") {
		t.Fatalf("unexpected prefix: %s", id)
	}
	if len(id) <= len("usr_") {
		t.Fatalf("id too short: %s", id)
	}
}

func TestNewFlagIDPrefixAndFormat(t *testing.T) {
	id, err := newFlagID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.HasPrefix(id, "flg_") {
		t.Fatalf("unexpected prefix: %s", id)
	}
	if len(id) <= len("flg_") {
		t.Fatalf("id too short: %s", id)
	}
}
