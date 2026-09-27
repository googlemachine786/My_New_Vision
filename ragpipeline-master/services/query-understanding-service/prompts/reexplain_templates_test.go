package prompts

import (
	"strings"
	"testing"
)

func TestGetAllReexplainTemplates(t *testing.T) {
	templates := GetAllReexplainTemplates()

	if len(templates) != 5 {
		t.Errorf("GetAllReexplainTemplates() returned %d templates, want 5", len(templates))
	}

	// Verify all templates have unique names
	seen := make(map[string]bool)
	for _, tmpl := range templates {
		if seen[tmpl.Name] {
			t.Errorf("Duplicate template name: %s", tmpl.Name)
		}
		seen[tmpl.Name] = true

		if tmpl.Name == "" {
			t.Error("Template has empty name")
		}
		if tmpl.Description == "" {
			t.Errorf("Template %s has empty description", tmpl.Name)
		}
		if tmpl.PromptFunc == nil {
			t.Errorf("Template %s has nil PromptFunc", tmpl.Name)
		}
	}
}

func TestTemplateNames(t *testing.T) {
	expectedNames := []string{
		"analogy_based",
		"step_by_step",
		"real_world_example",
		"simplified_language",
		"visual_description",
	}

	templates := GetAllReexplainTemplates()
	for i, expected := range expectedNames {
		if i >= len(templates) {
			t.Errorf("Missing template at index %d: %s", i, expected)
			continue
		}
		if templates[i].Name != expected {
			t.Errorf("Template[%d].Name = %q, want %q", i, templates[i].Name, expected)
		}
	}
}

func TestTemplatePromptFunc(t *testing.T) {
	templates := GetAllReexplainTemplates()

	for _, tmpl := range templates {
		t.Run(tmpl.Name, func(t *testing.T) {
			prompt := tmpl.PromptFunc("What is gravity?", "Gravity is a force that...")

			if prompt == "" {
				t.Error("PromptFunc returned empty string")
			}

			// Verify prompt contains the original query
			if !strings.Contains(prompt, "What is gravity?") {
				t.Errorf("Prompt for %s does not contain original query", tmpl.Name)
			}

			// Verify prompt contains the previous response
			if !strings.Contains(prompt, "Gravity is a force that...") {
				t.Errorf("Prompt for %s does not contain previous response", tmpl.Name)
			}

			// Verify prompt has instructions
			if !strings.Contains(prompt, "INSTRUCTIONS") && !strings.Contains(prompt, "instructions") {
				t.Errorf("Prompt for %s does not contain instructions", tmpl.Name)
			}
		})
	}
}

func TestSelectTemplateByIndex(t *testing.T) {
	tests := []struct {
		index    int
		wantName string
	}{
		{0, "analogy_based"},
		{1, "step_by_step"},
		{2, "real_world_example"},
		{3, "simplified_language"},
		{4, "visual_description"},
		{-1, "analogy_based"},  // Out of bounds defaults to first
		{5, "analogy_based"},   // Out of bounds defaults to first
		{100, "analogy_based"}, // Out of bounds defaults to first
	}

	for _, tt := range tests {
		t.Run(tt.wantName, func(t *testing.T) {
			got := SelectTemplateByIndex(tt.index)
			if got.Name != tt.wantName {
				t.Errorf("SelectTemplateByIndex(%d) = %q, want %q", tt.index, got.Name, tt.wantName)
			}
		})
	}
}

func TestAnalogyBasedTemplate(t *testing.T) {
	tmpl := AnalogyBasedTemplate()
	if tmpl.Name != "analogy_based" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "analogy_based")
	}
}

func TestStepByStepTemplate(t *testing.T) {
	tmpl := StepByStepTemplate()
	if tmpl.Name != "step_by_step" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "step_by_step")
	}
}

func TestRealWorldExampleTemplate(t *testing.T) {
	tmpl := RealWorldExampleTemplate()
	if tmpl.Name != "real_world_example" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "real_world_example")
	}
}

func TestSimplifiedLanguageTemplate(t *testing.T) {
	tmpl := SimplifiedLanguageTemplate()
	if tmpl.Name != "simplified_language" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "simplified_language")
	}
}

func TestVisualDescriptionTemplate(t *testing.T) {
	tmpl := VisualDescriptionTemplate()
	if tmpl.Name != "visual_description" {
		t.Errorf("Name = %q, want %q", tmpl.Name, "visual_description")
	}
}
