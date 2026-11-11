package settings_test

import (
	"encoding/json"
	"errors"
	"testing"

	"node-herder/mocks"
	"node-herder/models/settings"
)

// observationRepo is an in-memory settings repository that round-trips through JSON, so the
// tests exercise the same encode/decode path as the real store.
type observationRepo struct {
	mocks.NopSettingsrepo
	config *settings.AppConfig
	fail   bool
}

func (r *observationRepo) Load() (*settings.AppConfig, error) {
	encoded, _ := json.Marshal(r.config)
	var config settings.AppConfig
	if err := json.Unmarshal(encoded, &config); err != nil {
		return nil, err
	}
	return &config, nil
}

func (r *observationRepo) SaveAppConfig(config *settings.AppConfig) error {
	if r.fail {
		return errors.New("private raw storage reason")
	}
	encoded, _ := json.Marshal(config)
	return json.Unmarshal(encoded, &r.config)
}

func observationCache(t *testing.T) (*settings.AppConfigCache, *observationRepo) {
	t.Helper()
	config := settings.NewAppConfig()
	config.Bridge = settings.DefaultBridgeConfig()
	repo := &observationRepo{config: config}
	cache, err := settings.NewAppConfigCache(repo, nil)
	if err != nil {
		t.Fatal(err)
	}
	return cache, repo
}

func TestObservationIdentitiesRoundTrip(t *testing.T) {
	cache, _ := observationCache(t)

	if got := cache.ObservationIdentities(); got == nil || len(got) != 0 {
		t.Fatalf("new cache = %v, want an empty owned map", got)
	}

	if err := cache.SaveObservationIdentity("addr:aa", settings.ObservationIdentity{Decision: settings.DecisionIgnored, Label: "TV"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	got := cache.ObservationIdentities()
	if len(got) != 1 || got["addr:aa"].Decision != settings.DecisionIgnored || got["addr:aa"].Label != "TV" {
		t.Fatalf("decisions = %+v", got)
	}

	// The returned map must be owned: mutating it must not change cache state.
	got["addr:aa"] = settings.ObservationIdentity{Decision: settings.DecisionKnown}
	if cache.ObservationIdentities()["addr:aa"].Decision != settings.DecisionIgnored {
		t.Fatal("caller mutated cache state through the returned map")
	}

	if err := cache.DeleteObservationIdentity("addr:aa"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if len(cache.ObservationIdentities()) != 0 {
		t.Fatal("decision survived deletion")
	}
	// Deleting an absent decision is not an error.
	if err := cache.DeleteObservationIdentity("addr:missing"); err != nil {
		t.Fatalf("delete absent: %v", err)
	}
}

func TestSaveObservationIdentityNormalisesUnknownToDeletion(t *testing.T) {
	cache, _ := observationCache(t)
	if err := cache.SaveObservationIdentity("addr:aa", settings.ObservationIdentity{Decision: settings.DecisionKnown}); err != nil {
		t.Fatalf("save: %v", err)
	}
	// Saving the unknown default must remove the row rather than persisting a hidden value.
	if err := cache.SaveObservationIdentity("addr:aa", settings.ObservationIdentity{Decision: settings.DecisionUnknown}); err != nil {
		t.Fatalf("save unknown: %v", err)
	}
	if _, ok := cache.ObservationIdentities()["addr:aa"]; ok {
		t.Fatal("unknown decision was persisted instead of removed")
	}
	// An empty decision is treated as unknown.
	if err := cache.SaveObservationIdentity("addr:bb", settings.ObservationIdentity{Decision: ""}); err != nil {
		t.Fatalf("save empty: %v", err)
	}
	if _, ok := cache.ObservationIdentities()["addr:bb"]; ok {
		t.Fatal("empty decision was persisted instead of removed")
	}
}

func TestSaveObservationIdentityRejectsInvalidInput(t *testing.T) {
	cache, _ := observationCache(t)
	for _, test := range []struct {
		name  string
		key   string
		value settings.ObservationIdentity
	}{
		{"empty key", "", settings.ObservationIdentity{Decision: settings.DecisionKnown}},
		{"control character", "addr:\x00aa", settings.ObservationIdentity{Decision: settings.DecisionKnown}},
		{"newline", "addr:\naa", settings.ObservationIdentity{Decision: settings.DecisionKnown}},
		{"unknown decision", "addr:aa", settings.ObservationIdentity{Decision: "maybe"}},
		{"oversized key", string(make([]byte, 257)), settings.ObservationIdentity{Decision: settings.DecisionKnown}},
	} {
		t.Run(test.name, func(t *testing.T) {
			if err := cache.SaveObservationIdentity(test.key, test.value); err == nil {
				t.Fatal("invalid input was accepted")
			}
		})
	}
}

func TestObservationIdentitiesSurviveRestart(t *testing.T) {
	cache, repo := observationCache(t)
	if err := cache.SaveObservationIdentity("addr:aa", settings.ObservationIdentity{Decision: settings.DecisionIgnored}); err != nil {
		t.Fatalf("save: %v", err)
	}
	// A second cache over the same repository models a process restart.
	restarted, err := settings.NewAppConfigCache(repo, nil)
	if err != nil {
		t.Fatal(err)
	}
	if restarted.ObservationIdentities()["addr:aa"].Decision != settings.DecisionIgnored {
		t.Fatal("decision did not survive restart")
	}
}

func TestObservationIdentitiesAreStrippedFromPublicSnapshots(t *testing.T) {
	cache, _ := observationCache(t)
	if err := cache.SaveObservationIdentity("addr:aa", settings.ObservationIdentity{Decision: settings.DecisionIgnored, Label: "private"}); err != nil {
		t.Fatalf("save: %v", err)
	}
	snapshot, err := cache.LoadAppConfig()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	// Identity keys are household-adjacent, so they must never appear in a settings response.
	if snapshot.ObservationIdentities != nil {
		t.Fatalf("public snapshot leaked %d decisions", len(snapshot.ObservationIdentities))
	}
}

func TestSaveObservationIdentityReportsStorageFailure(t *testing.T) {
	cache, repo := observationCache(t)
	repo.fail = true
	if err := cache.SaveObservationIdentity("addr:aa", settings.ObservationIdentity{Decision: settings.DecisionKnown}); err == nil {
		t.Fatal("storage failure was concealed")
	}
	// A failed write must not publish the candidate state.
	if _, ok := cache.ObservationIdentities()["addr:aa"]; ok {
		t.Fatal("failed write published state")
	}
}

func TestValidObservationDecision(t *testing.T) {
	for _, decision := range []settings.ObservationDecision{settings.DecisionUnknown, settings.DecisionKnown, settings.DecisionIgnored} {
		if !settings.ValidObservationDecision(decision) {
			t.Fatalf("%q rejected", decision)
		}
	}
	for _, decision := range []settings.ObservationDecision{"", "maybe", "Ignored"} {
		if settings.ValidObservationDecision(decision) {
			t.Fatalf("%q accepted", decision)
		}
	}
}
