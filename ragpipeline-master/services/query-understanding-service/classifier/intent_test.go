package classifier

import (
	"testing"
)

// TestQuery represents a query with its expected intent label
type TestQuery struct {
	Query          string
	ExpectedIntent IntentType
	Category       string // "easy", "ambiguous", "multi-intent"
}

// testSet100 contains 100 queries for intent classification validation (Story F)
var testSet100 = []TestQuery{
	// Concept queries (25 queries) - what is, explain, define
	{Query: "What is photosynthesis?", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "Explain Newton's laws of motion", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "Define momentum in physics", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "What are the causes of World War I?", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "Tell me about the water cycle", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "What is the meaning of democracy?", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "Explain how cells divide", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "What is a metaphor in literature?", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "Describe the structure of an atom", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "What is the Pythagorean theorem?", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "Who was Mahatma Gandhi?", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "When did India gain independence?", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "Why does the Earth orbit the Sun?", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "How does photosynthesis work?", ExpectedIntent: IntentConcept, Category: "ambiguous"},
	{Query: "What is the difference between speed and velocity?", ExpectedIntent: IntentConcept, Category: "ambiguous"},
	{Query: "Understand the concept of gravity", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "Explain the theory of relativity simply", ExpectedIntent: IntentConcept, Category: "ambiguous"},
	{Query: "What is an ecosystem?", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "Define the term biodiversity", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "What is the periodic table?", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "Explain what a fraction is", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "What is the meaning of supply and demand?", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "Describe the human digestive system", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "What is a prime number?", ExpectedIntent: IntentConcept, Category: "easy"},
	{Query: "Explain the concept of force in physics", ExpectedIntent: IntentConcept, Category: "easy"},

	// Numerical queries (25 queries) - solve, calculate, find value
	{Query: "Solve the equation 2x + 5 = 15", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Calculate the area of a circle with radius 7", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Find the value of x in 3x - 9 = 0", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "What is 15 percent of 200?", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Simplify the expression 4x + 2x - 3x", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Evaluate 2^3 + 3^2", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Calculate the derivative of x^2 + 3x", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Solve for y: 5y + 10 = 35", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "What is the square root of 144?", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Compute the integral of 2x dx", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Find the HCF of 24 and 36", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Calculate the volume of a cube with side 5cm", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Solve the quadratic equation x^2 - 5x + 6 = 0", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "What is 25 divided by 0.5?", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Evaluate sin(30 degrees)", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Math problem: find the perimeter of a rectangle", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Calculate simple interest on 1000 at 5% for 2 years", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Solve the system of equations: x+y=10, x-y=2", ExpectedIntent: IntentNumerical, Category: "ambiguous"},
	{Query: "Find the slope of the line passing through (2,3) and (4,7)", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Calculate the mean of 5, 10, 15, 20, 25", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "What is the factorial of 6?", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Solve this algebra problem: 7x = 49", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Calculate the compound interest on 5000 at 10% for 3 years", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Find the value of pi to 2 decimal places", ExpectedIntent: IntentNumerical, Category: "easy"},
	{Query: "Test question: evaluate 12 x 12", ExpectedIntent: IntentNumerical, Category: "easy"},

	// Fact-check queries (25 queries) - true/false, verify, is it correct
	{Query: "Is it true that the Earth is flat?", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Verify if 2 + 2 = 5", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Is this correct: water boils at 90 degrees Celsius?", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "True or false: the Moon orbits the Earth", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Compare mitosis and meiosis", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "What is the difference between mass and weight?", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Is Shakespeare the author of Hamlet?", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Verify: the capital of France is London", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Fact check: humans use only 10% of their brain", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Is it accurate that light travels faster than sound?", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Did Newton discover gravity?", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Check if photosynthesis produces oxygen", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Is this statement correct: all mammals lay eggs?", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Verify that the chemical formula for water is H2O", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "True or false: the Sun rises in the east", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Is it true that India has 28 states?", ExpectedIntent: IntentFactCheck, Category: "ambiguous"},
	{Query: "Compare renewable and non-renewable energy sources", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "What is the difference between a virus and a bacteria?", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Verify: Abraham Lincoln was the 16th US president", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Is this correct: the speed of light is 3x10^8 m/s?", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Fact check: the Great Wall of China is visible from space", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Did World War II end in 1945?", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "Is it true that electrons orbit the nucleus?", ExpectedIntent: IntentFactCheck, Category: "ambiguous"},
	{Query: "Confirm whether DNA stands for Deoxyribonucleic Acid", ExpectedIntent: IntentFactCheck, Category: "easy"},
	{Query: "True or false: sound travels faster in water than air", ExpectedIntent: IntentFactCheck, Category: "easy"},

	// Strategy queries (25 queries) - how to, steps, process
	{Query: "How to solve a quadratic equation?", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "What are the steps to balance a chemical equation?", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "Explain the process of photosynthesis step by step", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "How do I write a paragraph?", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "Walk me through the long division process", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "What is the best way to study for exams?", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "Guide me through solving a physics problem", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "How to prepare for a math test?", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "Steps to write an essay", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "How do I calculate percentage increase?", ExpectedIntent: IntentStrategy, Category: "ambiguous"},
	{Query: "Process for conducting a science experiment", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "How to find the LCM of two numbers?", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "What is the method to solve simultaneous equations?", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "Instructions for building a simple circuit", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "How to approach reading comprehension questions?", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "Steps to draw a geometric figure", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "How do I write a lab report?", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "What is the procedure for titration?", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "How to memorize the periodic table effectively?", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "Walk through solving this integral step by step", ExpectedIntent: IntentStrategy, Category: "ambiguous"},
	{Query: "Method for finding the area of a triangle", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "How to write a book review?", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "Process to convert fractions to decimals", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "Steps to create a timeline for a history project", ExpectedIntent: IntentStrategy, Category: "easy"},
	{Query: "How to analyze a poem for literary devices?", ExpectedIntent: IntentStrategy, Category: "easy"},
}

// TestFourCategoryIntentClassification validates >90% precision on the 100-query test set (Story F)
func TestFourCategoryIntentClassification(t *testing.T) {
	totalQueries := len(testSet100)
	if totalQueries != 100 {
		t.Errorf("Test set has %d queries, expected 100", totalQueries)
	}

	// Count by category
	categoryCount := make(map[IntentType]int)
	for _, tq := range testSet100 {
		categoryCount[tq.ExpectedIntent]++
	}

	// Verify 25 queries per category
	for _, intent := range ValidIntents() {
		count := categoryCount[intent]
		if count != 25 {
			t.Errorf("Category %s has %d queries, expected 25", intent, count)
		}
	}

	// Test keyword-based classification precision
	correct := 0
	misclassified := make([]TestQuery, 0)

	for _, tq := range testSet100 {
		predicted := ClassifyIntentFromKeywords(tq.Query)
		if predicted == tq.ExpectedIntent {
			correct++
		} else {
			misclassified = append(misclassified, tq)
		}
	}

	precision := float64(correct) / float64(totalQueries) * 100

	t.Logf("=== Intent Classification Results ===")
	t.Logf("Total queries: %d", totalQueries)
	t.Logf("Correctly classified: %d", correct)
	t.Logf("Misclassified: %d", len(misclassified))
	t.Logf("Precision: %.1f%%", precision)

	// Validate >90% precision
	if precision < 90.0 {
		t.Errorf("Precision %.1f%% is below 90%% threshold", precision)
		t.Logf("=== Misclassified Queries ===")
		for _, mq := range misclassified {
			predicted := ClassifyIntentFromKeywords(mq.Query)
			t.Logf("  Query: %q", mq.Query)
			t.Logf("    Expected: %s, Got: %s (category: %s)", mq.ExpectedIntent, predicted, mq.Category)
		}
	}
}

// TestMapLegacyToFourCategory validates the legacy-to-4-category mapping (Story F)
func TestMapLegacyToFourCategory(t *testing.T) {
	tests := []struct {
		name     string
		legacy   IntentType
		expected IntentType
	}{
		{"information_seeking -> concept", IntentInformationSeeking, IntentConcept},
		{"definition -> concept", IntentDefinition, IntentConcept},
		{"clarification -> concept", IntentClarification, IntentConcept},
		{"conversational -> concept", IntentConversational, IntentConcept},
		{"problem_solving -> numerical", IntentProblemSolving, IntentNumerical},
		{"exam_prep -> numerical", IntentExamPrep, IntentNumerical},
		{"content_retrieval -> fact_check", IntentContentRetrieval, IntentFactCheck},
		{"comparison -> fact_check", IntentComparison, IntentFactCheck},
		{"procedural -> strategy", IntentProcedural, IntentStrategy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapLegacyToFourCategory(tt.legacy)
			if got != tt.expected {
				t.Errorf("MapLegacyToFourCategory(%s) = %s, want %s", tt.legacy, got, tt.expected)
			}
		})
	}
}

// TestMapKeywordToIntent validates keyword-based intent mapping (Story F)
func TestMapKeywordToIntent(t *testing.T) {
	tests := []struct {
		name     string
		query    string
		expected IntentType
	}{
		{"what is -> concept", "What is photosynthesis?", IntentConcept},
		{"explain -> concept", "Explain Newton's laws", IntentConcept},
		{"define -> concept", "Define momentum", IntentConcept},
		{"solve -> numerical", "Solve 2x + 5 = 15", IntentNumerical},
		{"calculate -> numerical", "Calculate the area", IntentNumerical},
		{"find value -> numerical", "Find the value of x", IntentNumerical},
		{"true or false -> fact_check", "True or false: Earth is flat", IntentFactCheck},
		{"verify -> fact_check", "Verify this statement", IntentFactCheck},
		{"is it correct -> fact_check", "Is it correct that 2+2=5?", IntentFactCheck},
		{"how to -> strategy", "How to solve equations", IntentStrategy},
		{"steps -> strategy", "Steps to balance an equation", IntentStrategy},
		{"process -> strategy", "Process for conducting experiment", IntentStrategy},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := MapKeywordToIntent(tt.query)
			if got != tt.expected {
				t.Errorf("MapKeywordToIntent(%q) = %s, want %s", tt.query, got, tt.expected)
			}
		})
	}
}

// TestValidIntents_FourCategories validates the 4 required categories exist
func TestValidIntents_FourCategories(t *testing.T) {
	intents := ValidIntents()

	if len(intents) != 4 {
		t.Errorf("ValidIntents() returned %d intents, expected exactly 4", len(intents))
	}

	expectedIntents := map[IntentType]bool{
		IntentConcept:   false,
		IntentNumerical: false,
		IntentFactCheck: false,
		IntentStrategy:  false,
	}

	for _, intent := range intents {
		if _, ok := expectedIntents[intent]; !ok {
			t.Errorf("Unexpected intent: %s", intent)
		}
		expectedIntents[intent] = true
	}

	for intent, found := range expectedIntents {
		if !found {
			t.Errorf("Missing required intent: %s", intent)
		}
	}
}

// TestIsValidIntent_FourCategories validates all 4 categories pass IsValidIntent
func TestIsValidIntent_FourCategories(t *testing.T) {
	for _, intent := range ValidIntents() {
		if !IsValidIntent(string(intent)) {
			t.Errorf("ValidIntent(%q) returned false, expected true", intent)
		}
	}

	// Also verify legacy intents still work for backward compatibility
	legacyIntents := []string{
		"information_seeking", "problem_solving", "content_retrieval",
		"clarification", "comparison", "definition", "procedural",
		"conversational", "exam_prep",
	}
	for _, intent := range legacyIntents {
		if !IsValidIntent(intent) {
			t.Errorf("Legacy intent %q should be valid", intent)
		}
	}
}

// TestCategoryDistribution verifies balanced test set
func TestCategoryDistribution(t *testing.T) {
	distribution := make(map[IntentType]int)
	for _, tq := range testSet100 {
		distribution[tq.ExpectedIntent]++
	}

	expectedPerCategory := 25
	for intent, count := range distribution {
		if count != expectedPerCategory {
			t.Errorf("Category %s has %d queries, expected %d", intent, count, expectedPerCategory)
		}
	}
}
