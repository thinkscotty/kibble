package apikey

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"
	"unicode"
)

// words is a curated list of common English words, each at least 6 letters,
// used to generate human-readable API keys.
var words = []string{
	"abstract", "account", "achieve", "actions", "active", "advice",
	"afford", "almost", "amount", "animal", "answer", "appeal",
	"basket", "battle", "before", "belong", "beyond", "blanket",
	"bottle", "branch", "bridge", "bright", "broken", "budget",
	"butter", "button", "cabinet", "camera", "candle", "canyon",
	"carbon", "castle", "caught", "center", "chance", "change",
	"charge", "choice", "chosen", "circle", "classic", "clever",
	"climate", "closet", "coffee", "column", "combat", "coming",
	"common", "copper", "corner", "cosmos", "cotton", "couple",
	"course", "cousin", "create", "credit", "danger", "defeat",
	"defend", "define", "demand", "desert", "design", "detail",
	"detect", "device", "dinner", "direct", "divine", "dollar",
	"domain", "double", "dragon", "driven", "during", "eating",
	"effect", "effort", "eleven", "empire", "energy", "engine",
	"enough", "entire", "escape", "estate", "evolve", "expand",
	"expect", "export", "extend", "fabric", "factor", "farmer",
	"father", "finger", "flavor", "flight", "flower", "flying",
	"follow", "forest", "format", "fossil", "frozen", "future",
	"galaxy", "garden", "gather", "gentle", "global", "golden",
	"gospel", "govern", "ground", "growth", "guitar", "handle",
	"happen", "harbor", "heaven", "height", "hidden", "honest",
	"hunger", "hunter", "ignore", "impact", "import", "income",
	"indoor", "inform", "injure", "insect", "inside", "insist",
	"invest", "island", "jacket", "jersey", "jungle", "justice",
	"kernel", "kitten", "ladder", "laptop", "launch", "leader",
	"league", "lesson", "letter", "linear", "liquid", "listen",
	"little", "locket", "lumber", "magnet", "manner", "marble",
	"market", "master", "matter", "meadow", "mental", "method",
	"middle", "minute", "mirror", "mobile", "modern", "modest",
	"moment", "monkey", "mother", "motion", "muffin", "museum",
	"mustard", "myself", "narrow", "nation", "nature", "nearby",
	"neatly", "needle", "nephew", "nickel", "nobody",
	"normal", "notice", "number", "object", "obtain", "office",
	"online", "option", "orange", "origin", "output", "oxygen",
	"packet", "palace", "pallet", "pander", "parent", "parrot",
	"patrol", "people", "pepper", "period", "person", "pillow",
	"planet", "plenty", "pocket", "police", "polish", "poster",
	"potato", "powder", "prefer", "prince", "profit", "prompt",
	"public", "purple", "puzzle", "rabbit", "racket", "random",
	"ranger", "reason", "record", "reform", "region", "remote",
	"rental", "repair", "repeat", "rescue", "resort", "result",
	"reveal", "review", "ribbon", "rocket", "rubber", "runner",
	"rustic", "saddle", "safety", "sailor", "salmon", "sample",
	"sandal", "secret", "select", "senior", "series", "settle",
	"shadow", "shaker", "shield", "signal", "silver", "simple",
	"sister", "sketch", "smooth", "socket", "sought",
	"source", "spider", "spirit", "spread", "spring", "square",
	"stable", "statue", "steady", "stolen", "strict", "strike",
	"strong", "studio", "submit", "sudden", "summer", "sunset",
	"supper", "supply", "switch", "symbol", "system", "tablet",
	"talent", "target", "temple", "tender", "thread", "ticket",
	"timber", "tissue", "toilet", "tongue", "toward", "travel",
	"treaty", "triple", "tunnel", "turtle", "twelve", "unique",
	"update", "upload", "valley", "velvet", "vendor", "vessel",
	"victim", "violet", "virtue", "volume", "waffle", "wallet",
	"wander", "wealth", "weapon", "weekly", "weight", "window",
	"winner", "winter", "wisdom", "wonder", "worker", "worthy",
	"yellow", "zenith", "zipper",
}

// Generate creates a human-readable API key in the format:
// woRd-aNother-keyWord-12345
//
// 3-5 random words (each 6+ letters) with 2-7 randomly capitalized letters
// spread across all words, separated by dashes, ending with a random 5-digit number.
func Generate() (string, error) {
	// Pick 3-5 words
	wordCountBig, err := rand.Int(rand.Reader, big.NewInt(3))
	if err != nil {
		return "", fmt.Errorf("random word count: %w", err)
	}
	wordCount := int(wordCountBig.Int64()) + 3 // 3, 4, or 5

	chosenWords := make([][]rune, 0, wordCount)
	used := make(map[int]bool)

	for i := 0; i < wordCount; i++ {
		var idx int
		for {
			idxBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(words))))
			if err != nil {
				return "", fmt.Errorf("random word index: %w", err)
			}
			idx = int(idxBig.Int64())
			if !used[idx] {
				break
			}
		}
		used[idx] = true
		chosenWords = append(chosenWords, []rune(words[idx]))
	}

	// Collect all letter positions across all words
	type pos struct{ word, idx int }
	var positions []pos
	for w, runes := range chosenWords {
		for i := range runes {
			positions = append(positions, pos{w, i})
		}
	}

	// Pick 2-7 random positions to capitalize
	capCountBig, err := rand.Int(rand.Reader, big.NewInt(6))
	if err != nil {
		return "", fmt.Errorf("random cap count: %w", err)
	}
	capCount := int(capCountBig.Int64()) + 2 // 2-7
	if capCount > len(positions) {
		capCount = len(positions)
	}

	// Fisher-Yates shuffle first capCount positions
	for i := 0; i < capCount; i++ {
		jBig, err := rand.Int(rand.Reader, big.NewInt(int64(len(positions)-i)))
		if err != nil {
			return "", fmt.Errorf("random shuffle: %w", err)
		}
		j := int(jBig.Int64()) + i
		positions[i], positions[j] = positions[j], positions[i]
	}

	for _, p := range positions[:capCount] {
		chosenWords[p.word][p.idx] = unicode.ToUpper(chosenWords[p.word][p.idx])
	}

	// Build key parts
	parts := make([]string, 0, wordCount+1)
	for _, runes := range chosenWords {
		parts = append(parts, string(runes))
	}

	// Append a random 5-digit number (10000-99999)
	numBig, err := rand.Int(rand.Reader, big.NewInt(90000))
	if err != nil {
		return "", fmt.Errorf("random number: %w", err)
	}
	num := int(numBig.Int64()) + 10000
	parts = append(parts, fmt.Sprintf("%d", num))

	return strings.Join(parts, "-"), nil
}
