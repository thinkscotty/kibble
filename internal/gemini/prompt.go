package gemini

import (
	"fmt"
	"regexp"
	"strings"
)

var numberingPattern = regexp.MustCompile(`^\s*(?:\d+[\.\)]\s*|[-*]\s+)`)

func BuildPrompt(topic, description, customInstructions, toneInstructions string, count, minWords, maxWords int) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(
		"Generate exactly %d unique, interesting, and accurate facts about the topic: \"%s\".\n",
		count, topic))

	if description != "" {
		sb.WriteString(fmt.Sprintf("Topic description: %s\n", description))
	}

	if customInstructions != "" {
		sb.WriteString(fmt.Sprintf("Additional instructions: %s\n", customInstructions))
	}

	if toneInstructions != "" {
		sb.WriteString(fmt.Sprintf("Tone and style: %s\n", toneInstructions))
	}

	if minWords > 0 && maxWords > 0 {
		sb.WriteString(fmt.Sprintf("Each fact should be between %d and %d words long.\n", minWords, maxWords))
	} else if minWords > 0 {
		sb.WriteString(fmt.Sprintf("Each fact should be at least %d words long.\n", minWords))
	} else if maxWords > 0 {
		sb.WriteString(fmt.Sprintf("Each fact should be at most %d words long.\n", maxWords))
	}

	sb.WriteString("\nIMPORTANT: Return ONLY the facts as a numbered list (1., 2., 3., etc.), one per line. ")
	sb.WriteString("Do not include any other text, headers, or explanations. ")
	sb.WriteString("Each fact should be a single, self-contained sentence or short paragraph.")

	return sb.String()
}

// ParseFacts extracts individual facts from the Gemini response text.
func ParseFacts(resp GenerateResponse) []string {
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return nil
	}

	text := resp.Candidates[0].Content.Parts[0].Text
	lines := strings.Split(text, "\n")

	var facts []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		cleaned := stripNumbering(line)
		if cleaned != "" {
			facts = append(facts, cleaned)
		}
	}
	return facts
}

func stripNumbering(s string) string {
	return strings.TrimSpace(numberingPattern.ReplaceAllString(s, ""))
}
