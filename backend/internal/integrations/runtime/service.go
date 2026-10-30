package runtime

// Owns integration service dependencies, worker lifecycle, reads and shared state.

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"sync"
	"time"

	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/definitions"
	"node-herder/internal/integrations/operations"
	"node-herder/internal/integrations/transport/httpclient"
	model "node-herder/models/integrations"
	"node-herder/repository"
	"node-herder/utils"
)

// DefinitionFiles is the fixed-directory host YAML boundary, independent of runtime storage.
type DefinitionFiles interface {
	// Read returns exact YAML and its hash; missing/unreadable files return an error.
	Read(string) (string, string, error)
	// List returns canonical type/provider keys inside the configured host directory.
	List() ([]string, error)
	// Write compares the expected hash before replacing YAML; empty hash means create only.
	Write(string, string, string) error
	// Delete removes one definition only when its expected hash still matches.
	Delete(string, string) error
}

// Options supplies private storage, host definitions, clock and the policy-bound HTTP client.
type Options struct {
	Repository   *repository.IntegrationRepository
	Definitions  DefinitionFiles
	HTTP         operations.Doer
	Clock        utils.Clock
	Capabilities *capabilities.Registry
	Shipped      fs.FS
}

// Service owns admission, four workers and a local-only janitor. All provider work is explicitly
// admitted by a mutation; GET, startup and the janitor never poll an external provider.
type Service struct {
	shipped             fs.FS
	capabilities        *capabilities.Registry
	compiler            *definitions.Compiler
	mu                  sync.Mutex
	repo                *repository.IntegrationRepository
	files               DefinitionFiles
	http                operations.Doer
	clock               utils.Clock
	ctx                 context.Context
	cancel              context.CancelFunc
	wg                  sync.WaitGroup
	queue               chan string
	running             map[string]context.CancelFunc
	closed              bool
	closeDone           chan struct{}
	publicationFailures map[string]publicationFailure
	availability        *model.Failure
	providerDeadlines   map[string]time.Time
	janitor             utils.Timer
	refreshTimer        utils.Timer
}

// NewService loads local state and starts owned workers; it performs no provider request.
func NewService(parent context.Context, o Options) (*Service, error) {
	if o.Capabilities == nil {
		return nil, errors.New("integration capabilities registry is required")
	}
	if o.Repository == nil || o.Definitions == nil {
		return nil, errors.New("integration repositories are required")
	}
	if o.Clock == nil {
		o.Clock = utils.NewRealClock()
	}
	if o.HTTP == nil {
		o.HTTP = httpclient.NewHTTPClient()
	}
	ctx, cancel := context.WithCancel(parent)
	s := &Service{
		capabilities: o.Capabilities, compiler: definitions.NewCompiler(o.Capabilities),
		repo:              o.Repository,
		files:             o.Definitions,
		shipped:           o.Shipped,
		http:              o.HTTP,
		clock:             o.Clock,
		ctx:               ctx,
		cancel:            cancel,
		queue:             make(chan string, 64),
		running:           map[string]context.CancelFunc{},
		providerDeadlines: map[string]time.Time{},
	}
	if e := s.loadDefinitions(); e != nil {
		cancel()
		return nil, e
	}
	if e := s.recoverOperations(); e != nil {
		cancel()
		return nil, e
	}
	for n := 0; n < 4; n++ {
		s.wg.Add(1)
		go s.worker()
	}
	s.scheduleCleanup()
	s.scheduleRefresh()
	return s, nil
}

// DegradedService preserves a readable availability report when private storage cannot open.
func DegradedService(code string) *Service {
	message := "Integrations are unavailable. Check the persistent key, data and configuration mounts."
	switch code {
	case "key_unavailable":
		message = "Integrations cannot open their storage key. Check the persistent key volume or restore the matching key backup."
	case "storage_unavailable":
		message = "Integrations cannot open their saved data. Check the data mount, permissions and matching storage key; another backend may already be using this database."
	case "config_directory_unwritable":
		message = "Integrations cannot access their provider configuration directory. Check that the configuration mount is writable."
	}
	return &Service{availability: model.Fail(code, message)}
}

// ready rejects actions when startup failed or the service has already closed.
func (s *Service) ready() error {
	if s.availability != nil {
		return s.availability
	}
	if s.closed || s.ctx != nil && s.ctx.Err() != nil {
		return model.Fail("service_unavailable", "Integrations are shutting down.")
	}
	return nil
}

// Close cancels and joins work before closing the database; a deadline leaves active writers intact.
func (s *Service) Close(ctx context.Context) error {
	if s.repo == nil {
		return nil
	}
	s.mu.Lock()
	if s.closeDone == nil {
		s.closeDone = make(chan struct{})
		s.closed = true
		s.cancel()
		if s.janitor != nil {
			s.janitor.Stop()
		}
		if s.refreshTimer != nil {
			s.refreshTimer.Stop()
		}
		for _, cancel := range s.running {
			cancel()
		}
		if client, ok := s.http.(interface{ CloseIdleConnections() }); ok {
			client.CloseIdleConnections()
		}
		go func() { s.wg.Wait(); close(s.closeDone) }()
	}
	done := s.closeDone
	s.mu.Unlock()
	select {
	case <-ctx.Done():
		return fmt.Errorf("integrations shutdown did not finish: %w", ctx.Err())
	case <-done:
		return s.repo.Close()
	}
}

// Catalog lists supported types and local provider definitions without network I/O.
func (s *Service) Catalog() (CatalogDTO, error) {
	result := CatalogDTO{
		Available: s.availability == nil,
		Error:     s.availability,
		Types:     []TypeDTO{},
		Providers: []ProviderDTO{},
	}
	seenTypes := map[string]bool{}
	for _, descriptor := range s.capabilities.Descriptors() {
		if seenTypes[descriptor.IntegrationType] {
			continue
		}
		seenTypes[descriptor.IntegrationType] = true
		result.Types = append(result.Types, TypeDTO{ID: descriptor.IntegrationType, Name: descriptor.Name, Description: descriptor.Description})
	}
	if s.repo == nil {
		return result, nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	err := s.repo.View(func(tx *repository.IntegrationTx) error {
		return tx.Each("definitions", func(_ string, b []byte) error {
			var r model.DefinitionRecord
			if e := repository.DecodeIntegrationRecord(b, &r); e != nil {
				return e
			}
			c, e := s.compiler.CompileYAML(r.YAML)
			if e != nil {
				result.Providers = append(result.Providers, ProviderDTO{
					Key:             r.Key,
					ID:              strings.Split(r.Key, "/")[1],
					IntegrationType: strings.Split(r.Key, "/")[0],
					Name:            r.Key,
					Available:       false,
					DiskError:       r.DiskError,
				})
				return nil
			}
			d := c.Definition
			result.Providers = append(result.Providers, ProviderDTO{
				Key:             r.Key,
				ID:              d.ID,
				Name:            d.Name,
				IntegrationType: d.IntegrationType,
				Origin:          c.Origin,
				Available:       r.DiskError == "",
				DiskError:       r.DiskError,
			})
			return nil
		})
	})
	return result, err
}

// definition reads and compiles the active stored provider definition for the transaction.
func (s *Service) definition(tx *repository.IntegrationTx, key string) (model.DefinitionRecord, *definitions.Compiled, error) {
	var r model.DefinitionRecord
	ok, e := tx.Get("definitions", key, &r)
	if e != nil {
		return r, nil, e
	}
	if !ok {
		return r, nil, model.Fail("not_found", "Provider not found.")
	}
	c, e := s.compiler.CompileYAML(r.YAML)
	if c != nil {
		c.Revision = r.Revision
	}
	return r, c, e
}

// revisionETag identifies a particular setup or instance revision for conditional updates.
func revisionETag(kind, id string, revision uint64) string {
	return fmt.Sprintf(`"%s:%s:%d"`, kind, id, revision)
}

// checkRevision rejects a missing or stale revision before applying a change.
func checkRevision(expected, actual string) error {
	if expected == "" {
		return model.Fail("precondition_required", "Reload this page before saving changes.")
	}
	if expected != actual {
		return model.Fail("revision_conflict", "This item changed in another session. Reload it before saving.")
	}
	return nil
}

// instanceDTO builds a sanitized instance view without exposing private credential values.
func (s *Service) instanceDTO(i model.Instance) InstanceDTO {
	status := "ready"
	if i.Snapshot == nil {
		status = "pending"
	}
	if i.LastError != nil {
		status = i.LastError.Code
	}
	if !i.Enabled {
		status = "disabled"
	}
	if diagnostic, ok := s.publicationFailures[i.OperationID]; ok && diagnostic.Generation == i.Generation {
		i.LastError = diagnostic.Error
		status = "save_failed"
	}
	resources := []model.Resource{}
	sources := []model.SourceReference{}
	if i.Snapshot != nil {
		projected, err := s.projectResource(*i.Snapshot, s.clock.Now())
		if err == nil {
			var references []model.SourceReference
			references, err = s.capabilities.Sources(projected)
			if err == nil {
				resources = append(resources, projected)
				sources = append(sources, references...)
			}
		}
		if err != nil {
			i.LastError = failure(err)
			status = i.LastError.Code
		}
	}
	slots := []string{}
	for slot := range i.Credentials {
		slots = append(slots, slot)
	}
	sort.Strings(slots)
	var schedule *ScheduleDTO
	if i.Schedule != nil {
		schedule = &ScheduleDTO{
			Enabled:             i.Schedule.Enabled,
			IntervalSeconds:     i.Schedule.IntervalSeconds,
			LastAttemptAt:       i.Schedule.LastAttemptAt,
			LastSuccessAt:       i.Schedule.LastSuccessAt,
			LastFailureAt:       i.Schedule.LastFailureAt,
			ConsecutiveFailures: i.Schedule.ConsecutiveFailures,
		}
		if i.Schedule.Enabled && !i.Schedule.NextDueAt.IsZero() {
			due := i.Schedule.NextDueAt
			schedule.NextDueAt = &due
		}
	}
	return InstanceDTO{
		Sources:              sources,
		CredentialSlots:      slots,
		Runs:                 recentRuns(i.Runs, s.clock.Now()),
		Schedule:             schedule,
		ID:                   i.ID,
		Name:                 i.Name,
		ProviderKey:          i.DefinitionKey,
		Revision:             i.Revision,
		ETag:                 revisionETag("instance", i.ID, i.Revision),
		Enabled:              i.Enabled,
		Status:               status,
		Postcode:             i.Inputs[i.SearchInput],
		OperationID:          i.OperationID,
		CredentialConfigured: len(i.Credentials) > 0,
		Resources:            resources,
		Error:                i.LastError,
		UpdatedAt:            i.UpdatedAt,
	}
}

// setupDTO builds a sanitized setup view and projects any saved preview.
func (s *Service) setupDTO(p model.Setup) SetupDTO {
	postcode := ""
	configured := false
	if c, e := s.compiler.CompileYAML(p.DefinitionYAML); e == nil {
		configured = credential(c.Definition, p.Credentials) != nil
		postcode = p.Inputs[c.Definition.Setup.Search.Input]
	}
	if p.Preview != nil {
		projected, err := s.projectResource(*p.Preview, s.clock.Now())
		if err != nil {
			p.Preview = nil
			p.LastError = failure(err)
			p.Status = "error"
		} else {
			p.Preview = &projected
		}
	}

	return SetupDTO{
		ID:                   p.ID,
		Revision:             p.Revision,
		ETag:                 revisionETag("setup", p.ID, p.Revision),
		ProviderKey:          p.DefinitionKey,
		Status:               p.Status,
		Choices:              p.Choices,
		Postcode:             postcode,
		InstanceID:           p.InstanceID,
		OperationID:          p.OperationID,
		ExpiresAt:            p.ExpiresAt,
		CredentialConfigured: configured,
		Error:                p.LastError,
		Preview:              p.Preview,
	}
}

// operationDTO removes private attempt details from the returned operation view.
func operationDTO(o model.Operation) OperationDTO {
	return OperationDTO{
		ID:         o.ID,
		Status:     o.Status,
		Kind:       o.Kind,
		SetupID:    o.SetupID,
		InstanceID: o.InstanceID,
		RetryAt:    o.RetryAt,
		ExpiresAt:  o.ExpiresAt,
		Error:      o.LastError,
	}
}

// Instances returns shared, redacted connection summaries and current read-time projections.
func (s *Service) Instances() ([]InstanceDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if e := s.ready(); e != nil {
		return nil, e
	}
	out := []InstanceDTO{}
	e := s.repo.View(func(tx *repository.IntegrationTx) error {
		return tx.Each("instances", func(_ string, b []byte) error {
			var i model.Instance
			if e := repository.DecodeIntegrationRecord(b, &i); e != nil {
				return e
			}
			out = append(out, s.instanceDTO(i))
			return nil
		})
	})
	return out, e
}

// Instance reads one connection without refreshing its provider.
func (s *Service) Instance(id string) (InstanceDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out InstanceDTO
	if e := s.ready(); e != nil {
		return out, e
	}
	e := s.repo.View(func(tx *repository.IntegrationTx) error {
		var i model.Instance
		ok, e := tx.Get("instances", id, &i)
		if e != nil {
			return e
		}
		if !ok {
			return model.Fail("not_found", "Integration not found.")
		}
		out = s.instanceDTO(i)
		return nil
	})
	return out, e
}

// Operation reads durable progress and any uncommitted-publication diagnostic.
func (s *Service) Operation(id string) (OperationDTO, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out OperationDTO
	if e := s.ready(); e != nil {
		return out, e
	}
	e := s.repo.View(func(tx *repository.IntegrationTx) error {
		var o model.Operation
		ok, e := tx.Get("operations", id, &o)
		if e != nil {
			return e
		}
		if !ok {
			return model.Fail("not_found", "Operation not found.")
		}
		out = operationDTO(o)
		if diagnostic, ok := s.publicationFailures[o.ID]; ok && diagnostic.Generation == o.Generation {
			out.Error = diagnostic.Error
			out.Status = "save_failed"
		}
		return nil
	})
	return out, e
}
func (s *Service) enqueue(id string) {
	delete(s.publicationFailures, id)
	if s.ctx == nil {
		return
	}
	// A shutdown may cancel the context while accepted work is still durable. Do not wedge the
	// caller on a full queue; restart recovery marks the unclaimed operation retryable.
	select {
	case s.queue <- id:
	case <-s.ctx.Done():
	}
}

// hasCapacity checks worker queue capacity before admitting durable work.
func (s *Service) hasCapacity() error {
	if len(s.queue) >= cap(s.queue) {
		return model.Fail("busy", "Integrations are busy. Try again shortly.")
	}
	return nil
}

// recentRuns bounds retained run history by age and count.
func recentRuns(runs []model.RunSummary, now time.Time) []model.RunSummary {
	out := []model.RunSummary{}
	for _, r := range runs {
		if r.At.After(now.Add(-7 * 24 * time.Hour)) {
			out = append(out, r)
		}
	}
	if len(out) > 20 {
		out = out[len(out)-20:]
	}
	return out
}

// recordRun adds a sanitized operation outcome to an instance history.
func (s *Service) recordRun(i *model.Instance, o model.Operation, outcome, message string) {
	runs := []model.RunSummary{}
	for _, r := range i.Runs {
		if r.OperationID != o.ID {
			runs = append(runs, r)
		}
	}
	runs = append(runs, model.RunSummary{OperationID: o.ID, Kind: o.Kind, Trigger: o.Trigger, Outcome: outcome, Message: message, At: s.clock.Now().UTC()})
	i.Runs = recentRuns(runs, s.clock.Now())
}
