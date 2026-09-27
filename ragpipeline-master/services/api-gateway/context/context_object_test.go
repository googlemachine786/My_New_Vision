package context

import (
	"testing"
	"time"
)

func TestNewContextObject(t *testing.T) {
	ctx := NewContextObject("sess1", "user1")

	if ctx.SessionID != "sess1" {
		t.Errorf("expected session ID sess1, got %s", ctx.SessionID)
	}
	if ctx.UserID != "user1" {
		t.Errorf("expected user ID user1, got %s", ctx.UserID)
	}
	if ctx.PracticeScores == nil {
		t.Error("expected PracticeScores to be initialized")
	}
	if len(ctx.PracticeErrors) != 0 {
		t.Errorf("expected empty PracticeErrors, got %d items", len(ctx.PracticeErrors))
	}
}

func TestAddPracticeError(t *testing.T) {
	ctx := NewContextObject("sess1", "user1")

	ctx.AddPracticeError("algebra", "sign error")
	if len(ctx.PracticeErrors) != 1 {
		t.Fatalf("expected 1 error, got %d", len(ctx.PracticeErrors))
	}
	if ctx.PracticeErrors[0].Count != 1 {
		t.Errorf("expected count 1, got %d", ctx.PracticeErrors[0].Count)
	}

	// Add same error again
	ctx.AddPracticeError("algebra", "sign error")
	if len(ctx.PracticeErrors) != 1 {
		t.Fatalf("expected 1 error (deduplicated), got %d", len(ctx.PracticeErrors))
	}
	if ctx.PracticeErrors[0].Count != 2 {
		t.Errorf("expected count 2, got %d", ctx.PracticeErrors[0].Count)
	}

	// Add different error
	ctx.AddPracticeError("geometry", "angle calculation")
	if len(ctx.PracticeErrors) != 2 {
		t.Fatalf("expected 2 errors, got %d", len(ctx.PracticeErrors))
	}
}

func TestRecordPracticeScore(t *testing.T) {
	ctx := NewContextObject("sess1", "user1")

	ctx.RecordPracticeScore("algebra", 0.85)
	if score, ok := ctx.PracticeScores["algebra"]; !ok {
		t.Error("expected algebra score to exist")
	} else if score != 0.85 {
		t.Errorf("expected score 0.85, got %f", score)
	}
}

func TestIncrementTurns(t *testing.T) {
	ctx := NewContextObject("sess1", "user1")
	oldUpdated := ctx.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	ctx.IncrementTurns()

	if ctx.SessionTurns != 1 {
		t.Errorf("expected 1 turn, got %d", ctx.SessionTurns)
	}
	if !ctx.UpdatedAt.After(oldUpdated) {
		t.Error("expected UpdatedAt to be updated")
	}
}

func TestUpdateMethods(t *testing.T) {
	ctx := NewContextObject("sess1", "user1")

	ctx.UpdateBoard("CBSE")
	if ctx.Board != "CBSE" {
		t.Errorf("expected board CBSE, got %s", ctx.Board)
	}

	ctx.UpdateGrade("10")
	if ctx.Grade != "10" {
		t.Errorf("expected grade 10, got %s", ctx.Grade)
	}

	ctx.UpdateLanguage("English")
	if ctx.Language != "English" {
		t.Errorf("expected language English, got %s", ctx.Language)
	}

	ctx.UpdateCurrentChapter("Optics")
	if ctx.CurrentChapter != "Optics" {
		t.Errorf("expected chapter Optics, got %s", ctx.CurrentChapter)
	}

	ctx.UpdateCurrentSubject("Physics")
	if ctx.CurrentSubject != "Physics" {
		t.Errorf("expected subject Physics, got %s", ctx.CurrentSubject)
	}
}

func TestFormatErrorsForPrompt(t *testing.T) {
	errors := []PracticeError{
		{Topic: "algebra", ErrorMsg: "sign error", Count: 1},
		{Topic: "geometry", ErrorMsg: "angle calculation", Count: 3},
	}

	result := FormatErrorsForPrompt(errors)
	if result == "" {
		t.Error("expected non-empty string")
	}
}

func TestFormatErrorsForPrompt_Empty(t *testing.T) {
	result := FormatErrorsForPrompt(nil)
	if result != "No recent mistakes recorded." {
		t.Errorf("expected 'No recent mistakes recorded.', got %s", result)
	}
}

func TestFormatContextForPrompt(t *testing.T) {
	ctx := NewContextObject("sess1", "user1")
	ctx.UpdateBoard("CBSE")
	ctx.UpdateGrade("10")
	ctx.UpdateLanguage("English")
	ctx.UpdateCurrentSubject("Physics")
	ctx.UpdateCurrentChapter("Optics")

	result := FormatContextForPrompt(ctx)
	if result == "" {
		t.Error("expected non-empty context string")
	}
}

func TestFormatContextForPrompt_Nil(t *testing.T) {
	result := FormatContextForPrompt(nil)
	if result != "" {
		t.Errorf("expected empty string for nil context, got %s", result)
	}
}
