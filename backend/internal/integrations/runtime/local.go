package runtime

// Resolves a provider that performs no request.
//
// Every other provider builds an operation, enqueues it, performs HTTP and normalises the response.
// A local provider has nothing to fetch: its input is a value the hub already produced. Forcing it
// through the operation pipeline would mean inventing a request, so this path exists instead.
//
// Two lifecycle paths touch it, and they differ on purpose. Adding and enabling probe the source
// once, so a broken source is reported where the operator acts on it. Reading resolves the value
// fresh every time, because that value is the hub's current state — see ADR-005.

import (
	"time"

	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/definitions"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

// publishLocal records a local provider's current health without storing its value.
//
// The value itself is resolved on read (ADR-005): a local provider reads state the hub already
// holds, so a stored copy could only ever be older than the truth, and nothing would refresh it —
// the scheduler skips local instances and there is no request to refresh. Storing it produced a
// census frozen at the moment it was added, which is exactly what the operator reported seeing.
//
// What remains here is the failure: probing now means adding or enabling a broken source says so
// immediately, instead of looking like a working installation in a quiet room.
func (s *Service) publishLocal(tx *repository.IntegrationTx, i *model.Instance, c *definitions.Compiled) (model.Instance, error) {
	// Allocated once and reused, as the HTTP path does. Regenerating it per publish would present the
	// same integration as a different resource on every refresh.
	if i.ResourceID == "" {
		i.ResourceID = i.ID
	}
	now := s.clock.Now()
	// A snapshot written by an earlier build of this branch is dropped rather than migrated: it
	// describes a moment that has passed, and the read path no longer consults it.
	i.Snapshot = nil
	if _, err := s.resolveLocal(*i, c, now, nil); err != nil {
		i.LastError = failureOf(err)
	} else {
		i.LastError = nil
	}
	i.Revision++
	i.UpdatedAt = now
	if err := tx.Put("instances", i.ID, *i); err != nil {
		return *i, err
	}
	return *i, nil
}

// resolveLocal normalises the hub's current value through the capability contract.
//
// identity is nil on the read path, where a reservation could not be committed and must not be
// invented. A local capability that needed one would be a design error rather than a missing
// argument, so the read path fails loudly instead of allocating an unowned id.
func (s *Service) resolveLocal(i model.Instance, c *definitions.Compiled, now time.Time, identity func(string) (string, error)) (capabilities.Result, error) {
	if s.localSource == nil {
		return capabilities.Result{}, model.Fail("service_unavailable", "Local integrations are unavailable in this process.")
	}
	root, err := s.localSource()
	if err != nil {
		return capabilities.Result{}, err
	}
	return s.capabilities.Normalize(c.Definition, root, capabilities.Context{
		NewIdentity:     identity,
		InstanceID:      i.ID,
		ResourceID:      i.ResourceID,
		ProviderID:      c.Definition.ID,
		ProviderVersion: "sha256:" + c.Hash,
		ProfileRevision: c.Revision,
		BindingRevision: i.BindingRevision,
		Now:             now,
	})
}

// resolveLocalForRead produces the resource a local provider currently describes.
//
// The caller holds the service lock, and the source is the observation service, whose snapshot
// takes the observation mutex. That order — integrations, then observations — must stay one-way:
// the observation service asks whether it is wanted from its own worker, where no observation lock
// is held.
func (s *Service) resolveLocalForRead(i model.Instance, c *definitions.Compiled) (capabilities.Result, error) {
	return s.resolveLocal(i, c, s.clock.Now(), nil)
}

// failureOf adapts an error to the stored failure shape, keeping a stable code where one exists.
func failureOf(err error) *model.Failure {
	if failure, ok := err.(*model.Failure); ok {
		return failure
	}
	return &model.Failure{Code: "local_source_failed", Message: "The integration could not read its source."}
}

// isLocal reports whether a definition publishes without a request.
func isLocal(c *definitions.Compiled) bool {
	return c != nil && c.Definition.Local
}
