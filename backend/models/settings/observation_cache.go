package settings

// This file persists operator triage decisions for observed wireless identities. It mirrors
// the Home layout write path: build a candidate, save it, then publish the new value while
// keeping the AppConfig pointer shared with DeviceConfigCache.

import "errors"

// errObservationKey is returned for a key that cannot be persisted safely.
var errObservationKey = errors.New("invalid observation identity key")

// maxObservationKeyBytes bounds a persisted identity key. Keys are opaque but must not be
// able to grow unbounded through a hostile or misbehaving source.
const maxObservationKeyBytes = 256

// ObservationIdentities returns an owned copy of every persisted triage decision. A missing
// map is a valid empty result, so an existing store reads back unchanged.
func (s *AppConfigCache) ObservationIdentities() map[string]ObservationIdentity {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return cloneObservationIdentities(s.appConfig.ObservationIdentities)
}

// SaveObservationIdentity upserts one decision. An empty decision normalises to unknown, and
// saving unknown removes the entry instead, so "revert" cannot leave a hidden default behind.
func (s *AppConfigCache) SaveObservationIdentity(key string, identity ObservationIdentity) error {
	if err := validateObservationKey(key); err != nil {
		return err
	}
	if identity.Decision == "" {
		identity.Decision = DecisionUnknown
	}
	if !ValidObservationDecision(identity.Decision) {
		return errObservationKey
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	identities := cloneObservationIdentities(s.appConfig.ObservationIdentities)
	if identity.Decision == DecisionUnknown {
		delete(identities, key)
	} else {
		identities[key] = identity
	}
	return s.commitObservationIdentitiesLocked(identities)
}

// DeleteObservationIdentity removes one decision, reverting the identity to unknown.
func (s *AppConfigCache) DeleteObservationIdentity(key string) error {
	if err := validateObservationKey(key); err != nil {
		return err
	}
	s.mutex.Lock()
	defer s.mutex.Unlock()

	identities := cloneObservationIdentities(s.appConfig.ObservationIdentities)
	if _, ok := identities[key]; !ok {
		// Deleting an absent decision is not an error: the caller's intent is already true.
		return nil
	}
	delete(identities, key)
	return s.commitObservationIdentitiesLocked(identities)
}

// commitObservationIdentitiesLocked persists a candidate and publishes it. The key holds the
// write lock for the whole sequence so two saves cannot interleave and lose a decision.
func (s *AppConfigCache) commitObservationIdentitiesLocked(identities map[string]ObservationIdentity) error {
	candidate := *s.appConfig
	candidate.Hub = cloneHubConfig(s.appConfig.Hub)
	candidate.ObservationIdentities = identities
	if err := s.store.SaveAppConfig(&candidate); err != nil {
		return err
	}
	s.appConfig.ObservationIdentities = identities
	return nil
}

// validateObservationKey rejects keys that cannot be persisted or would be ambiguous.
func validateObservationKey(key string) error {
	if key == "" || len(key) > maxObservationKeyBytes {
		return errObservationKey
	}
	for _, r := range key {
		if r == 0 || r < 0x20 || r == 0x7F {
			return errObservationKey
		}
	}
	return nil
}

// cloneObservationIdentities returns an owned copy so callers cannot mutate cache state.
func cloneObservationIdentities(source map[string]ObservationIdentity) map[string]ObservationIdentity {
	identities := make(map[string]ObservationIdentity, len(source))
	for key, identity := range source {
		identities[key] = identity
	}
	return identities
}
