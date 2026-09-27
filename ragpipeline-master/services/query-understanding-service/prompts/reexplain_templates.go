// Package prompts provides prompt templates for the Query Understanding Service.
package prompts

// ReexplainTemplate represents a strategy for re-explaining a concept
type ReexplainTemplate struct {
	Name        string
	Description string
	PromptFunc  func(originalQuery, previousResponse string) string
}

// GetAllReexplainTemplates returns all 5 re-explain template strategies (Story D)
func GetAllReexplainTemplates() []ReexplainTemplate {
	return []ReexplainTemplate{
		AnalogyBasedTemplate(),
		StepByStepTemplate(),
		RealWorldExampleTemplate(),
		SimplifiedLanguageTemplate(),
		VisualDescriptionTemplate(),
	}
}

// AnalogyBasedTemplate uses analogies to explain concepts differently
func AnalogyBasedTemplate() ReexplainTemplate {
	return ReexplainTemplate{
		Name:        "analogy_based",
		Description: "Explain using a real-world analogy from a different domain",
		PromptFunc: func(originalQuery, previousResponse string) string {
			return `You are an expert educator. The student didn't fully understand the previous explanation, so please re-explain the concept using a real-world analogy.

ORIGINAL QUESTION: ` + originalQuery + `

PREVIOUS EXPLANATION: ` + previousResponse + `

INSTRUCTIONS:
1. Identify the core concept(s) from the previous explanation
2. Create a relatable analogy from everyday life (sports, cooking, travel, etc.)
3. Map each part of the analogy back to the actual concept
4. Keep the analogy simple and memorable
5. After the analogy, briefly restate the key takeaway

Format your response as:
"Let me explain this differently using an analogy..."
[Analogy explanation]
"So in summary..." [brief takeaway]`
		},
	}
}

// StepByStepTemplate breaks down concepts into sequential steps
func StepByStepTemplate() ReexplainTemplate {
	return ReexplainTemplate{
		Name:        "step_by_step",
		Description: "Break the concept into clear, numbered steps",
		PromptFunc: func(originalQuery, previousResponse string) string {
			return `You are an expert educator. The student needs a more structured explanation. Please re-explain the concept by breaking it down into clear, numbered steps.

ORIGINAL QUESTION: ` + originalQuery + `

PREVIOUS EXPLANATION: ` + previousResponse + `

INSTRUCTIONS:
1. Identify the key components of the concept
2. Break them into 3-5 logical, sequential steps
3. Explain each step clearly with a brief header
4. Show how the steps connect to form the complete picture
5. Use bullet points for clarity

Format your response as:
"Let me break this down step by step:"
Step 1: [title] - [explanation]
Step 2: [title] - [explanation]
...
"In summary, these steps work together because..."`
		},
	}
}

// RealWorldExampleTemplate uses concrete examples from different domains
func RealWorldExampleTemplate() ReexplainTemplate {
	return ReexplainTemplate{
		Name:        "real_world_example",
		Description: "Use a concrete example from a different domain than before",
		PromptFunc: func(originalQuery, previousResponse string) string {
			return `You are an expert educator. The student learns better through concrete examples. Please re-explain using a different real-world example than what was used before.

ORIGINAL QUESTION: ` + originalQuery + `

PREVIOUS EXPLANATION: ` + previousResponse + `

INSTRUCTIONS:
1. Identify the core concept
2. Choose a DIFFERENT domain for the example (if previous was math, use science; if science, use daily life)
3. Walk through the example step by step
4. Explicitly connect the example back to the abstract concept
5. Make it practical and memorable

Format your response as:
"Here's a different way to think about this with an example..."
[Example walkthrough]
"This illustrates the concept because..."`
		},
	}
}

// SimplifiedLanguageTemplate uses simpler vocabulary and shorter sentences
func SimplifiedLanguageTemplate() ReexplainTemplate {
	return ReexplainTemplate{
		Name:        "simplified_language",
		Description: "Explain using simpler language, as if teaching a younger student",
		PromptFunc: func(originalQuery, previousResponse string) string {
			return `You are an expert educator. The student needs a simpler explanation. Please re-explain as if you're talking to someone 2-3 grades younger.

ORIGINAL QUESTION: ` + originalQuery + `

PREVIOUS EXPLANATION: ` + previousResponse + `

INSTRUCTIONS:
1. Use shorter sentences (max 15-20 words each)
2. Avoid jargon - if you must use technical terms, define them immediately
3. Use "you" and "your" to make it conversational
4. Focus on the ONE most important idea (not all details)
5. Use everyday words instead of academic language

Format your response as:
"Let me explain this in a simpler way..."
[Simple explanation with short sentences]
"The main thing to remember is..."`
		},
	}
}

// VisualDescriptionTemplate creates mental images through descriptive language
func VisualDescriptionTemplate() ReexplainTemplate {
	return ReexplainTemplate{
		Name:        "visual_description",
		Description: "Create a mental picture through vivid descriptive language",
		PromptFunc: func(originalQuery, previousResponse string) string {
			return `You are an expert educator. The student learns better through visualization. Please re-explain by helping them picture the concept in their mind.

ORIGINAL QUESTION: ` + originalQuery + `

PREVIOUS EXPLANATION: ` + previousResponse + `

INSTRUCTIONS:
1. Start with "Imagine..." or "Picture this..."
2. Create a vivid mental image with sensory details
3. Walk through the visualization step by step
4. Connect what they're "seeing" to the actual concept
5. Use descriptive, colorful language

Format your response as:
"Picture this in your mind..."
[Visualization walkthrough]
"What you're seeing here is actually..." [connection to concept]`
		},
	}
}

// SelectTemplateByIndex returns a template by its index (for cycling through strategies)
func SelectTemplateByIndex(index int) ReexplainTemplate {
	templates := GetAllReexplainTemplates()
	if index < 0 || index >= len(templates) {
		// Default to first template
		return templates[0]
	}
	return templates[index]
}
