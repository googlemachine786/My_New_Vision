package prompts

import (
	"fmt"
	"strings"
)

// Template versions for A/B testing
const (
	TemplateV1 = "v1"
	TemplateV2 = "v2"
)

// QueryRewritePrompt generates a prompt for rewriting a query with conversation context.
func QueryRewritePrompt(query string, history []HistoryMessage, version string) string {
	switch version {
	case TemplateV2:
		return queryRewritePromptV2(query, history)
	default:
		return queryRewritePromptV1(query, history)
	}
}

func queryRewritePromptV1(query string, history []HistoryMessage) string {
	historyText := formatHistory(history)

	return fmt.Sprintf(`You are an expert query rewriter for an educational RAG system. Your task is to rewrite the user's latest query into a standalone, self-contained question that incorporates context from the conversation history.

CONVERSATION HISTORY:
%s

LATEST QUERY:
%s

RULES:
1. Replace pronouns and ambiguous references with specific entities from history
2. Preserve the user's original intent and meaning
3. Add necessary context from previous turns without changing the query's focus
4. Keep the rewritten query concise and natural
5. If the query is already standalone and clear, return it unchanged
6. Maintain the user's language and tone
7. Do not add information that was not in the history

OUTPUT FORMAT:
Return ONLY a valid JSON object with this exact structure:
{
  "rewritten_query": "the rewritten query text",
  "confidence": 0.95,
  "changes_made": ["brief description of changes"]
}

The confidence value should reflect how confident you are in the rewrite (0.0 to 1.0).
If the query needed no changes, set confidence to 1.0.
Return ONLY the JSON object, no markdown formatting or explanations.`, historyText, query)
}

func queryRewritePromptV2(query string, history []HistoryMessage) string {
	historyText := formatHistory(history)

	return fmt.Sprintf(`You are an expert query rewriter for an educational RAG system focused on Indian curriculum content. Rewrite the user's latest query into a standalone question using conversation history.

CONVERSATION HISTORY:
%s

LATEST QUERY:
%s

RULES:
1. Resolve coreferences (pronouns, demonstratives) to specific entities
2. Preserve domain-specific terminology unchanged
3. If query references previous answer, incorporate that context
4. Detect follow-up questions and merge with original query intent
5. Keep grade/subject/chapter references intact
6. Return unchanged if already standalone

OUTPUT FORMAT:
{
  "rewritten_query": "rewritten text",
  "confidence": 0.95,
  "changes_made": ["change 1", "change 2"]
}

Return ONLY valid JSON, no additional text.`, historyText, query)
}

// SelfQueryParsePrompt generates a prompt for extracting metadata filters from a query.
func SelfQueryParsePrompt(query string, version string) string {
	switch version {
	case TemplateV2:
		return selfQueryParsePromptV2(query)
	default:
		return selfQueryParsePromptV1(query)
	}
}

func selfQueryParsePromptV1(query string) string {
	return fmt.Sprintf(`You are a metadata extraction engine for an educational content RAG system. Extract structured filters from the user's query.

QUERY:
%s

AVAILABLE FILTER FIELDS:
- grade: Grade/class numbers (e.g., "class 10", "grade 9", "12th") -> array of strings like ["10"]
- subject: Academic subjects (math, physics, chemistry, biology, english, history, geography, science, computers, economics) -> array of strings
- chapter: Chapter numbers or names (e.g., "chapter 3", "chapter on optics") -> array of strings
- topic: Specific topics or concepts (e.g., "optics", "derivatives", "photosynthesis") -> array of strings
- content_type: Type of content (theory, exercises, examples, solutions, summary, notes, formulas) -> array of strings
- board: Education board (CBSE, ICSE, state boards) -> array of strings
- language: Language of content (english, hindi, etc.) -> array of strings

EXTRACTION RULES:
1. Only extract filters that are explicitly mentioned or strongly implied
2. Normalize values: lowercase subject, numeric grade numbers, lowercase content_type
3. For chapter references, extract both number and name if available (e.g., "chapter 3 optics" -> chapter: ["3", "optics"])
4. Do not hallucate filters that are not mentioned
5. Use empty arrays for fields with no matches
6. Be conservative: prefer missing filters over incorrect ones

OUTPUT FORMAT:
Return ONLY a valid JSON object:
{
  "filters": {
    "grade": [],
    "subject": [],
    "chapter": [],
    "topic": [],
    "content_type": [],
    "board": [],
    "language": []
  },
  "confidence": 0.90,
  "explanation": "brief explanation of extracted filters"
}

Return ONLY the JSON object, no markdown or explanations.`, query)
}

func selfQueryParsePromptV2(query string) string {
	return fmt.Sprintf(`Extract metadata filters from this educational query for structured content retrieval.

QUERY:
%s

FILTER SCHEMA:
- grade: Grade numbers 1-12, extracted as strings ["10", "12"]
- subject: math, physics, chemistry, biology, english, history, geography, science, computers, economics, accountancy, business studies
- chapter: Chapter numbers or names
- topic: Specific concepts or topics
- content_type: theory, exercises, examples, solutions, summary, notes, formulas, questions, answers
- board: CBSE, ICSE, state board names
- language: english, hindi, tamil, etc.

RULES:
1. Extract only explicitly mentioned or strongly implied filters
2. Normalize: lowercase subjects and content types, numeric grades
3. Empty arrays for unspecified fields
4. Conservative extraction: fewer correct > more incorrect
5. Recognize Indian curriculum terminology

JSON OUTPUT ONLY:
{
  "filters": {
    "grade": [],
    "subject": [],
    "chapter": [],
    "topic": [],
    "content_type": [],
    "board": [],
    "language": []
  },
  "confidence": 0.0,
  "explanation": ""
}`, query)
}

// IntentClassificationPrompt generates a prompt for classifying query intent.
func IntentClassificationPrompt(query string, version string) string {
	switch version {
	case TemplateV2:
		return intentClassificationPromptV2(query)
	default:
		return intentClassificationPromptV1(query)
	}
}

func intentClassificationPromptV1(query string) string {
	return fmt.Sprintf(`You are an intent classifier for an educational RAG system. Classify the user's query into one of the predefined intent categories.

QUERY:
%s

INTENT CATEGORIES:
- information_seeking: User wants factual information or explanations (e.g., "What is photosynthesis?", "Explain Newton's laws")
- problem_solving: User needs help solving a specific problem or exercise (e.g., "Solve this equation", "Find the derivative of...")
- content_retrieval: User wants specific content by type or location (e.g., "Show me chapter 5 exercises", "Give me the summary of...")
- clarification: User is asking for clarification or elaboration (e.g., "Can you explain that more?", "What do you mean by...")
- comparison: User wants to compare concepts (e.g., "Difference between mitosis and meiosis", "Compare X and Y")
- definition: User wants a definition (e.g., "Define momentum", "What is the meaning of...")
- procedural: User wants step-by-step instructions (e.g., "How do I balance this equation?", "Steps to solve...")
- conversational: General conversation, greetings, or non-topic queries (e.g., "Hi", "Thanks", "That helped")
- exam_prep: Exam-related queries (e.g., "Previous year questions", "Important questions for board exam")

CLASSIFICATION RULES:
1. Choose the SINGLE most dominant intent
2. If multiple intents apply, pick the primary one
3. Consider the expected answer format implied by the query
4. Default to information_seeking if uncertain

OUTPUT FORMAT:
Return ONLY valid JSON:
{
  "intent": "intent_category_name",
  "confidence": 0.92,
  "secondary_intent": "other_category_or_null",
  "reasoning": "brief explanation of classification"
}

Return ONLY the JSON object.`, query)
}

func intentClassificationPromptV2(query string) string {
	return fmt.Sprintf(`Classify this educational query into the most appropriate intent category.

QUERY:
%s

CATEGORIES:
- information_seeking: Factual questions, explanations
- problem_solving: Math problems, exercises, equations
- content_retrieval: Finding specific content by chapter/grade/type
- clarification: Follow-ups, elaboration requests
- comparison: Comparing concepts, differences
- definition: Definitions, meanings
- procedural: How-to, step-by-step instructions
- conversational: Greetings, acknowledgments, chit-chat
- exam_prep: Exam questions, preparation, important topics

RULES:
1. Single best-fit category
2. Provide confidence 0.0-1.0
3. Note secondary intent if applicable
4. Brief reasoning

JSON ONLY:
{
  "intent": "category",
  "confidence": 0.0,
  "secondary_intent": null,
  "reasoning": "why this category"
}`, query)
}

// HistoryMessage represents a single message in conversation history.
type HistoryMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// formatHistory converts history messages into a readable text format.
func formatHistory(history []HistoryMessage) string {
	if len(history) == 0 {
		return "(no history)"
	}

	var sb strings.Builder
	for i, msg := range history {
		if i > 0 {
			sb.WriteString("\n")
		}
		role := msg.Role
		if role == "" {
			role = "unknown"
		}
		sb.WriteString(fmt.Sprintf("[%s]: %s", role, msg.Content))
	}
	return sb.String()
}

// ContextParams holds enriched context parameters for prompt generation.
type ContextParams struct {
	Board          string
	Grade          string
	Language       string
	CurrentChapter string
	CurrentSubject string
	PracticeErrors []PracticeError
}

// PracticeError represents a recurring student mistake.
type PracticeError struct {
	Topic    string
	ErrorMsg string
	Count    int
}

// GenerateLLMPrompt generates a fully enriched prompt with user context and retrieved sources.
// This is the main entry point for LLM prompt generation with context enrichment.
func GenerateLLMPrompt(query string, sources []string, ctx *ContextParams, version string) string {
	var sb strings.Builder

	// Write context prefix if available
	if ctx != nil {
		sb.WriteString(buildContextPrefix(ctx))
	}

	// Standard RAG instructions
	sb.WriteString(`You are a helpful educational assistant for students studying the Indian curriculum.

ANSWERING GUIDELINES:
1. Base your answer primarily on the provided context below
2. If the context doesn't contain enough information, say so clearly
3. Use age-appropriate language for the student's grade level
4. Provide clear, structured explanations
5. Include examples when helpful
6. If you're uncertain about any part, acknowledge the uncertainty

`)

	// Add retrieved context
	if len(sources) > 0 {
		sb.WriteString("RETRIEVED CONTEXT:\n")
		for i, src := range sources {
			sb.WriteString(fmt.Sprintf("[Source %d]\n%s\n\n", i+1, src))
		}
		sb.WriteString("\n")
	}

	// Add the query
	sb.WriteString("STUDENT QUERY:\n")
	sb.WriteString(query)
	sb.WriteString("\n\nYOUR ANSWER:\n")

	return sb.String()
}

func buildContextPrefix(ctx *ContextParams) string {
	var sb strings.Builder

	contextParts := make([]string, 0, 4)
	if ctx.Grade != "" {
		contextParts = append(contextParts, "Grade "+ctx.Grade)
	}
	if ctx.Board != "" {
		contextParts = append(contextParts, ctx.Board+" curriculum")
	}
	if ctx.Language != "" {
		contextParts = append(contextParts, "in "+ctx.Language)
	}

	if len(contextParts) > 0 {
		sb.WriteString(fmt.Sprintf("You are teaching a %s student studying %s.\n",
			ctx.Grade,
			strings.Join(contextParts, " ")))
	}

	if ctx.CurrentSubject != "" {
		sb.WriteString(fmt.Sprintf("Current subject: %s.\n", ctx.CurrentSubject))
	}
	if ctx.CurrentChapter != "" {
		sb.WriteString(fmt.Sprintf("Current chapter: %s.\n", ctx.CurrentChapter))
	}

	if len(ctx.PracticeErrors) > 0 {
		sb.WriteString("\nRecent mistakes to be mindful of:\n")
		for _, pe := range ctx.PracticeErrors {
			if pe.Count > 1 {
				sb.WriteString(fmt.Sprintf("- %s: %s (occurred %d times)\n", pe.Topic, pe.ErrorMsg, pe.Count))
			} else {
				sb.WriteString(fmt.Sprintf("- %s: %s\n", pe.Topic, pe.ErrorMsg))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
