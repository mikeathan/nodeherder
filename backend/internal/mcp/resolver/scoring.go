package resolver

import (
	"strings"
	"unicode"
)

// TokenConfig holds the configuration for token-based scoring.
type TokenConfig struct {
	CommonWords    map[string]float64
	MinScore       float64
	AmbiguityDelta float64
}

// DefaultTokenConfig provides sensible defaults for device name matching.
var DefaultTokenConfig = TokenConfig{
	CommonWords: map[string]float64{
		"sensor":      0.15,
		"device":      0.15,
		"room":        0.4,
		"temperature": 0.3,
		"humidity":    0.3,
		"motion":      0.3,
		"presence":    0.3,
		"light":       0.4,
		"switch":      0.3,
		"door":        0.4,
		"window":      0.4,
		"alarm":       0.4,
		"socket":      0.3,
		"power":       0.3,
	},
	MinScore:       0.3,
	AmbiguityDelta: 0.15,
}

// Scorer performs token-based scoring for device name matching.
type Scorer struct {
	config TokenConfig
}

// NewScorer creates a new scorer with the given configuration.
func NewScorer(config TokenConfig) *Scorer {
	return &Scorer{config: config}
}

// Score calculates a match score between a target name and a device name.
// Returns a score between 0 and 1, where 1 is a perfect match.
func (s *Scorer) Score(targetName, deviceName string) float64 {
	targetTokens := tokenize(targetName)
	deviceTokens := tokenize(deviceName)

	if len(targetTokens) == 0 || len(deviceTokens) == 0 {
		return 0
	}

	var deviceTotalWeight float64
	for _, token := range deviceTokens {
		deviceTotalWeight += s.getTokenWeight(token)
	}

	var matchedWeight float64
	// Check which device tokens are present in the target query
	// This is "Recall": how much of the device name did the user mention?
	for _, deviceToken := range deviceTokens {
		tokenWeight := s.getTokenWeight(deviceToken)

		for _, targetToken := range targetTokens {
			if strings.EqualFold(deviceToken, targetToken) {
				matchedWeight += tokenWeight
				break
			}
		}
	}

	if deviceTotalWeight == 0 {
		return 0
	}

	return matchedWeight / deviceTotalWeight
}

// getTokenWeight returns the weight for a token.
// Common words get lower weight, specific words get weight 1.0.
func (s *Scorer) getTokenWeight(token string) float64 {
	lower := strings.ToLower(token)
	if weight, ok := s.config.CommonWords[lower]; ok {
		return weight
	}
	return 1.0 // Specific/unique words get full weight
}

// tokenize splits a string into lowercase tokens.
func tokenize(s string) []string {
	var tokens []string
	var current strings.Builder

	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			current.WriteRune(unicode.ToLower(r))
		} else if current.Len() > 0 {
			tokens = append(tokens, current.String())
			current.Reset()
		}
	}

	if current.Len() > 0 {
		tokens = append(tokens, current.String())
	}

	return tokens
}

// IsAmbiguous returns true if the top two scores are within the ambiguity delta.
func (s *Scorer) IsAmbiguous(scores []float64) bool {
	if len(scores) < 2 {
		return false
	}

	// Scores should be sorted descending
	return scores[0]-scores[1] < s.config.AmbiguityDelta
}

// MeetsMinScore returns true if the score meets the minimum threshold.
func (s *Scorer) MeetsMinScore(score float64) bool {
	return score >= s.config.MinScore
}
