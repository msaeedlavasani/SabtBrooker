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

// TestBuildClaimFields verifies field construction
func TestBuildClaimFields(t *testing.T) {
	claimType := "ownership"
	hasGov := true

	input := UpdateClaimInput{
		ClaimType:           &claimType,
		HasGovernmentRights: &hasGov,
	}

	fields := buildClaimFields(input)

	if len(fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(fields))
	}
	if fields["claim_type"] != "ownership" {
		t.Errorf("expected claim_type=ownership, got %v", fields["claim_type"])
	}
	if fields["has_government_rights"] != true {
		t.Errorf("expected has_government_rights=true, got %v", fields["has_government_rights"])
	}
}

// TestBuildClaimFields_Empty verifies empty input produces empty map
func TestBuildClaimFields_Empty(t *testing.T) {
	fields := buildClaimFields(UpdateClaimInput{})
	if len(fields) != 0 {
		t.Errorf("expected 0 fields for empty input, got %d", len(fields))
	}
}

// TestBuildClaimFields_All verifies all fields are included
func TestBuildClaimFields_All(t *testing.T) {
	claimType := "ownership"
	ownership := "land_and_building"
	mainPlate := "123"
	subPlate := "456"
	section := "7"
	totalShare := 6
	partialShare := 2
	hasGov := false
	ref := "12345"
	method := "human"

	input := UpdateClaimInput{
		ClaimType:           &claimType,
		OwnershipType:       &ownership,
		MainPlateNumber:     &mainPlate,
		SubPlateNumber:      &subPlate,
		PlateSection:        &section,
		TotalShare:          &totalShare,
		PartialShare:        &partialShare,
		HasGovernmentRights: &hasGov,
		TreasuryPaymentRef:  &ref,
		LegalAdviceMethod:   &method,
	}

	fields := buildClaimFields(input)

	if len(fields) != 10 {
		t.Errorf("expected 10 fields, got %d: %v", len(fields), fields)
	}
}
