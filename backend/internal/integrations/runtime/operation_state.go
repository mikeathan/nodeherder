package runtime

import (
	"errors"
	"time"

	"node-herder/internal/integrations/operations"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

func failure(e error) *model.Failure {
	var f *model.Failure
	if errors.As(e, &f) {
		return f
	}
	return model.Fail("storage_error", "The integration change could not be saved. Check persistent storage and try again.")
}
func providerFailure(r operations.Response) *model.Failure {
	if r.Status >= 500 {
		return &model.Failure{Code: "provider_unavailable", Message: "The provider is temporarily unavailable.", RetryAt: r.RetryAt}
	}
	f := model.Fail("provider_error", "The provider rejected the request. Check Setup or Configure operations.")
	switch r.Status {
	case 401, 403:
		f = model.Fail("reconnect_required", "The provider did not accept the credential. Replace it in Setup.")
	case 429:
		f = model.Fail("rate_limited", "The provider's request limit was reached. Try again after the indicated time.")
		f.RetryAt = r.RetryAt
	case 503:
		f = model.Fail("provider_unavailable", "The provider is temporarily unavailable.")
		f.RetryAt = r.RetryAt
	case 404:
		f = model.Fail("provider_not_found", "The provider did not find the configured request or saved lookup. Check the configuration.")
	}
	return f
}
func canRenew(d model.Definition, r operations.Response) bool {
	code, _ := operations.Pointer(r.Body, d.Errors.Code)
	for _, rule := range d.Errors.Rules {
		if rule.Status == r.Status && rule.Code == code && rule.Action == "renewal" {
			return true
		}
	}
	return false
}
func (s *Service) nextCheck(d model.Definition, retry *time.Time) time.Time {
	t := s.clock.Now().Add(time.Duration(d.Policy.PollMinSeconds) * time.Second)
	if retry != nil && retry.After(t) {
		t = *retry
	}
	return t
}
func clearAttempt(o *model.Operation) {
	o.Body = nil
	o.Inputs = nil
	o.State = nil
	o.IdempotencyKey = ""
	o.ETag = ""
	o.DefinitionYAML = ""
}
func (s *Service) failOperation(tx *repository.IntegrationTx, o *model.Operation, owner operationOwner, f *model.Failure) error {
	o.LastError = f
	o.Status = "failed"
	now := s.clock.Now()
	if f.Code == "provider_unavailable" || f.Code == "rate_limited" || f.Code == "invalid_response" {
		o.Status = "retry_required"
		o.RetryAt = now.Add(5 * time.Second)
		if f.RetryAt != nil && f.RetryAt.After(o.RetryAt) {
			o.RetryAt = *f.RetryAt
		}
	}
	if o.Kind == "search" || owner.candidate || owner.test {
		p := owner.setup
		p.LastError = f
		p.Status = "error"
		p.Revision++
		if e := tx.Put("setups", p.ID, p); e != nil {
			return e
		}
	} else {
		i := owner.instance
		i.LastError = f
		s.recordRun(&i, *o, o.Status, f.Message)
		i.Revision++
		i.UpdatedAt = now
		if o.Status == "failed" {
			i.OperationID = ""
		}
		if e := tx.Put("instances", i.ID, i); e != nil {
			return e
		}
	}
	if o.Status == "failed" {
		clearAttempt(o)
	}
	return tx.Put("operations", o.ID, o)
}
func (s *Service) savePending(tx *repository.IntegrationTx, o model.Operation, owner operationOwner) error {
	if owner.candidate {
		p := owner.setup
		p.State = o.State
		p.Status = "pending"
		p.Revision++
		if e := tx.Put("setups", p.ID, p); e != nil {
			return e
		}
	} else {
		i := owner.instance
		i.State = o.State
		i.LastError = nil
		i.Revision++
		i.UpdatedAt = s.clock.Now()
		if e := tx.Put("instances", i.ID, i); e != nil {
			return e
		}
	}
	return tx.Put("operations", o.ID, o)
}
