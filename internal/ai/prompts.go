package ai

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/thinkscotty/kibble/internal/feeds"
)

var numberingPattern = regexp.MustCompile(`^\s*(?:\d+[\.\)]\s*|[-*]\s+)`)

// BuildFactsPrompt constructs the prompt for generating facts.
func BuildFactsPrompt(topic, description, customInstructions, toneInstructions string, count, minWords, maxWords int) string {
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

// BuildFactsPromptWithContext constructs a fact prompt augmented with research context (RAG).
func BuildFactsPromptWithContext(topic, description, customInstructions, toneInstructions string, count, minWords, maxWords int, context string) string {
	var sb strings.Builder

	sb.WriteString("=== REFERENCE MATERIAL ===\n")
	sb.WriteString("Use the following reference material to ensure accuracy and depth. ")
	sb.WriteString("You may also draw on your general knowledge, but prefer facts grounded in this material.\n\n")
	sb.WriteString(context)
	sb.WriteString("\n\n=== END REFERENCE MATERIAL ===\n\n")

	sb.WriteString(BuildFactsPrompt(topic, description, customInstructions, toneInstructions, count, minWords, maxWords))

	return sb.String()
}

// ParseFactsFromText extracts individual facts from AI response text.
func ParseFactsFromText(text string) []string {
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

// BuildDiscoverPrompt constructs the prompt for discovering news sources.
func BuildDiscoverPrompt(topicName, description, sourcingInstructions string, suggestedFeeds []feeds.Feed, communityDomains []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`You are a helpful assistant that discovers reliable web sources for news topics.

Topic: %s
Description: %s

`, topicName, description))

	if sourcingInstructions != "" {
		sb.WriteString(sourcingInstructions)
		sb.WriteString("\n\n")
	}

	if len(suggestedFeeds) > 0 {
		sb.WriteString("Here are known-good RSS feeds that may be relevant to this topic. PREFER these feeds when they match the topic well, as they are verified to work:\n\n")
		for _, f := range suggestedFeeds {
			sb.WriteString(fmt.Sprintf("- %s (%s)\n", f.Name, f.URL))
		}
		sb.WriteString("\nYou may include additional sources beyond this list if needed to cover the topic well.\n\n")
	}

	if len(communityDomains) > 0 {
		sb.WriteString("COMMUNITY-RECOMMENDED SOURCES:\nThe following domains are frequently shared and upvoted in relevant Reddit communities. Consider including sources from these domains:\n")
		for _, domain := range communityDomains {
			sb.WriteString(fmt.Sprintf("- %s\n", domain))
		}
		sb.WriteString("\n")
	}

	sb.WriteString(`Find 6-10 reliable sources that provide ongoing news and updates related to this topic.

SOURCE DIVERSITY REQUIREMENTS:
- Include a MIX of source types: at least 1-2 niche/independent/specialized sources alongside major outlets
- Prefer sources with RSS feeds or well-structured HTML content
- Avoid paywalled or heavily JavaScript-dependent sites
- Look for: independent blogs, industry newsletters, specialized publications, official project/org feeds

SOURCE TYPES (include a variety):
- RSS feeds (preferred — most reliable for scraping)
- News websites with accessible article content
- Reddit subreddits (format as https://reddit.com/r/subredditname) — include 1-2 relevant active subreddits
- Technical blogs, official project blogs, or organizational feeds
- Niche community sites or specialized publications

For each source, provide:
1. The URL (must be a real, working URL — prefer RSS feed URLs ending in /feed, /rss, .xml when you know them)
2. A short name for the source
3. A brief description of what content it provides

IMPORTANT: Return ONLY a valid JSON array with no additional text, markdown, or explanation.

Format:
[
  {"url": "https://example.com/feed", "name": "Example News", "description": "Daily updates on topic"},
  {"url": "https://reddit.com/r/technology", "name": "r/technology", "description": "Tech news and discussion"}
]`)

	return sb.String()
}

// BuildDiscoverPromptWithContext constructs a source discovery prompt augmented with research context (RAG).
func BuildDiscoverPromptWithContext(topicName, description, sourcingInstructions string, suggestedFeeds []feeds.Feed, communityDomains []string, context string) string {
	var sb strings.Builder

	sb.WriteString("=== BACKGROUND RESEARCH ===\n")
	sb.WriteString("The following research material provides context about this topic. ")
	sb.WriteString("Use it to identify more specific, niche sources that cover this subject area.\n\n")
	sb.WriteString(context)
	sb.WriteString("\n\n=== END BACKGROUND RESEARCH ===\n\n")

	sb.WriteString(BuildDiscoverPrompt(topicName, description, sourcingInstructions, suggestedFeeds, communityDomains))

	return sb.String()
}

// BuildSummarizePrompt constructs the prompt for summarizing scraped content.
func BuildSummarizePrompt(topicName string, scrapedContent []ScrapedContent, summarizingInstructions, toneInstructions string, maxStories, minWords, maxWords int, existingTitles []string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(`You are a news summarization assistant. Analyze the following scraped content and create clear, informative news summaries.

Topic: %s

`, topicName))

	if summarizingInstructions != "" {
		sb.WriteString(summarizingInstructions)
		sb.WriteString("\n\n")
	}

	if toneInstructions != "" {
		sb.WriteString("Tone and style: ")
		sb.WriteString(toneInstructions)
		sb.WriteString("\n\n")
	}

	if minWords > 0 && maxWords > 0 {
		sb.WriteString(fmt.Sprintf("Each story summary should be between %d and %d words long.\n\n", minWords, maxWords))
	} else if minWords > 0 {
		sb.WriteString(fmt.Sprintf("Each story summary should be at least %d words long.\n\n", minWords))
	} else if maxWords > 0 {
		sb.WriteString(fmt.Sprintf("Each story summary should be at most %d words long.\n\n", maxWords))
	}

	sb.WriteString("Scraped Content:\n")
	for i, content := range scrapedContent {
		sb.WriteString(fmt.Sprintf("\n--- Source %d: %s ---\nURL: %s\n%s\n",
			i+1, content.SourceName, content.URL, content.Content))
	}

	sb.WriteString(fmt.Sprintf(`
From the content above, identify the %d most interesting and relevant news stories.

IMPORTANT FILTERING RULES:
- ONLY include content that DIRECTLY relates to the topic "%s"
- Skip any content that is off-topic or only tangentially related
- For Reddit posts, focus on substantive discussions and news, not casual comments or memes
- Prioritize recent, newsworthy content over general discussion

SOURCE DIVERSITY:
- Distribute stories across different sources. Avoid selecting more than 2 stories from the same source.
- If multiple sources report the same event, pick the best-written version from one source only.

EDITORIAL QUALITY:
- Choose stories a curious, informed reader would find surprising, illuminating, or consequential
- Prefer stories with genuine news value over routine announcements or press releases
- Skip listicles, opinion pieces with no news hook, and repackaged wire stories

`, maxStories, topicName))

	// Add dedup context if existing titles are provided
	if len(existingTitles) > 0 {
		sb.WriteString("\nDEDUPLICATION:\nThe following stories have already been published recently. Do NOT repeat these topics or events:\n")
		for _, title := range existingTitles {
			sb.WriteString(fmt.Sprintf("- %s\n", title))
		}
		sb.WriteString("Select stories covering DIFFERENT events or angles than those listed above.\n")
	}

	sb.WriteString(`
For each story:
1. Create a compelling headline (title)
2. Write a summary focusing on key facts and why this story matters
3. Include the source URL where the story was found
4. Include the source name/title

IMPORTANT: Return ONLY a valid JSON array with no additional text, markdown, or explanation.

Format:
[
  {"title": "Headline Here", "summary": "Summary text here...", "source_url": "https://source.com/article", "source_title": "Source Name"}
]`)

	return sb.String()
}

// CleanJSONResponse strips markdown code fences from JSON responses.
func CleanJSONResponse(response string) string {
	response = strings.TrimSpace(response)
	if strings.HasPrefix(response, "```json") {
		response = strings.TrimPrefix(response, "```json")
	} else if strings.HasPrefix(response, "```") {
		response = strings.TrimPrefix(response, "```")
	}
	if strings.HasSuffix(response, "```") {
		response = strings.TrimSuffix(response, "```")
	}
	return strings.TrimSpace(response)
}

// ExtractJSON attempts to extract valid JSON from a potentially messy AI response.
// It tries direct parsing first, then strips markdown fences, then finds JSON delimiters.
func ExtractJSON(raw string) string {
	raw = strings.TrimSpace(raw)

	// Try as-is first
	if looksLikeJSON(raw) {
		return raw
	}

	// Strip markdown code fences
	cleaned := CleanJSONResponse(raw)
	if looksLikeJSON(cleaned) {
		return cleaned
	}

	// Find first [ and last ] for arrays
	if start := strings.Index(raw, "["); start >= 0 {
		if end := strings.LastIndex(raw, "]"); end > start {
			candidate := raw[start : end+1]
			if looksLikeJSON(candidate) {
				return candidate
			}
		}
	}

	// Find first { and last } for objects
	if start := strings.Index(raw, "{"); start >= 0 {
		if end := strings.LastIndex(raw, "}"); end > start {
			candidate := raw[start : end+1]
			if looksLikeJSON(candidate) {
				return candidate
			}
		}
	}

	// Return cleaned version as best effort
	return cleaned
}

func looksLikeJSON(s string) bool {
	s = strings.TrimSpace(s)
	return (strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]")) ||
		(strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}"))
}
