package resolver

import (
	"sort"
)

// DeviceInfo contains the minimal device information needed for resolution.
type DeviceInfo struct {
	ID   string
	Name string
}

// DeviceStore provides access to devices for resolution.
type DeviceStore interface {
	AllDeviceInfo() ([]DeviceInfo, error)
}

// Resolver resolves natural language target names to device IDs.
type Resolver struct {
	store  DeviceStore
	scorer *Scorer
}

// New creates a new Resolver.
func New(store DeviceStore, opts ...Option) *Resolver {
	r := &Resolver{
		store:  store,
		scorer: NewScorer(DefaultTokenConfig),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

// Option configures a Resolver.
type Option func(*Resolver)

// WithTokenConfig sets a custom token configuration.
func WithTokenConfig(config TokenConfig) Option {
	return func(r *Resolver) {
		r.scorer = NewScorer(config)
	}
}

// Resolve attempts to resolve a target name to a device ID.
// Returns the device ID if unambiguous, or an error with candidates if ambiguous.
func (r *Resolver) Resolve(targetName string) (string, []Candidate, error) {
	devices, err := r.store.AllDeviceInfo()
	if err != nil {
		return "", nil, err
	}

	candidates := r.scoreDevices(targetName, devices)

	if len(candidates) == 0 {
		return "", nil, &DeviceNotFoundError{TargetName: targetName}
	}

	// Check for exact match first
	for _, c := range candidates {
		if c.Score >= 0.99 {
			return c.DeviceID, nil, nil
		}
	}

	// Get top candidates that meet minimum score
	var validCandidates []Candidate
	for _, c := range candidates {
		if r.scorer.MeetsMinScore(c.Score) {
			validCandidates = append(validCandidates, c)
		}
	}

	if len(validCandidates) == 0 {
		return "", nil, &DeviceNotFoundError{TargetName: targetName}
	}

	// Check for ambiguity
	if len(validCandidates) > 1 {
		scores := make([]float64, len(validCandidates))
		for i, c := range validCandidates {
			scores[i] = c.Score
		}
		if r.scorer.IsAmbiguous(scores) {
			return "", validCandidates, &AmbiguousDeviceError{
				TargetName: targetName,
				Candidates: validCandidates,
			}
		}
	}

	// Return the best match
	return validCandidates[0].DeviceID, nil, nil
}

// scoreDevices scores all devices against the target name.
func (r *Resolver) scoreDevices(targetName string, devices []DeviceInfo) []Candidate {
	var candidates []Candidate

	for _, d := range devices {
		score := r.scorer.Score(targetName, d.Name)
		if score > 0 {
			candidates = append(candidates, Candidate{
				DeviceID: d.ID,
				Name:     d.Name,
				Score:    score,
			})
		}
	}

	// Sort by score descending
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].Score > candidates[j].Score
	})

	return candidates
}
