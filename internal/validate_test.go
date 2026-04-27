package flagapi

import (
	"strings"
	"testing"
)

func TestNormalizeFlagInputAllowsCountryOnly(t *testing.T) {
	country := "IN"
	flag, err := normalizeFlagInput(createFlagRequest{
		Name:    "country only",
		Country: &country,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flag.Rule != "country:IN" {
		t.Fatalf("unexpected rule: %s", flag.Rule)
	}
}

func TestNormalizeFlagInputAllowsPercentageOnly(t *testing.T) {
	pct := 30
	flag, err := normalizeFlagInput(createFlagRequest{
		Name:       "pct only",
		Percentage: &pct,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flag.Rule != "pct:30" {
		t.Fatalf("unexpected rule: %s", flag.Rule)
	}
}

func TestNormalizeFlagInputAllowsCountryAndPercentage(t *testing.T) {
	country := "in"
	pct := 25
	flag, err := normalizeFlagInput(createFlagRequest{
		Name:       "combined",
		Country:    &country,
		Percentage: &pct,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if flag.Rule != "country_pct:IN:25" {
		t.Fatalf("unexpected rule: %s", flag.Rule)
	}
}

func TestNormalizeFlagInputRejectsWhenMissingBoth(t *testing.T) {
	_, err := normalizeFlagInput(createFlagRequest{Name: "missing rule"})
	if err == nil {
		t.Fatal("expected error but got nil")
	}
	if !strings.Contains(err.Error(), "at least one of country or percentage is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}
