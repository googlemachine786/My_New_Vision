package enrichment

import "strings"

// PromptPrefixBuilder constructs context-aware prompt prefixes for LLM queries.
type PromptPrefixBuilder struct {
	board          string
	grade          string
	language       string
	currentChapter string
	currentSubject string
	practiceErrors []PracticeError
}

// PracticeError represents a recurring student mistake.
type PracticeError struct {
	Topic    string
	ErrorMsg string
	Count    int
}

// NewPromptPrefixBuilder creates a new builder with fluent API.
func NewPromptPrefixBuilder() *PromptPrefixBuilder {
	return &PromptPrefixBuilder{}
}

// Board sets the curriculum board.
func (b *PromptPrefixBuilder) Board(board string) *PromptPrefixBuilder {
	b.board = board
	return b
}

// Grade sets the student grade.
func (b *PromptPrefixBuilder) Grade(grade string) *PromptPrefixBuilder {
	b.grade = grade
	return b
}

// Language sets the content language.
func (b *PromptPrefixBuilder) Language(language string) *PromptPrefixBuilder {
	b.language = language
	return b
}

// CurrentChapter sets the current chapter.
func (b *PromptPrefixBuilder) CurrentChapter(chapter string) *PromptPrefixBuilder {
	b.currentChapter = chapter
	return b
}

// CurrentSubject sets the current subject.
func (b *PromptPrefixBuilder) CurrentSubject(subject string) *PromptPrefixBuilder {
	b.currentSubject = subject
	return b
}

// PracticeErrors sets the list of recent practice errors.
func (b *PromptPrefixBuilder) PracticeErrors(errors []PracticeError) *PromptPrefixBuilder {
	b.practiceErrors = errors
	return b
}

// Build constructs the final prompt prefix string.
// Returns an empty string if no context is available.
func (b *PromptPrefixBuilder) Build() string {
	var sb strings.Builder

	// Build the teaching context line
	contextParts := make([]string, 0, 4)
	if b.grade != "" {
		contextParts = append(contextParts, "Grade "+b.grade)
	}
	if b.board != "" {
		contextParts = append(contextParts, b.board+" curriculum")
	}
	if b.language != "" {
		contextParts = append(contextParts, "in "+b.language)
	}

	if len(contextParts) > 0 {
		sb.WriteString("You are teaching a " + contextParts[0] + " student studying ")
		sb.WriteString(strings.Join(contextParts[1:], " "))
		sb.WriteString(".\n")
	}

	if b.currentSubject != "" {
		sb.WriteString("Current subject: " + b.currentSubject + ".\n")
	}
	if b.currentChapter != "" {
		sb.WriteString("Current chapter: " + b.currentChapter + ".\n")
	}

	// Add practice errors awareness
	if len(b.practiceErrors) > 0 {
		sb.WriteString("\nRecent mistakes to be mindful of:\n")
		for _, pe := range b.practiceErrors {
			if pe.Count > 1 {
				sb.WriteString("- " + pe.Topic + ": " + pe.ErrorMsg + " (occurred " + itos(pe.Count) + " times)\n")
			} else {
				sb.WriteString("- " + pe.Topic + ": " + pe.ErrorMsg + "\n")
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// BuildWithQuery appends the user query to the built prefix.
func (b *PromptPrefixBuilder) BuildWithQuery(query string) string {
	prefix := b.Build()
	if prefix == "" {
		return query
	}
	return prefix + "\nAnswer the following query: " + query + "\n"
}

// itos converts an int to string without importing strconv for a single use.
func itos(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	if neg {
		digits = append([]byte{'-'}, digits...)
	}
	return string(digits)
}
