package classifier

import (
	"testing"
)

func TestMapKeywordToIntent_ConceptQueries(t *testing.T) {
	queries := []struct {
		query string
		want  IntentType
	}{
		{"What is photosynthesis?", IntentConcept},
		{"Explain Newton's laws", IntentConcept},
		{"Define momentum", IntentConcept},
		{"Tell me about the solar system", IntentConcept},
		{"Who was Abraham Lincoln?", IntentConcept},
		{"When did World War II end?", IntentConcept},
		{"Why does rain fall?", IntentConcept},
		{"How does a car engine work?", IntentConcept},
		{"Describe the process of evolution", IntentConcept},
		{"Understand the concept of energy", IntentConcept},
	}

	for _, tt := range queries {
		t.Run(tt.query, func(t *testing.T) {
			got := MapKeywordToIntent(tt.query)
			if got != tt.want {
				t.Errorf("MapKeywordToIntent(%q) = %s, want %s", tt.query, got, tt.want)
			}
		})
	}
}

func TestMapKeywordToIntent_NumericalQueries(t *testing.T) {
	queries := []struct {
		query string
		want  IntentType
	}{
		{"Solve 2x + 5 = 15", IntentNumerical},
		{"Calculate the area of a circle", IntentNumerical},
		{"Find the value of x", IntentNumerical},
		{"What is 15% of 200?", IntentNumerical},
		{"Evaluate 2^3 + 3^2", IntentNumerical},
		{"Compute the derivative", IntentNumerical},
		{"Simplify the expression", IntentNumerical},
		{"Math problem: find the perimeter", IntentNumerical},
	}

	for _, tt := range queries {
		t.Run(tt.query, func(t *testing.T) {
			got := MapKeywordToIntent(tt.query)
			if got != tt.want {
				t.Errorf("MapKeywordToIntent(%q) = %s, want %s", tt.query, got, tt.want)
			}
		})
	}
}

func TestMapKeywordToIntent_FactCheckQueries(t *testing.T) {
	queries := []struct {
		query string
		want  IntentType
	}{
		{"True or false: Earth is flat", IntentFactCheck},
		{"Verify if 2+2=5", IntentFactCheck},
		{"Is it correct that water boils at 90C?", IntentFactCheck},
		{"Fact check: humans use 10% of brain", IntentFactCheck},
		{"Compare mitosis and meiosis", IntentFactCheck},
		{"Difference between mass and weight", IntentFactCheck},
		{"Is Shakespeare the author of Hamlet?", IntentFactCheck},
		{"Check if photosynthesis produces oxygen", IntentFactCheck},
	}

	for _, tt := range queries {
		t.Run(tt.query, func(t *testing.T) {
			got := MapKeywordToIntent(tt.query)
			if got != tt.want {
				t.Errorf("MapKeywordToIntent(%q) = %s, want %s", tt.query, got, tt.want)
			}
		})
	}
}

func TestMapKeywordToIntent_StrategyQueries(t *testing.T) {
	queries := []struct {
		query string
		want  IntentType
	}{
		{"How to solve a quadratic equation?", IntentStrategy},
		{"Steps to balance a chemical equation", IntentStrategy},
		{"Process for conducting an experiment", IntentStrategy},
		{"Method for finding the GCD", IntentStrategy},
		{"Guide me through long division", IntentStrategy},
		{"Walk me through the steps", IntentStrategy},
		{"Instructions for building a circuit", IntentStrategy},
	}

	for _, tt := range queries {
		t.Run(tt.query, func(t *testing.T) {
			got := MapKeywordToIntent(tt.query)
			if got != tt.want {
				t.Errorf("MapKeywordToIntent(%q) = %s, want %s", tt.query, got, tt.want)
			}
		})
	}
}

func TestMapLegacyToFourCategory_AllMappings(t *testing.T) {
	tests := []struct {
		legacy IntentType
		want   IntentType
	}{
		{IntentInformationSeeking, IntentConcept},
		{IntentDefinition, IntentConcept},
		{IntentClarification, IntentConcept},
		{IntentConversational, IntentConcept},
		{IntentProblemSolving, IntentNumerical},
		{IntentExamPrep, IntentNumerical},
		{IntentContentRetrieval, IntentFactCheck},
		{IntentComparison, IntentFactCheck},
		{IntentProcedural, IntentStrategy},
	}

	for _, tt := range tests {
		t.Run(string(tt.legacy), func(t *testing.T) {
			got := MapLegacyToFourCategory(tt.legacy)
			if got != tt.want {
				t.Errorf("MapLegacyToFourCategory(%s) = %s, want %s", tt.legacy, got, tt.want)
			}
		})
	}
}

func TestClassifyIntentFromKeywords(t *testing.T) {
	// Test that the function is a simple wrapper
	got := ClassifyIntentFromKeywords("What is gravity?")
	if got != IntentConcept {
		t.Errorf("ClassifyIntentFromKeywords() = %s, want %s", got, IntentConcept)
	}
}

func TestContainsIntentKeyword(t *testing.T) {
	tests := []struct {
		query   string
		keyword string
		want    bool
	}{
		{"what is photosynthesis", "what is", true},
		{"what_is_photosynthesis", "what is", true},
		{"what-is-photosynthesis", "what is", true},
		{"how to solve", "how to", true},
		{"solve equation", "how to", false},
		{"", "what is", false},
	}

	for _, tt := range tests {
		t.Run(tt.query+"_"+tt.keyword, func(t *testing.T) {
			got := containsIntentKeyword(tt.query, tt.keyword)
			if got != tt.want {
				t.Errorf("containsIntentKeyword(%q, %q) = %v, want %v", tt.query, tt.keyword, got, tt.want)
			}
		})
	}
}
