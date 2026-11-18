package runtime

// Checks the local-provider lifecycle: add, enable, disable, remove.
//
// This file owns `local.go`, `local_instances.go` and `local_providers.go`. The properties tested
// here are the ones that make a local provider honest rather than merely convenient: it cannot be
// authored as a document, it cannot be added twice, disabling clears its published value, and
// removal does not touch the device that provides its input.

import (
	"testing"
	"time"

	"node-herder/internal/integrations/capabilities/observation"
	"node-herder/internal/integrations/definitions"
	"node-herder/mocks"
	model "node-herder/models/integrations"
	observationsmodel "node-herder/models/observations"
	"node-herder/repository"
)

// localFixture registers a local provider over a stub observation snapshot.
//
// The snapshot is deliberately explicit per test rather than shared: several assertions are about
// what the published census says, and a mutable global would make them order-dependent. The clock is
// returned too, because the scheduler test has to advance it.
func localFixture(t *testing.T, snapshot observationsmodel.Snapshot) (*Service, *mocks.MockClock, model.Definition) {
	t.Helper()
	s, clock := serviceFixture(t, providerFunc(standardProvider))
	definition := model.Definition{
		Local:           true,
		SchemaVersion:   model.FormatVersion,
		IntegrationType: observation.IntegrationType,
		ID:              "hub",
		Name:            "Test local",
		Capability:      observation.ContractVersion,
	}
	s.localDefinitions = []model.Definition{definition}
	s.localCompiled = map[string]*definitions.Compiled{}
	s.localSource = func() (any, error) { return snapshot, nil }
	if e := s.installLocalDefinitions(); e != nil {
		t.Fatal(e)
	}
	return s, clock, definition
}

// TestLocalProviderRejectsAuthoredDocument guards the boundary that keeps a fetchless provider out
// of operator-authored YAML. Without it, a document could claim to be local and skip every request
// and input rule.
func TestLocalProviderRejectsAuthoredDocument(t *testing.T) {
	s, _, definition := localFixture(t, observationsmodel.Snapshot{})
	_, e := s.compiler.CompileLocal(model.Definition{Local: false, SchemaVersion: model.FormatVersion})
	if e == nil {
		t.Fatal("CompileLocal accepted a non-local definition")
	}
	compiled, e := s.compiler.CompileLocal(definition)
	if e != nil || !compiled.Definition.Local {
		t.Fatal(compiled, e)
	}
}

// TestLocalProviderAddEnableDisableRemove walks the lifecycle the operator sees.
func TestLocalProviderAddEnableDisableRemove(t *testing.T) {
	snapshot := observationsmodel.Snapshot{
		Sources: []observationsmodel.SourceDescriptor{{ID: "ble-1", Name: "Hallway"}},
		Identities: []observationsmodel.IdentityView{
			{Key: "aa:bb:cc:dd:ee:ff", Stable: true},
			{Key: "11:22:33:44:55:66", Stable: true, Classification: &observationsmodel.ClassificationView{Class: "likely-neighbour"}},
		},
	}
	s, _, definition := localFixture(t, snapshot)
	key := definitions.DefinitionKey(definition)

	if s.LocalEnabled(key) {
		t.Fatal("a local provider is enabled before it is added")
	}

	added, e := s.AddLocal(key, "")
	if e != nil {
		t.Fatal(e)
	}
	if !added.Enabled || len(added.Resources) == 0 {
		t.Fatal("adding did not produce an enabled, published instance", added)
	}
	if !s.LocalEnabled(key) {
		t.Fatal("LocalEnabled disagrees with the added instance")
	}

	// A second add would read the same radio and publish the same census twice.
	if _, e := s.AddLocal(key, ""); e == nil {
		t.Fatal("a duplicate local instance was accepted")
	}

	disabled, e := s.UpdateInstance(added.ID, added.Name, false, added.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if disabled.Enabled {
		t.Fatal("disable did not take effect")
	}
	// The snapshot must be gone: it describes a radio that is no longer being read.
	if len(disabled.Resources) != 0 {
		t.Fatal("disabling left a stale snapshot behind")
	}
	if s.LocalEnabled(key) {
		t.Fatal("LocalEnabled still reports an enabled instance after disabling")
	}

	reenabled, e := s.UpdateInstance(added.ID, added.Name, true, disabled.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if !reenabled.Enabled || len(reenabled.Resources) == 0 {
		t.Fatal("re-enabling did not republish", reenabled)
	}

	if e := s.DeleteInstance(added.ID, reenabled.ETag); e != nil {
		t.Fatal(e)
	}
	if _, e := s.Instance(added.ID); e == nil {
		t.Fatal("the instance survived removal")
	}
}

// TestLocalProviderCensusIsHonest checks that the published value reports what the author would
// actually read off the radar: identities shown, identities hidden, and which source reported.
func TestLocalProviderCensusIsHonest(t *testing.T) {
	snapshot := observationsmodel.Snapshot{
		Sources: []observationsmodel.SourceDescriptor{{
			ID:          "ble-1",
			Name:        "Hallway",
			Diagnostics: observationsmodel.SourceDiagnostics{Advertisements: 12, Decoded: 9, Dropped: 3},
		}},
		Identities: []observationsmodel.IdentityView{
			{Key: "aa:bb", Stable: true, Classification: &observationsmodel.ClassificationView{Class: "infrastructure", ShowByDefault: true}},
			{Key: "cc:dd", Stable: true, Classification: &observationsmodel.ClassificationView{Class: "likely-neighbour"}},
			{Key: "ee:ff", Stable: true},
		},
	}
	s, _, definition := localFixture(t, snapshot)
	added, e := s.AddLocal(definitions.DefinitionKey(definition), "")
	if e != nil {
		t.Fatal(e)
	}
	var census struct {
		Identities int `json:"identities"`
		Hidden     int `json:"hidden"`
		Sources    []struct {
			ID             string `json:"id"`
			Advertisements uint64 `json:"advertisements"`
		} `json:"sources"`
	}
	if e := added.Resources[0].Decode(&census); e != nil {
		t.Fatal(e)
	}
	// The third identity has no classification at all, which must not count as hidden: no evidence
	// is not the same as evidence of a demoted class.
	if census.Identities != 3 {
		t.Fatalf("identities = %d, want 3", census.Identities)
	}
	if census.Hidden != 1 {
		t.Fatalf("hidden = %d, want 1", census.Hidden)
	}
	if len(census.Sources) != 1 || census.Sources[0].Advertisements != 12 {
		t.Fatalf("sources = %+v", census.Sources)
	}
}

// TestLocalProviderCensusIsReadOnEveryRead pins ADR-005, and pins the defect it was written for.
//
// The stored-census design captured the value once and never refreshed it, because local providers
// are skipped by the scheduler and the UI hides Refresh. The operator saw "0 devices heard" beside
// "571 heard" on a page whose numbers could never change again. Resolving on read means the same
// instance reports the current state twice in a row without any refresh action in between.
func TestLocalProviderCensusIsReadOnEveryRead(t *testing.T) {
	var snapshot observationsmodel.Snapshot
	s, _, definition := localFixture(t, observationsmodel.Snapshot{})
	s.localSource = func() (any, error) { return snapshot, nil }
	added, e := s.AddLocal(definitions.DefinitionKey(definition), "")
	if e != nil {
		t.Fatal(e)
	}
	if devices(t, added) != 0 {
		t.Fatal("the fixture started with identities")
	}

	// The hub starts hearing something; nothing tells the integration about it.
	snapshot = observationsmodel.Snapshot{
		Sources:    []observationsmodel.SourceDescriptor{{ID: "ble-1", Name: "Hallway"}},
		Identities: []observationsmodel.IdentityView{{Key: "aa:bb", Stable: true}, {Key: "cc:dd", Stable: true}},
	}

	after, e := s.Instance(added.ID)
	if e != nil {
		t.Fatal(e)
	}
	if got := devices(t, after); got != 2 {
		t.Fatalf("devices = %d after the hub heard two, want 2: the census is not resolved on read", got)
	}
	// A second read must agree with the first, not replay a stored copy of it.
	again, e := s.Instance(added.ID)
	if e != nil {
		t.Fatal(e)
	}
	if got := devices(t, again); got != 2 {
		t.Fatalf("second read reported %d devices, want 2", got)
	}

	// Nothing may have been persisted: a stored census is the bug this test exists for.
	if e := s.repo.View(func(tx *repository.IntegrationTx) error {
		var stored model.Instance
		found, e := tx.Get("instances", added.ID, &stored)
		if e != nil || !found {
			return e
		}
		if stored.Snapshot != nil {
			t.Fatal("a local provider persisted a snapshot; it must be resolved on read")
		}
		return nil
	}); e != nil {
		t.Fatal(e)
	}
}

// TestLocalProviderReportsWhetherItIsContinuous lets the client stop guessing.
//
// Three client-side derivations existed and disagreed. This field is the single answer, and it must
// not be inferred from anything the browser can see: for exactly this kind of provider the
// definition is deliberately unavailable, which is what made every client rule a guess.
func TestLocalProviderReportsWhetherItIsContinuous(t *testing.T) {
	s, _, definition := localFixture(t, observationsmodel.Snapshot{})
	local, e := s.AddLocal(definitions.DefinitionKey(definition), "")
	if e != nil {
		t.Fatal(e)
	}
	if !local.Continuous {
		t.Fatal("a local provider is not reported as continuous")
	}
	// The board reads the whole list, not one instance, and it must carry the flag too.
	listed, e := s.Instances()
	if e != nil {
		t.Fatal(e)
	}
	if len(listed) != 1 || !listed[0].Continuous {
		t.Fatalf("instances = %+v, want one continuous local provider", listed)
	}
}

// TestDisabledLocalProviderPublishesNoCensus keeps "disabled" and "hearing nothing" distinct.
//
// With the value resolved on read, resolving while disabled would describe a radio nobody is
// reading — the same class of lie as a stale snapshot, with a fresher timestamp on it.
func TestDisabledLocalProviderPublishesNoCensus(t *testing.T) {
	snapshot := observationsmodel.Snapshot{
		Sources:    []observationsmodel.SourceDescriptor{{ID: "ble-1", Name: "Hallway"}},
		Identities: []observationsmodel.IdentityView{{Key: "aa:bb", Stable: true}},
	}
	s, _, definition := localFixture(t, snapshot)
	added, e := s.AddLocal(definitions.DefinitionKey(definition), "")
	if e != nil {
		t.Fatal(e)
	}
	if devices(t, added) != 1 {
		t.Fatal("the enabled instance did not publish")
	}
	disabled, e := s.UpdateInstance(added.ID, added.Name, false, added.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if len(disabled.Resources) != 0 {
		t.Fatal("a disabled local provider still published a census")
	}
}

// devices reads the census identity count out of an instance's resources.
func devices(t *testing.T, instance InstanceDTO) int {
	t.Helper()
	if len(instance.Resources) == 0 {
		return 0
	}
	var census struct {
		Identities int `json:"identities"`
	}
	if e := instance.Resources[0].Decode(&census); e != nil {
		t.Fatal(e)
	}
	return census.Identities
}

// TestLocalProviderSourceFailureIsRecorded makes sure a broken source cannot look like an empty
// room. The failure must reach the instance so the card shows an error instead of "no devices".
func TestLocalProviderSourceFailureIsRecorded(t *testing.T) {
	s, _, definition := localFixture(t, observationsmodel.Snapshot{})
	s.localSource = func() (any, error) { return nil, model.Fail("source_down", "The sensor is unreachable.") }
	added, e := s.AddLocal(definitions.DefinitionKey(definition), "")
	if e != nil {
		t.Fatal(e)
	}
	if added.Error == nil || added.Error.Code != "source_down" {
		t.Fatalf("last error = %+v", added.Error)
	}
	if len(added.Resources) != 0 {
		t.Fatal("a failed publish left a snapshot behind")
	}
}

// TestRecoveredLocalProviderStopsReportingFailure is the regression for a card stuck on "attention".
//
// `publishLocal` persists a probe failure, and the DTO seeded its status from that field. Nothing
// cleared it, so a source that failed once — while being added, or before the observation service
// was wired — reported the failure forever even after it started working. The controls that would
// clear it (Refresh, re-enable) are hidden for a continuous provider, so it was permanent.
func TestRecoveredLocalProviderStopsReportingFailure(t *testing.T) {
	s, _, definition := localFixture(t, observationsmodel.Snapshot{})
	down := true
	s.localSource = func() (any, error) {
		if down {
			return nil, model.Fail("source_down", "The sensor is unreachable.")
		}
		return observationsmodel.Snapshot{
			Sources:    []observationsmodel.SourceDescriptor{{ID: "ble-1", Name: "Hallway"}},
			Identities: []observationsmodel.IdentityView{{Key: "aa:bb", Stable: true}},
		}, nil
	}
	added, e := s.AddLocal(definitions.DefinitionKey(definition), "")
	if e != nil {
		t.Fatal(e)
	}
	if added.Error == nil {
		t.Fatal("the fixture did not record a failure")
	}

	down = false
	recovered, e := s.Instance(added.ID)
	if e != nil {
		t.Fatal(e)
	}
	if recovered.Error != nil {
		t.Fatalf("last error = %+v after the source recovered", recovered.Error)
	}
	if recovered.Status != "ready" {
		t.Fatalf("status = %q, want ready once the source reads successfully", recovered.Status)
	}
	if devices(t, recovered) != 1 {
		t.Fatal("the recovered read did not produce a census")
	}
}

// TestDisabledLocalProviderNeverProjectsAStoredCensus pins the ADR-005 invariant for a record left
// behind by an earlier build of this branch.
//
// That build stored a census on the instance. `instanceDTO` now resolves on read, but only for an
// *enabled* local provider; a disabled one used to fall through to projecting the stored payload,
// presenting a frozen census — "39 devices from 1 source" — on a card for a radio nothing is
// reading. This is the upgrade/restore case, so it is written directly to the record rather than
// through `UpdateInstance`, whose disable path clears the snapshot itself.
func TestDisabledLocalProviderNeverProjectsAStoredCensus(t *testing.T) {
	s, _, definition := localFixture(t, observationsmodel.Snapshot{})
	added, e := s.AddLocal(definitions.DefinitionKey(definition), "")
	if e != nil {
		t.Fatal(e)
	}
	stale, e := model.NewResource(map[string]any{
		"contract_version": "observation_census/v1",
		"resource_id":      added.ID,
		"revision":         1,
		"identities":       39,
		"hidden":           0,
		"sources":          []any{},
		"observed_at":      "2026-11-13T12:00:00Z",
	})
	if e != nil {
		t.Fatal(e)
	}
	// Exactly what an earlier build left on disk: disabled, with the census it published.
	if e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		var instance model.Instance
		found, e := tx.Get("instances", added.ID, &instance)
		if e != nil || !found {
			return e
		}
		instance.Enabled = false
		instance.Snapshot = &stale
		return tx.Put("instances", added.ID, instance)
	}); e != nil {
		t.Fatal(e)
	}

	instance, e := s.Instance(added.ID)
	if e != nil {
		t.Fatal(e)
	}
	if instance.Enabled {
		t.Fatal("the fixture did not disable the instance")
	}
	if len(instance.Resources) != 0 {
		t.Fatalf("a disabled local provider published %d stored resources", len(instance.Resources))
	}
}

// TestLocalProviderIsNeverScheduled pins that a local provider has no cadence.
//
// Its value is the hub's current state, so a timer would republish the same census and, worse, would
// make a disabled instance look busy. Two things must hold: enabling a schedule is refused by
// validation, and nothing admits the instance into the refresh queue even when its due time has
// passed. The first test covers validation; the second covers the scheduler's own guard.
func TestLocalProviderIsNeverScheduled(t *testing.T) {
	s, clock, definition := localFixture(t, observationsmodel.Snapshot{})
	added, e := s.AddLocal(definitions.DefinitionKey(definition), "")
	if e != nil {
		t.Fatal(e)
	}
	// A local provider declares no refresh policy, so scheduling it is rejected rather than
	// silently producing a timer that fires and does nothing.
	if _, e := s.UpdateSchedule(added.ID, true, 60, added.ETag); e == nil {
		t.Fatal("a local provider accepted a refresh schedule")
	}

	// The scheduler must also skip it directly. This is the part a stolen refresh policy — or a
	// future change to the validator — could expose, so it is driven past both guards: the instance
	// is given a due schedule by hand and the collection loop is left to decide.
	if e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		var instance model.Instance
		found, e := tx.Get("instances", added.ID, &instance)
		if e != nil || !found {
			return e
		}
		instance.Schedule = &model.RefreshSchedule{Enabled: true, IntervalSeconds: 60, NextDueAt: clock.Now().Add(-time.Hour)}
		return tx.Put("instances", added.ID, instance)
	}); e != nil {
		t.Fatal(e)
	}
	// admitDue is asserted directly rather than through runDue: runDue logs and continues on an
	// admission error, so an instance that failed to be admitted would look identical to one that
	// was skipped. A nil return with no operation is the guard working; an error is a local provider
	// being driven into operation building, where it has no request to perform.
	if e := s.admitDue(added.ID, clock.Now()); e != nil {
		t.Fatalf("admitDue refused a local provider: %v", e)
	}
	after, e := s.Instance(added.ID)
	if e != nil {
		t.Fatal(e)
	}
	if after.OperationID != "" {
		t.Fatalf("a local provider was admitted into the refresh queue: %q", after.OperationID)
	}
}

// TestLocalProviderAppearsInCatalogBeforeBeingAdded is what makes it addable at all.
//
// A provider that only appears once configured would be unreachable: the operator has no way to
// create the first instance. Availability means "can be added", not "is running".
func TestLocalProviderAppearsInCatalogBeforeBeingAdded(t *testing.T) {
	s, _, _ := localFixture(t, observationsmodel.Snapshot{})
	catalog, e := s.Catalog()
	if e != nil {
		t.Fatal(e)
	}
	listed := false
	for _, provider := range catalog.Providers {
		if provider.IntegrationType == observation.IntegrationType {
			listed = true
			if !provider.Available {
				t.Fatal("the local provider is listed as unavailable, so it cannot be added")
			}
			if provider.RequiresCredential {
				t.Fatal("a local provider must not require a credential")
			}
		}
	}
	if !listed {
		t.Fatal("the local provider is missing from the catalog")
	}
	// Nothing is configured until the operator adds it.
	for _, entry := range catalog.Types {
		if entry.ID == observation.IntegrationType && entry.ConfiguredCount != 0 {
			t.Fatalf("configured count = %d before being added", entry.ConfiguredCount)
		}
	}
}

// TestLocalRevisionAdvancesOncePerAction guards against the double-write the local branch could
// introduce: publishing commits the instance itself, so a caller that also wrote it would advance
// the revision twice per click and invalidate the ETag it just returned.
func TestLocalRevisionAdvancesOncePerAction(t *testing.T) {
	s, _, definition := localFixture(t, observationsmodel.Snapshot{})
	added, e := s.AddLocal(definitions.DefinitionKey(definition), "")
	if e != nil {
		t.Fatal(e)
	}
	renamed, e := s.UpdateInstance(added.ID, "Renamed", true, added.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if renamed.Revision != added.Revision+1 {
		t.Fatalf("rename revision = %d, want %d", renamed.Revision, added.Revision+1)
	}
	disabled, e := s.UpdateInstance(added.ID, "Renamed", false, renamed.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if disabled.Revision != renamed.Revision+1 {
		t.Fatalf("disable revision = %d, want %d", disabled.Revision, renamed.Revision+1)
	}
}
