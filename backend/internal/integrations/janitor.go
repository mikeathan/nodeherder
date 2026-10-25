package integrations

import (
	model "node-herder/models/integrations"
	"node-herder/repository"
	"time"
)

func (s *Service) recoverOperations() error {
	return s.repo.Update(func(tx *repository.IntegrationTx) error {
		if e := tx.Each("limits", func(_ string, b []byte) error {
			var limit struct {
				Origin   string    `json:"origin"`
				Deadline time.Time `json:"deadline"`
			}
			if e := repository.DecodeIntegrationRecord(b, &limit); e != nil {
				return e
			}
			s.providerDeadlines[limit.Origin] = limit.Deadline
			return nil
		}); e != nil {
			return e
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
}
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
		s.wg.Add(1)
		s.mu.Unlock()
		defer s.wg.Done()
		_ = s.Cleanup()
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
	err := s.repo.Update(func(tx *repository.IntegrationTx) error {
		expiredSetups := []model.Setup{}
		if e := tx.Each("setups", func(_ string, b []byte) error {
			var p model.Setup
			if e := repository.DecodeIntegrationRecord(b, &p); e != nil {
				return e
			}
			if !s.clock.Now().Before(p.ExpiresAt) {
				expiredSetups = append(expiredSetups, p)
			}
			return nil
		}); e != nil {
			return e
		}
		for _, p := range expiredSetups {
			if p.Status == "expired" || p.Status == "completed" {
				if e := tx.Delete("setups", p.ID); e != nil {
					return e
				}
				continue
			}
			if cancel := s.running[p.OperationID]; cancel != nil {
				cancel()
			}
			if p.OperationID != "" {
				if e := tx.Delete("operations", p.OperationID); e != nil {
					return e
				}
			}
			p = model.Setup{ID: p.ID, DefinitionKey: p.DefinitionKey, Status: "expired", Revision: p.Revision + 1, Generation: p.Generation + 1, Choices: []model.Choice{}, ExpiresAt: s.clock.Now().Add(24 * time.Hour)}
			if e := tx.Put("setups", p.ID, p); e != nil {
				return e
			}
		}
		expiredOps := []model.Operation{}
		if e := tx.Each("operations", func(_ string, b []byte) error {
			var o model.Operation
			if e := repository.DecodeIntegrationRecord(b, &o); e != nil {
				return e
			}
			if !s.clock.Now().Before(o.ExpiresAt) {
				expiredOps = append(expiredOps, o)
			}
			return nil
		}); e != nil {
			return e
		}
		for _, o := range expiredOps {
			if cancel := s.running[o.ID]; cancel != nil {
				cancel()
			}
			if o.InstanceID != "" {
				var i model.Instance
				ok, e := tx.Get("instances", o.InstanceID, &i)
				if e != nil {
					return e
				}
				if ok && i.OperationID == o.ID {
					i.OperationID = ""
					i.LastError = model.Fail("operation_expired", "The pending lookup expired. Repeat Setup to reconnect this address.")
					i.Revision++
					if e = tx.Put("instances", i.ID, i); e != nil {
						return e
					}
				}
			}
			if e := tx.Delete("operations", o.ID); e != nil {
				return e
			}
		}
		histories := []model.Instance{}
		if e := tx.Each("instances", func(_ string, b []byte) error {
			var i model.Instance
			if e := repository.DecodeIntegrationRecord(b, &i); e != nil {
				return e
			}
			pruned := recentRuns(i.Runs, s.clock.Now())
			if len(pruned) != len(i.Runs) {
				i.Runs = pruned
				histories = append(histories, i)
			}
			return nil
		}); e != nil {
			return e
		}
		for _, i := range histories {
			if e := tx.Put("instances", i.ID, i); e != nil {
				return e
			}
		}
		previous := []model.DefinitionRecord{}
		if e := tx.Each("definitions", func(_ string, b []byte) error {
			var d model.DefinitionRecord
			if e := repository.DecodeIntegrationRecord(b, &d); e != nil {
				return e
			}
			if d.PreviousYAML != "" && !s.clock.Now().Before(d.PreviousExpiresAt) {
				d.PreviousYAML = ""
				previous = append(previous, d)
			}
			return nil
		}); e != nil {
			return e
		}
		for _, d := range previous {
			if e := tx.Put("definitions", d.Key, d); e != nil {
				return e
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	_, err = s.repo.Rotate(32)
	return err
}
