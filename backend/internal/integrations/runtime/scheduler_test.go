package runtime

import (
	"context"
	"net/http"
	"testing"
	"time"

	"node-herder/mocks"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

func TestConfiguredInstanceDefaultsToProviderSchedule(t *testing.T) {
	s, _ := serviceFixture(t, providerFunc(standardProvider))
	i := configuredFixture(t, s)
	if i.Schedule == nil || !i.Schedule.Enabled || i.Schedule.IntervalSeconds != 432000 || i.Schedule.NextDueAt == nil {
		t.Fatal(i.Schedule)
	}
}

func TestUpdateScheduleValidationAndNoProviderWork(t *testing.T) {
	calls := 0
	s, _ := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return standardProvider(r)
	}))
	i := configuredFixture(t, s)
	before := calls
	if _, e := s.UpdateSchedule(i.ID, true, 30, i.ETag); e == nil || failure(e).Code != "invalid_schedule" {
		t.Fatal("accepted interval below the definition minimum", e)
	}
	if _, e := s.UpdateSchedule(i.ID, true, 3600, `"stale"`); e == nil || failure(e).Code != "revision_conflict" {
		t.Fatal("accepted stale revision", e)
	}
	updated, e := s.UpdateSchedule(i.ID, true, 3600, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if updated.Schedule == nil || !updated.Schedule.Enabled || updated.Schedule.IntervalSeconds != 3600 {
		t.Fatal(updated.Schedule)
	}
	if updated.Schedule.NextDueAt == nil || !updated.Schedule.NextDueAt.After(s.clock.Now()) {
		t.Fatal("schedule was immediately due", updated.Schedule)
	}
	if calls != before {
		t.Fatal("saving a schedule contacted the provider")
	}
}

func TestSchedulerAdmitsDueRefresh(t *testing.T) {
	calls := 0
	s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return standardProvider(r)
	}))
	i := configuredFixture(t, s)
	before := calls
	i, e := s.UpdateSchedule(i.ID, true, 3600, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	clock.Advance(2 * time.Hour)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	if calls != before+1 {
		t.Fatalf("expected one scheduled refresh, got %d calls", calls)
	}
	after, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if after.Schedule == nil || after.Schedule.LastAttemptAt == nil || after.Schedule.LastSuccessAt == nil || after.Schedule.ConsecutiveFailures != 0 {
		t.Fatal(after.Schedule)
	}
	if after.Schedule.NextDueAt == nil || !after.Schedule.NextDueAt.After(clock.Now()) {
		t.Fatal("next due was not advanced", after.Schedule)
	}
	if len(after.Runs) == 0 || after.Runs[len(after.Runs)-1].Trigger != "scheduled" {
		t.Fatal("scheduled run was not attributed", after.Runs)
	}
}

func TestSchedulerSkipsManualDisabledAndNotDue(t *testing.T) {
	calls := 0
	s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return standardProvider(r)
	}))
	i := configuredFixture(t, s)
	before := calls
	i, e := s.UpdateSchedule(i.ID, true, 3600, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	clock.Advance(30 * time.Minute)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	if len(s.queue) != 0 {
		t.Fatal("admitted work before it was due")
	}
	i, e = s.UpdateSchedule(i.ID, false, 0, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	clock.Advance(2 * time.Hour)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	if len(s.queue) != 0 {
		t.Fatal("admitted work while automatic refresh was off")
	}
	i, e = s.UpdateSchedule(i.ID, true, 3600, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	i, e = s.UpdateInstance(i.ID, i.Name, false, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	clock.Advance(2 * time.Hour)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	if len(s.queue) != 0 {
		t.Fatal("admitted work for a disabled instance")
	}
	if calls != before {
		t.Fatal("skipped schedules contacted the provider")
	}
}

func TestSchedulerCoalescesWithManualRefresh(t *testing.T) {
	calls := 0
	s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return standardProvider(r)
	}))
	i := configuredFixture(t, s)
	before := calls
	i, e := s.UpdateSchedule(i.ID, true, 3600, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Refresh(i.ID, i.ETag); e != nil {
		t.Fatal(e)
	}
	clock.Advance(2 * time.Hour)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	if len(s.queue) != 1 {
		t.Fatalf("expected the manual operation only, got %d queued", len(s.queue))
	}
	runQueued(t, s)
	if calls != before+1 {
		t.Fatalf("expected one refresh, got %d calls", calls)
	}
}

func TestSchedulerBacksOffAndResumesRetry(t *testing.T) {
	fail := true
	s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/v1/schedules/synthetic" && fail {
			return reply(429, `{"problem":"rate_limited"}`), nil
		}
		return standardProvider(r)
	}))
	i := configuredFixture(t, s)
	i, e := s.UpdateSchedule(i.ID, true, 3600, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	clock.Advance(2 * time.Hour)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	after, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if after.Schedule == nil || after.Schedule.ConsecutiveFailures != 1 || after.Schedule.LastFailureAt == nil || after.Schedule.NextDueAt == nil {
		t.Fatal(after.Schedule)
	}
	next := *after.Schedule.NextDueAt
	if !next.After(clock.Now()) {
		t.Fatal("retry was not backed off", next)
	}
	before := len(s.queue)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	if len(s.queue) != before {
		t.Fatal("retried before the backoff deadline")
	}
	fail = false
	clock.Advance(next.Sub(clock.Now()) + time.Second)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	final, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if final.Schedule == nil || final.Schedule.ConsecutiveFailures != 0 || final.Schedule.LastSuccessAt == nil {
		t.Fatal(final.Schedule)
	}
}

func TestRefreshJitterStaysWithinBound(t *testing.T) {
	for _, interval := range []time.Duration{time.Minute, time.Hour, 5 * 24 * time.Hour} {
		limit := interval / 10
		if limit > maximumJitter {
			limit = maximumJitter
		}
		for _, id := range []string{"a", "b", "instance"} {
			if jitter := refreshJitter(id, interval); jitter < 0 || jitter >= limit {
				t.Fatalf("jitter %s for %s/%s exceeds %s", jitter, id, interval, limit)
			}
		}
	}
}

// backoffFixture drives two rate-limited refreshes so the instance carries exponential backoff.
func backoffFixture(t *testing.T) (*Service, *mocks.MockClock, InstanceDTO) {
	t.Helper()
	failures := 0
	s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/v1/schedules/synthetic" && failures < 2 {
			failures++
			response := reply(429, `{"problem":"rate_limited"}`)
			response.Header.Set("Retry-After", "1")
			return response, nil
		}
		return standardProvider(r)
	}))
	i := configuredFixture(t, s)
	i, e := s.UpdateSchedule(i.ID, true, 3600, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	clock.Advance(2 * time.Hour)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	clock.Advance(5 * time.Second)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	second, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if second.Schedule == nil || second.Schedule.ConsecutiveFailures != 2 || second.Schedule.NextDueAt == nil {
		t.Fatal(second.Schedule)
	}
	return s, clock, second
}

func TestSchedulerHonorsScheduleBackoffForExistingOperation(t *testing.T) {
	s, clock, second := backoffFixture(t)
	// A cadence edit must not replace the retry deadline with the new five-day interval.
	backoff := *second.Schedule.NextDueAt
	updated, e := s.UpdateSchedule(second.ID, true, 432000, second.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if updated.Schedule == nil || updated.Schedule.NextDueAt == nil || !updated.Schedule.NextDueAt.Equal(backoff) {
		t.Fatalf("schedule edit postponed the retry: %+v", updated.Schedule)
	}
	if updated.Schedule.IntervalSeconds != 432000 {
		t.Fatal("new cadence was not stored", updated.Schedule)
	}
	// The operation retry deadline is five seconds, but the recorded exponential backoff is longer.
	clock.Advance(6 * time.Second)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	if len(s.queue) != 0 {
		t.Fatal("resumed an existing operation before its schedule backoff")
	}
	clock.Advance(4 * time.Second)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	final, e := s.Instance(second.ID)
	if e != nil {
		t.Fatal(e)
	}
	if final.Schedule == nil || final.Schedule.ConsecutiveFailures != 0 || final.Schedule.LastSuccessAt == nil {
		t.Fatal(final.Schedule)
	}
}

func TestInstanceReenablePreservesRetryDeadline(t *testing.T) {
	s, clock, second := backoffFixture(t)
	backoff := *second.Schedule.NextDueAt
	disabled, e := s.UpdateInstance(second.ID, second.Name, false, second.ETag)
	if e != nil {
		t.Fatal(e)
	}
	enabled, e := s.UpdateInstance(second.ID, second.Name, true, disabled.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if enabled.Schedule == nil || enabled.Schedule.NextDueAt == nil || !enabled.Schedule.NextDueAt.Equal(backoff) {
		t.Fatalf("re-enable postponed the retry: %+v", enabled.Schedule)
	}
	clock.Advance(6 * time.Second)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	if len(s.queue) != 0 {
		t.Fatal("resumed an existing operation before its schedule backoff")
	}
	clock.Advance(4 * time.Second)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	final, e := s.Instance(second.ID)
	if e != nil {
		t.Fatal(e)
	}
	if final.Schedule == nil || final.Schedule.ConsecutiveFailures != 0 || final.Schedule.LastSuccessAt == nil {
		t.Fatal(final.Schedule)
	}
}

func TestScheduleReenablePreservesRetryDeadline(t *testing.T) {
	s, clock, second := backoffFixture(t)
	backoff := *second.Schedule.NextDueAt
	disabled, e := s.UpdateSchedule(second.ID, false, 3600, second.ETag)
	if e != nil {
		t.Fatal(e)
	}
	enabled, e := s.UpdateSchedule(second.ID, true, 432000, disabled.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if enabled.Schedule == nil || enabled.Schedule.NextDueAt == nil || !enabled.Schedule.NextDueAt.Equal(backoff) {
		t.Fatalf("schedule re-enable postponed the retry: %+v", enabled.Schedule)
	}
	clock.Advance(6 * time.Second)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	if len(s.queue) != 0 {
		t.Fatal("resumed an existing operation before its schedule backoff")
	}
	clock.Advance(4 * time.Second)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	final, e := s.Instance(second.ID)
	if e != nil {
		t.Fatal(e)
	}
	if final.Schedule == nil || final.Schedule.ConsecutiveFailures != 0 || final.Schedule.LastSuccessAt == nil {
		t.Fatal(final.Schedule)
	}
}

// interruptedRefreshFixture leaves a queued refresh that has no active backoff while the failure
// count is positive from an earlier terminal failure.
func interruptedRefreshFixture(t *testing.T) (*Service, *mocks.MockClock, InstanceDTO) {
	t.Helper()
	s, clock := serviceFixture(t, providerFunc(standardProvider))
	i := configuredFixture(t, s)
	i, e := s.UpdateSchedule(i.ID, true, 432000, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if e = s.repo.Update(func(tx *repository.IntegrationTx) error {
		var instance model.Instance
		if _, e := tx.Get("instances", i.ID, &instance); e != nil {
			return e
		}
		instance.Schedule.ConsecutiveFailures = 1
		instance.Schedule.BackoffUntil = nil
		instance.Schedule.NextDueAt = clock.Now().Add(-time.Minute)
		return tx.Put("instances", instance.ID, instance)
	}); e != nil {
		t.Fatal(e)
	}
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	if len(s.queue) != 1 {
		t.Fatalf("expected one admitted refresh, got %d", len(s.queue))
	}
	<-s.queue
	current, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	return s, clock, current
}

func TestInterruptedRefreshUsesRetryNotRegularCadence(t *testing.T) {
	s, clock, current := interruptedRefreshFixture(t)
	if e := s.recoverOperations(); e != nil {
		t.Fatal(e)
	}
	clock.Advance(6 * time.Second)
	if e := s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	final, e := s.Instance(current.ID)
	if e != nil {
		t.Fatal(e)
	}
	if final.OperationID != "" || final.Error != nil || final.Schedule == nil || final.Schedule.LastSuccessAt == nil {
		t.Fatal(final)
	}
}

func TestInterruptedRefreshSurvivesInstanceReenable(t *testing.T) {
	s, clock, current := interruptedRefreshFixture(t)
	disabled, e := s.UpdateInstance(current.ID, current.Name, false, current.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.UpdateInstance(current.ID, current.Name, true, disabled.ETag); e != nil {
		t.Fatal(e)
	}
	clock.Advance(6 * time.Second)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	final, e := s.Instance(current.ID)
	if e != nil {
		t.Fatal(e)
	}
	if final.OperationID != "" || final.Error != nil || final.Schedule == nil || final.Schedule.LastSuccessAt == nil {
		t.Fatal(final)
	}
}

func TestSchedulerBacksOffAfterServerError(t *testing.T) {
	failures := 0
	s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/v1/schedules/synthetic" && failures < 1 {
			failures++
			response := reply(500, `{}`)
			response.Header.Set("Retry-After", "1")
			return response, nil
		}
		return standardProvider(r)
	}))
	i := configuredFixture(t, s)
	i, e := s.UpdateSchedule(i.ID, true, 3600, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	clock.Advance(2 * time.Hour)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	failed, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if failed.Schedule == nil || failed.Schedule.ConsecutiveFailures != 1 || failed.Schedule.NextDueAt == nil {
		t.Fatal(failed.Schedule)
	}
	clock.Advance(time.Second)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	if len(s.queue) != 0 {
		t.Fatal("resumed a server error before its backoff")
	}
	clock.Advance(5 * time.Second)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	final, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if final.Schedule == nil || final.Schedule.ConsecutiveFailures != 0 || final.Schedule.LastSuccessAt == nil {
		t.Fatal(final.Schedule)
	}
}

func TestSchedulerUnchangedResponsePreservesContentAge(t *testing.T) {
	conditional := false
	s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/v1/schedules/synthetic" {
			if r.Header.Get("If-None-Match") == `"v1"` {
				conditional = true
				return reply(304, ""), nil
			}
			response := reply(200, fixtureSchedule())
			response.Header.Set("ETag", `"v1"`)
			return response, nil
		}
		return standardProvider(r)
	}))
	i := configuredFixture(t, s)
	i, e := s.UpdateSchedule(i.ID, true, 3600, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	clock.Advance(2 * time.Hour)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	first, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	before := wasteSchedule(t, first.Resources[0])
	revision := first.Resources[0].Revision
	clock.Advance(2 * time.Hour)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	second, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	after := wasteSchedule(t, second.Resources[0])
	if !conditional || second.Resources[0].Revision != revision {
		t.Fatal("304 changed resource content or was not conditional", conditional, revision, second.Resources[0].Revision)
	}
	if !after.LastCheckedAt.After(before.LastCheckedAt) {
		t.Fatal("304 did not advance the check time", before.LastCheckedAt, after.LastCheckedAt)
	}
	if !after.CapturedAt.Equal(before.CapturedAt) {
		t.Fatal("304 reset content age", before.CapturedAt, after.CapturedAt)
	}
	if second.Schedule == nil || second.Schedule.LastSuccessAt == nil {
		t.Fatal(second.Schedule)
	}
}

func TestSchedulerStopsAdmittingAtQueueCapacity(t *testing.T) {
	s, clock := serviceFixture(t, providerFunc(standardProvider))
	i := configuredFixture(t, s)
	i, e := s.UpdateSchedule(i.ID, true, 3600, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	clock.Advance(2 * time.Hour)
	for len(s.queue) < cap(s.queue) {
		s.queue <- "filler"
	}
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	if len(s.queue) != cap(s.queue) {
		t.Fatal("admission changed a full queue")
	}
	for len(s.queue) > 0 {
		<-s.queue
	}
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	if len(s.queue) != 1 {
		t.Fatalf("admission did not resume after capacity freed: %d", len(s.queue))
	}
}

func TestSchedulerPersistsBlockedDefinitionAndClearsOnRepair(t *testing.T) {
	s, clock := serviceFixture(t, providerFunc(standardProvider))
	i := configuredFixture(t, s)
	i, e := s.UpdateSchedule(i.ID, true, 3600, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if e = s.repo.Update(func(tx *repository.IntegrationTx) error {
		var record model.DefinitionRecord
		ok, e := tx.Get("definitions", i.ProviderKey, &record)
		if e != nil {
			return e
		}
		if !ok {
			return model.Fail("not_found", "missing definition")
		}
		record.DiskError = "The host file is invalid."
		return tx.Put("definitions", i.ProviderKey, record)
	}); e != nil {
		t.Fatal(e)
	}
	clock.Advance(2 * time.Hour)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	if len(s.queue) != 0 {
		t.Fatal("admitted provider work for a blocked definition")
	}
	blocked, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if blocked.Error == nil || blocked.Error.Code != "definition_unavailable" {
		t.Fatalf("blocked failure was not persisted: %+v", blocked.Error)
	}
	if blocked.Schedule == nil || blocked.Schedule.NextDueAt == nil || !blocked.Schedule.NextDueAt.Before(clock.Now()) {
		t.Fatalf("blocked due time changed unexpectedly: %+v", blocked.Schedule)
	}
	// A repeated scan is skipped without admitting work or changing the visible failure.
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	if len(s.queue) != 0 {
		t.Fatal("repeat scan admitted work")
	}
	repeated, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if repeated.Error == nil || repeated.Error.Code != "definition_unavailable" {
		t.Fatalf("blocked failure was lost: %+v", repeated.Error)
	}
	// Repairing the definition through the normal reload path clears the scheduler failure.
	definition, e := s.Definition(i.ProviderKey)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.ReloadDefinition(i.ProviderKey, definition.ETag, nil); e != nil {
		t.Fatal(e)
	}
	repaired, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if repaired.Error != nil {
		t.Fatalf("repair did not clear the failure: %+v", repaired.Error)
	}
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	final, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if final.Schedule == nil || final.Schedule.LastSuccessAt == nil {
		t.Fatal(final.Schedule)
	}
}

func TestSchedulerDefinitionFailureDistinguishesCredentialAndOrigin(t *testing.T) {
	if schedulerDefinitionFailure(&model.Failure{Code: "reconnect_required"}) {
		t.Fatal("credential failure would be cleared by a definition repair")
	}
	if !schedulerDefinitionFailure(&model.Failure{Code: "reconnect_required", RetryAction: destinationReviewAction}) {
		t.Fatal("destination-review failure would not be cleared")
	}
	if !schedulerDefinitionFailure(&model.Failure{Code: "definition_unavailable"}) {
		t.Fatal("definition failure would not be cleared")
	}
}

func TestDefinitionRepairDoesNotClearProviderAuthenticationFailure(t *testing.T) {
	s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		if r.URL.Path == "/v1/schedules/synthetic" {
			return reply(401, `{"problem":"unauthorized"}`), nil
		}
		return standardProvider(r)
	}))
	i := configuredFixture(t, s)
	i, e := s.UpdateSchedule(i.ID, true, 3600, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	clock.Advance(2 * time.Hour)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	failed, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if failed.Error == nil || failed.Error.Code != "reconnect_required" || failed.Error.RetryAction != "" {
		t.Fatalf("expected a provider credential failure: %+v", failed.Error)
	}
	// Force a same-origin definition rebind; the rejected credential must remain visible.
	if e = s.repo.Update(func(tx *repository.IntegrationTx) error {
		var record model.DefinitionRecord
		if _, e := tx.Get("definitions", i.ProviderKey, &record); e != nil {
			return e
		}
		record.DiskError = "repair"
		return tx.Put("definitions", i.ProviderKey, record)
	}); e != nil {
		t.Fatal(e)
	}
	definition, e := s.Definition(i.ProviderKey)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.ReloadDefinition(i.ProviderKey, definition.ETag, nil); e != nil {
		t.Fatal(e)
	}
	repaired, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if repaired.Error == nil || repaired.Error.Code != "reconnect_required" || repaired.Error.RetryAction != "" {
		t.Fatalf("definition repair hid a credential failure: %+v", repaired.Error)
	}
}

func TestDefinitionRepairClearsSchedulerOriginFailure(t *testing.T) {
	s, clock := serviceFixture(t, providerFunc(standardProvider))
	i := configuredFixture(t, s)
	i, e := s.UpdateSchedule(i.ID, true, 3600, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if e = s.repo.Update(func(tx *repository.IntegrationTx) error {
		var instance model.Instance
		if _, e := tx.Get("instances", i.ID, &instance); e != nil {
			return e
		}
		instance.Origin = "https://old.example"
		return tx.Put("instances", instance.ID, instance)
	}); e != nil {
		t.Fatal(e)
	}
	clock.Advance(2 * time.Hour)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	blocked, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if blocked.Error == nil || blocked.Error.Code != "reconnect_required" || blocked.Error.RetryAction != destinationReviewAction {
		t.Fatalf("expected a destination-review failure: %+v", blocked.Error)
	}
	// Restore the active origin and force a same-origin definition rebind; the scheduler failure clears.
	if e = s.repo.Update(func(tx *repository.IntegrationTx) error {
		var instance model.Instance
		if _, e := tx.Get("instances", i.ID, &instance); e != nil {
			return e
		}
		instance.Origin = "https://whenisbins.com"
		if e := tx.Put("instances", instance.ID, instance); e != nil {
			return e
		}
		var record model.DefinitionRecord
		if _, e := tx.Get("definitions", i.ProviderKey, &record); e != nil {
			return e
		}
		record.DiskError = "repair"
		return tx.Put("definitions", i.ProviderKey, record)
	}); e != nil {
		t.Fatal(e)
	}
	definition, e := s.Definition(i.ProviderKey)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.ReloadDefinition(i.ProviderKey, definition.ETag, nil); e != nil {
		t.Fatal(e)
	}
	repaired, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if repaired.Error != nil {
		t.Fatalf("origin repair did not clear the scheduler failure: %+v", repaired.Error)
	}
}

func TestPendingRenewalCannotExpireBeforeSchedulerResumesIt(t *testing.T) {
	pending := false
	s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		switch {
		case r.URL.Path == "/v1/schedules/synthetic":
			return reply(404, `{"problem":"no_schedule"}`), nil
		case pending && r.URL.Path == "/v1/lookups" && r.Method == "POST":
			return reply(201, `{"id":"lookup-2","status":"running"}`), nil
		case r.URL.Path == "/v1/lookups/lookup-2":
			return reply(200, `{"status":"done","result":`+fixtureSchedule()+`}`), nil
		default:
			return standardProvider(r)
		}
	}))
	i := configuredFixture(t, s)
	i, e := s.UpdateSchedule(i.ID, true, 432000, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	pending = true
	clock.Advance(6 * 24 * time.Hour)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	after, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if after.OperationID == "" {
		t.Fatal("pending renewal was not retained")
	}
	if after.Schedule == nil || after.Schedule.NextDueAt == nil {
		t.Fatal(after.Schedule)
	}
	if after.Schedule.NextDueAt.After(clock.Now().Add(10 * time.Second)) {
		t.Fatalf("pending renewal wake was not brought forward: %s", *after.Schedule.NextDueAt)
	}
	// Simulate a path that failed to update the schedule; the outstanding operation must still be rechecked.
	if e = s.repo.Update(func(tx *repository.IntegrationTx) error {
		var instance model.Instance
		if _, e := tx.Get("instances", i.ID, &instance); e != nil {
			return e
		}
		instance.Schedule.NextDueAt = clock.Now().Add(5 * 24 * time.Hour)
		return tx.Put("instances", instance.ID, instance)
	}); e != nil {
		t.Fatal(e)
	}
	clock.Advance(6 * time.Second)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	final, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if final.OperationID != "" || final.Error != nil || final.Schedule == nil || final.Schedule.LastSuccessAt == nil {
		t.Fatal(final)
	}
}

func TestLegacyInstanceWithoutScheduleStaysManual(t *testing.T) {
	calls := 0
	s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return standardProvider(r)
	}))
	i := configuredFixture(t, s)
	if e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		var instance model.Instance
		if _, e := tx.Get("instances", i.ID, &instance); e != nil {
			return e
		}
		instance.Schedule = nil
		return tx.Put("instances", instance.ID, instance)
	}); e != nil {
		t.Fatal(e)
	}
	legacy, e := s.Instance(i.ID)
	if e != nil {
		t.Fatal(e)
	}
	if legacy.Schedule != nil {
		t.Fatal("legacy instance was enrolled automatically", legacy.Schedule)
	}
	before := calls
	clock.Advance(24 * time.Hour)
	if e = s.runDue(); e != nil {
		t.Fatal(e)
	}
	if len(s.queue) != 0 || calls != before {
		t.Fatal("legacy instance received automatic work")
	}
	if _, e = s.Refresh(legacy.ID, legacy.ETag); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	if calls != before+1 {
		t.Fatal("manual refresh stopped working")
	}
}

func TestSchedulerTimerRunsDueWork(t *testing.T) {
	calls := 0
	s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		return standardProvider(r)
	}))
	i := configuredFixture(t, s)
	i, e := s.UpdateSchedule(i.ID, true, 3600, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	before := calls
	if e = s.repo.Update(func(tx *repository.IntegrationTx) error {
		var instance model.Instance
		if _, e := tx.Get("instances", i.ID, &instance); e != nil {
			return e
		}
		instance.Schedule.NextDueAt = clock.Now().Add(-time.Second)
		return tx.Put("instances", instance.ID, instance)
	}); e != nil {
		t.Fatal(e)
	}
	s.scheduleRefresh()
	clock.Advance(time.Minute)
	runQueued(t, s)
	if calls != before+1 {
		t.Fatalf("timer did not admit due work: %d calls", calls)
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if e = s.Close(ctx); e != nil {
		t.Fatal(e)
	}
}

func TestSchedulerStopsWithService(t *testing.T) {
	s, clock := serviceFixture(t, providerFunc(standardProvider))
	i := configuredFixture(t, s)
	if e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		var instance model.Instance
		if _, e := tx.Get("instances", i.ID, &instance); e != nil {
			return e
		}
		instance.Schedule.NextDueAt = clock.Now().Add(-time.Minute)
		return tx.Put("instances", instance.ID, instance)
	}); e != nil {
		t.Fatal(e)
	}
	s.scheduleRefresh()
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if e := s.Close(ctx); e != nil {
		t.Fatal(e)
	}
	clock.Advance(2 * time.Minute)
	if len(s.queue) != 0 {
		t.Fatal("scheduler ran after close")
	}
}
