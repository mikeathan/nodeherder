package runtime

// Schedules automatic provider refreshes from persisted due times.

import (
	"hash/fnv"
	"sort"
	"time"

	model "node-herder/models/integrations"
	"node-herder/repository"
	"node-herder/utils"
)

// The coordinator wakes on a fixed cadence so a due time is admitted within one minute without
// computing timer deltas or keeping per-instance timers. The provider minimum is advisory UX,
// not a trusted provider maximum; the model package owns the shared interval bounds.
const (
	scheduleWake  = time.Minute
	maximumJitter = 5 * time.Minute
)

// destinationReviewAction marks a reconnect_required failure caused by a provider-origin change,
// not by a provider rejecting the saved credential. Only this variant is cleared by definition repair.
const destinationReviewAction = "review_destination"

// destinationReviewFailure reports a provider-origin change that needs explicit review. It must
// remain distinguishable from a provider authentication failure sharing the reconnect_required code.
func destinationReviewFailure(message string) *model.Failure {
	f := model.Fail("reconnect_required", message)
	f.RetryAction = destinationReviewAction
	return f
}

// scheduleRefresh owns the single due-work timer until service shutdown. The callback registers
// itself in the service wait group before work so Close can join it without closing the DB early.
func (s *Service) scheduleRefresh() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed || s.repo == nil || s.clock == nil || s.ctx == nil || s.ctx.Err() != nil {
		return
	}
	s.refreshTimer = s.clock.AfterFunc(scheduleWake, func() {
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			return
		}
		s.wg.Add(1)
		s.mu.Unlock()
		defer s.wg.Done()
		if err := s.runDue(); err != nil {
			utils.LogWarn("Integration schedule scan could not finish; retained due work will be retried.")
		}
		s.scheduleRefresh()
	})
}

// runDue admits or resumes enabled automatic work whose persisted due time has arrived. It also
// re-examines any outstanding continuation so a pending renewal cannot expire while the next
// regular interval is still in the future. It holds the service lock so admission cannot race
// instance edits or manual refresh; each instance is re-read in its own transaction so one bad
// definition does not roll back its neighbours.
func (s *Service) runDue() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e := s.ready(); e != nil {
		return e
	}
	now := s.clock.Now()
	due := []string{}
	e := s.repo.View(func(tx *repository.IntegrationTx) error {
		return tx.Each("instances", func(id string, data []byte) error {
			var i model.Instance
			if e := repository.DecodeIntegrationRecord(data, &i); e != nil {
				return e
			}
			if !i.Enabled || i.Schedule == nil || !i.Schedule.Enabled || i.Schedule.IntervalSeconds <= 0 {
				return nil
			}
			if i.OperationID == "" && i.Schedule.NextDueAt.After(now) {
				return nil
			}
			due = append(due, id)
			return nil
		})
	})
	if e != nil {
		return e
	}
	sort.Strings(due)
	for _, id := range due {
		if e := s.hasCapacity(); e != nil {
			return nil
		}
		if e := s.admitDue(id, now); e != nil {
			utils.LogWarnf("Integration schedule for %s was skipped: %s", id, failure(e).Code)
		}
	}
	return nil
}

// admitDue re-reads one instance and either resumes its pending refresh/renewal or admits a new
// refresh. It never starts provider I/O; accepted operation IDs are enqueued after the commit.
func (s *Service) admitDue(id string, now time.Time) error {
	admitted := ""
	e := s.repo.Update(func(tx *repository.IntegrationTx) error {
		var i model.Instance
		ok, e := tx.Get("instances", id, &i)
		if e != nil {
			return e
		}
		if !ok || !i.Enabled || i.Schedule == nil || !i.Schedule.Enabled || i.Schedule.IntervalSeconds <= 0 {
			return nil
		}
		if i.Snapshot == nil || scheduleBlocked(i) {
			return nil
		}
		if i.OperationID != "" {
			var o model.Operation
			ok, e := tx.Get("operations", i.OperationID, &o)
			if e != nil {
				return e
			}
			if ok && o.Status != "done" && o.Status != "failed" {
				if o.Kind != "refresh" && o.Kind != "renewal" {
					return nil
				}
				if o.Status == "queued" || o.Status == "running" || s.running[o.ID] != nil {
					return nil
				}
				waitUntil := o.RetryAt
				// A retryable provider failure records its backoff separately from the regular
				// cadence. Pending or interrupted attempts have no backoff and use RetryAt.
				if o.Status == "retry_required" && i.Schedule.BackoffUntil != nil && i.Schedule.BackoffUntil.After(waitUntil) {
					waitUntil = *i.Schedule.BackoffUntil
				}
				if now.Before(waitUntil) {
					return nil
				}
				if until := s.providerDeadlines[o.Origin]; now.Before(until) {
					return nil
				}
				if e = s.admitContinuation(tx, &o, now); e != nil {
					return s.persistAdmissionFailure(tx, &i, e)
				}
				admitted = o.ID
				return nil
			}
		}
		if i.Schedule.NextDueAt.After(now) {
			return nil
		}
		r, c, e := s.definition(tx, i.DefinitionKey)
		if e != nil {
			return s.persistAdmissionFailure(tx, &i, e)
		}
		if r.DiskError != "" {
			return s.persistAdmissionFailure(tx, &i, model.Fail("definition_unavailable", "Repair or reload the provider configuration first."))
		}
		if i.Origin != c.Origin {
			return s.persistAdmissionFailure(tx, &i, destinationReviewFailure("Review the changed provider destination in Configure operations."))
		}
		if until := s.providerDeadlines[c.Origin]; now.Before(until) {
			return nil
		}
		o, e := s.admitRefresh(tx, &i, c, "scheduled")
		if e != nil {
			return e
		}
		admitted = o.ID
		return nil
	})
	if e == nil && admitted != "" {
		s.enqueue(admitted)
	}
	return e
}

// applyScheduleSuccess records a terminal success and advances the next due time.
func (s *Service) applyScheduleSuccess(i *model.Instance, at time.Time) {
	if i.Schedule == nil {
		return
	}
	i.Schedule.LastAttemptAt = &at
	i.Schedule.LastSuccessAt = &at
	i.Schedule.BackoffUntil = nil
	i.Schedule.ConsecutiveFailures = 0
	if i.Schedule.Enabled {
		i.Schedule.NextDueAt = s.nextRefreshDue(at, i.Schedule.IntervalSeconds, i.ID)
	}
}

// finishInstanceSuccess applies successful schedule state and advances the instance revision/time
// once for a terminal refresh publication.
func (s *Service) finishInstanceSuccess(i *model.Instance) {
	now := s.clock.Now()
	s.applyScheduleSuccess(i, now)
	i.Revision++
	i.UpdatedAt = now
}

// applyScheduleFailure records a failure. Retryable failures use bounded exponential backoff and
// remain resumable; terminal failures wait for the selected interval, always respecting a provider
// retry deadline when one is later. Only retryable failures set BackoffUntil; the deadline is
// recorded even while automatic scheduling is off so re-enabling cannot postpone it.
func (s *Service) applyScheduleFailure(i *model.Instance, at time.Time, retryable bool, retryAt *time.Time) {
	if i.Schedule == nil {
		return
	}
	i.Schedule.LastAttemptAt = &at
	i.Schedule.LastFailureAt = &at
	i.Schedule.ConsecutiveFailures++
	interval := time.Duration(i.Schedule.IntervalSeconds) * time.Second
	if retryable {
		next := at.Add(retryDelay(i.Schedule.ConsecutiveFailures, interval))
		if retryAt != nil && retryAt.After(next) {
			next = *retryAt
		}
		i.Schedule.BackoffUntil = &next
		i.Schedule.NextDueAt = next
		return
	}
	i.Schedule.BackoffUntil = nil
	next := s.nextRefreshDue(at, i.Schedule.IntervalSeconds, i.ID)
	if retryAt != nil && retryAt.After(next) {
		next = *retryAt
	}
	i.Schedule.NextDueAt = next
}

// applySchedulePending brings the next wake forward to a pending continuation's retry time so the
// coordinator checks it before the operation's 24-hour expiry. It records the deadline even while
// automatic scheduling is off, so a disable/re-enable cannot lose it. A pending lookup is not a
// provider backoff, so any prior backoff is cleared.
func applySchedulePending(i *model.Instance, retryAt time.Time) {
	if i.Schedule == nil || retryAt.IsZero() {
		return
	}
	i.Schedule.BackoffUntil = nil
	if i.Schedule.NextDueAt.IsZero() || retryAt.Before(i.Schedule.NextDueAt) {
		i.Schedule.NextDueAt = retryAt
	}
}

// scheduleBlocked reports failures that require a household action; automatic work must not retry
// them until the binding is repaired or replaced.
func scheduleBlocked(i model.Instance) bool {
	if i.LastError == nil {
		return false
	}
	switch i.LastError.Code {
	case "reconnect_required", "restart_setup", "reconfigure_needed", "unsupported_setup",
		"unsupported_capability", "operation_expired", "definition_unavailable", "not_found",
		"invalid_definition", "reconciliation_required":
		return true
	default:
		return false
	}
}

// schedulerDefinitionFailure reports definition-level admission failures that a successful
// definition activation resolves. A reconnect_required failure is cleared only when it came from a
// destination review; a provider's rejected credential must remain visible.
func schedulerDefinitionFailure(f *model.Failure) bool {
	switch f.Code {
	case "definition_unavailable", "not_found", "invalid_definition", "unsupported_capability":
		return true
	case "reconnect_required":
		return f.RetryAction == destinationReviewAction
	default:
		return false
	}
}

// persistAdmissionFailure records a sanitized blocking admission failure on the instance so the
// settings can explain what needs repair. Transient capacity/backoff conditions are returned so the
// caller can log and retry them without mutating visible state.
func (s *Service) persistAdmissionFailure(tx *repository.IntegrationTx, i *model.Instance, cause error) error {
	f := failure(cause)
	if !schedulerBlockingFailure(f) {
		return cause
	}
	i.LastError = f
	i.Revision++
	i.UpdatedAt = s.clock.Now()
	return tx.Put("instances", i.ID, *i)
}

// schedulerBlockingFailure distinguishes a repair-needed failure from a transient condition.
func schedulerBlockingFailure(f *model.Failure) bool {
	switch f.Code {
	case "busy", "rate_limited", "retry_later", "storage_error":
		return false
	default:
		return true
	}
}

// nextRefreshDue adds deterministic jitter to the selected interval. Jitter is derived from the
// instance ID so restarts do not synchronise polls, and is bounded by the interval itself.
func (s *Service) nextRefreshDue(now time.Time, intervalSeconds int64, id string) time.Time {
	interval := time.Duration(intervalSeconds) * time.Second
	if interval <= 0 {
		return time.Time{}
	}
	return now.Add(interval + refreshJitter(id, interval))
}

// refreshJitter returns at most min(5 minutes, interval/10) for the instance.
func refreshJitter(id string, interval time.Duration) time.Duration {
	limit := interval / 10
	if limit > maximumJitter {
		limit = maximumJitter
	}
	if limit <= 0 {
		return 0
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(id))
	return time.Duration(h.Sum64() % uint64(limit))
}

// retryDelay doubles from five seconds up to one hour, never exceeding the selected interval.
func retryDelay(failures int, interval time.Duration) time.Duration {
	delay := 5 * time.Second
	for n := 1; n < failures && delay < time.Hour && n < 32; n++ {
		delay *= 2
	}
	if delay > time.Hour {
		delay = time.Hour
	}
	if interval > 0 && delay > interval {
		delay = interval
	}
	return delay
}
