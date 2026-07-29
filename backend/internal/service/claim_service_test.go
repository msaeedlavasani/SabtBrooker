package service

import (
	"testing"
)

// TestGenerateLegalAdvice_Ownership verifies advice for ownership claims
func TestGenerateLegalAdvice_Ownership(t *testing.T) {
	advice := generateLegalAdvice("ownership", "land_and_building")

	if advice.Action != "طرح دعوای اثبات مالکیت در مراجع قضایی" {
		t.Errorf("unexpected action: %s", advice.Action)
	}
	if len(advice.References) != 2 {
		t.Errorf("expected 2 references, got %d", len(advice.References))
	}
	if advice.Confidence != 0.85 {
		t.Errorf("expected confidence 0.85, got %f", advice.Confidence)
	}
}

// TestGenerateLegalAdvice_Easement verifies advice for easement claims
func TestGenerateLegalAdvice_Easement(t *testing.T) {
	advice := generateLegalAdvice("easement", "")

	if advice.Action != "طرح دعوای اثبات حق ارتفاق در مراجع قضایی" {
		t.Errorf("unexpected action: %s", advice.Action)
	}
	if advice.Confidence != 0.80 {
		t.Errorf("expected confidence 0.80, got %f", advice.Confidence)
	}
}

// TestGenerateLegalAdvice_Default verifies fallback advice
func TestGenerateLegalAdvice_Default(t *testing.T) {
	advice := generateLegalAdvice("", "")

	if advice.Confidence != 0.70 {
		t.Errorf("expected confidence 0.70 for default, got %f", advice.Confidence)
	}
	if len(advice.References) == 0 {
		t.Error("default advice should have at least one reference")
	}
}
