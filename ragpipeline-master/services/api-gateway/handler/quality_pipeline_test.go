package handler

import (
	"testing"
)

func TestParseWeekString(t *testing.T) {
	tests := []struct {
		name      string
		week      string
		wantYear  int
		wantWeek  int
		wantErr   bool
	}{
		{"valid week", "2026-W14", 2026, 14, false},
		{"valid week no dash", "2026W14", 2026, 14, false},
		{"week 1", "2026-W01", 2026, 1, false},
		{"week 53", "2026-W53", 2026, 53, false},
		{"invalid format", "invalid", 0, 0, true},
		{"empty string", "", 0, 0, true},
		{"week 0 invalid", "2026-W00", 0, 0, true},
		{"week 54 invalid", "2026-W54", 0, 0, true},
		{"lowercase w", "2026w14", 2026, 14, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			year, week, err := parseWeekString(tt.week)
			if tt.wantErr {
				if err == nil {
					t.Errorf("parseWeekString(%q) expected error, got nil", tt.week)
				}
			} else {
				if err != nil {
					t.Errorf("parseWeekString(%q) unexpected error: %v", tt.week, err)
				}
				if year != tt.wantYear {
					t.Errorf("parseWeekString(%q) year = %d, want %d", tt.week, year, tt.wantYear)
				}
				if week != tt.wantWeek {
					t.Errorf("parseWeekString(%q) week = %d, want %d", tt.week, week, tt.wantWeek)
				}
			}
		})
	}
}

func TestQualityPipeline_Constants(t *testing.T) {
	// Verify constants are reasonable
	if MaxReexplainAttempts != 3 {
		t.Errorf("MaxReexplainAttempts = %d, want 3", MaxReexplainAttempts)
	}
	if LowConfidenceThreshold != 0.6 {
		t.Errorf("LowConfidenceThreshold = %v, want 0.6", LowConfidenceThreshold)
	}
	if QualityFlaggedKeyPrefix == "" {
		t.Error("QualityFlaggedKeyPrefix is empty")
	}
	if QualityDatasetKeyPrefix == "" {
		t.Error("QualityDatasetKeyPrefix is empty")
	}
}
