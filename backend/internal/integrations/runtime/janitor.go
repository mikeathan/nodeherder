package runtime

// Recovers interrupted operations and schedules bounded cleanup of expired runtime records.

import (
	"time"

	model "node-herder/models/integrations"
	"node-herder/repository"
	"node-herder/utils"
)

// recoverOperations marks interrupted work for explicit recovery after service startup.
func (s *Service) recoverOperations() error {
	var deadlines map[string]time.Time
	err := s.repo.Update(func(tx *repository.IntegrationTx) error {
		var err error
		deadlines, err = activeProviderLimits(tx, s.clock.Now())
		if err != nil {
			return err
		}
		updates := []model.Operation{}
		if e := tx.Each("operations", func(_ string, b []byte) error {
			var o model.Operation
			if e := repository.DecodeIntegrationRecord(b, &o); e != nil {
				return e
			}
			if o.Status == "running" || o.Status == "queued" {
				o.Status = "retry_required"
				o.RetryAt = s.clock.Now().Add(5 * time.Second)
				o.LastError = model.Fail("interrupted", "The server restarted. Check this saved operation to continue.")
				updates = append(updates, o)
			}
			return nil
		}); e != nil {
			return e
		}
		for _, o := range updates {
			if e := tx.Put("operations", o.ID, o); e != nil {
				return e
			}
		}
		return nil
	})
	// Do not expose deadlines from a transaction that failed to commit.
	if err == nil {
		s.providerDeadlines = deadlines
	}
	return err
}

// scheduleCleanup owns the periodic cleanup timer until service shutdown.
func (s *Service) scheduleCleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.janitor = s.clock.AfterFunc(time.Minute, func() {
		s.mu.Lock()
		if s.closed {
			s.mu.Unlock()
			return
		}
		// Register the callback while holding the lock so Close can wait for it safely.
		s.wg.Add(1)
		s.mu.Unlock()
		defer s.wg.Done()
		if err := s.Cleanup(); err != nil {
			utils.LogWarn("Integrations cleanup could not finish; retained records will be retried.")
		}
		s.scheduleCleanup()
	})
}

// Cleanup only removes local expired private records; it never admits provider work.
func (s *Service) Cleanup() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e := s.ready(); e != nil {
		return e
	}
	var deadlines map[string]time.Time
	err := s.repo.Update(func(tx *repository.IntegrationTx) error {
		var err error
		deadlines, err = activeProviderLimits(tx, s.clock.Now())
		if err != nil {
			return err
		}
		if err := s.expireSetups(tx); err != nil {
			return err
		}
		if err := s.expireOperations(tx); err != nil {
			return err
		}
		if err := s.pruneRunHistories(tx); err != nil {
			return err
		}
		if err := s.prunePreviousDefinitions(tx); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}
	s.providerDeadlines = deadlines
	_, err = s.repo.Rotate(32)
	return err
}
