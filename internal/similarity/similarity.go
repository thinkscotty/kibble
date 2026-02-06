package similarity

import (
	"encoding/json"
	"strings"
	"unicode"
)

// StoredTrigrams holds minimal data for similarity comparison.
type StoredTrigrams struct {
	ID       int64
	Trigrams string
}

type Checker struct {
	threshold float64
	ngramSize int
}

func New(threshold float64, ngramSize int) *Checker {
	return &Checker{threshold: threshold, ngramSize: ngramSize}
}

// normalize lowercases, removes punctuation, and collapses whitespace.
func (c *Checker) normalize(text string) string {
	var sb strings.Builder
	prevSpace := false
	for _, r := range strings.ToLower(text) {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			sb.WriteRune(r)
			prevSpace = false
		} else if !prevSpace {
			sb.WriteRune(' ')
			prevSpace = true
		}
	}
	return strings.TrimSpace(sb.String())
}

// Trigrams extracts all character n-grams from the text.
func (c *Checker) Trigrams(text string) map[string]struct{} {
	normalized := c.normalize(text)
	set := make(map[string]struct{})
	runes := []rune(normalized)
	for i := 0; i <= len(runes)-c.ngramSize; i++ {
		gram := string(runes[i : i+c.ngramSize])
		set[gram] = struct{}{}
	}
	return set
}

// TrigramsToJSON serializes a trigram set for database storage.
func (c *Checker) TrigramsToJSON(trigrams map[string]struct{}) string {
	list := make([]string, 0, len(trigrams))
	for k := range trigrams {
		list = append(list, k)
	}
	data, _ := json.Marshal(list)
	return string(data)
}

// TrigramsFromJSON deserializes a stored trigram set.
func (c *Checker) TrigramsFromJSON(data string) map[string]struct{} {
	var list []string
	json.Unmarshal([]byte(data), &list)
	set := make(map[string]struct{}, len(list))
	for _, g := range list {
		set[g] = struct{}{}
	}
	return set
}

// JaccardSimilarity computes |A intersection B| / |A union B|.
func (c *Checker) JaccardSimilarity(a, b map[string]struct{}) float64 {
	if len(a) == 0 && len(b) == 0 {
		return 1.0
	}

	intersection := 0
	for k := range a {
		if _, ok := b[k]; ok {
			intersection++
		}
	}

	union := len(a) + len(b) - intersection
	if union == 0 {
		return 0
	}
	return float64(intersection) / float64(union)
}

// IsTooSimilar checks if newFact is too similar to any existing fact.
func (c *Checker) IsTooSimilar(newFactContent string, existingFacts []StoredTrigrams) bool {
	newTrigrams := c.Trigrams(newFactContent)
	for _, existing := range existingFacts {
		existingSet := c.TrigramsFromJSON(existing.Trigrams)
		sim := c.JaccardSimilarity(newTrigrams, existingSet)
		if sim >= c.threshold {
			return true
		}
	}
	return false
}
