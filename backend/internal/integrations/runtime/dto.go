package runtime

// Defines sanitized integration data returned to API and application callers.

import (
	"time"

	model "node-herder/models/integrations"
)

// CatalogDTO lists installed categories/providers without private inputs, attempts or secrets.
type CatalogDTO struct {
	Available bool           `json:"available"`
	Error     *model.Failure `json:"availability_error"`
	Types     []TypeDTO      `json:"types"`
	Providers []ProviderDTO  `json:"providers"`
}

// TypeDTO describes an installed capability category available in the Connections catalog.
type TypeDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// ProviderDTO describes local provider availability without exposing private connection inputs.
type ProviderDTO struct {
	Key             string `json:"key"`
	ID              string `json:"id"`
	Name            string `json:"name"`
	IntegrationType string `json:"integration_type"`
	Origin          string `json:"origin"`
	Available       bool   `json:"available"`
	DiskError       string `json:"disk_error,omitempty"`
}

// SetupDTO is the browser view of a shared wizard/test. It omits captured state, exact POST
// bytes and tokens; CredentialConfigured reports presence only.
type SetupDTO struct {
	Preview     *model.Resource `json:"preview,omitempty"`
	ID          string          `json:"id"`
	Revision    uint64          `json:"revision"`
	ETag        string          `json:"etag"`
	ProviderKey string          `json:"provider_key"`
	Status      string          `json:"status"`
	Choices     []model.Choice  `json:"choices"`
	// Postcode retains the v1 wire name for the saved search value; interpretation belongs to the flow.
	Postcode             string         `json:"postcode,omitempty"`
	InstanceID           string         `json:"instance_id,omitempty"`
	OperationID          string         `json:"operation_id,omitempty"`
	ExpiresAt            time.Time      `json:"expires_at"`
	CredentialConfigured bool           `json:"credential_configured"`
	Error                *model.Failure `json:"last_error"`
}

// ScheduleDTO exposes automatic-refresh state without private inputs. Nil NextDueAt means the
// schedule is off or has not been initialised; nil Last* values mean the event has not occurred.
type ScheduleDTO struct {
	Enabled             bool       `json:"enabled"`
	IntervalSeconds     int64      `json:"interval_seconds"`
	NextDueAt           *time.Time `json:"next_due_at"`
	LastAttemptAt       *time.Time `json:"last_attempt_at"`
	LastSuccessAt       *time.Time `json:"last_success_at"`
	LastFailureAt       *time.Time `json:"last_failure_at"`
	ConsecutiveFailures int        `json:"consecutive_failures"`
}

// InstanceDTO exposes saved connection status and validated resources. CredentialSlots
// contains names only so obsolete slots can be explicitly removed without revealing token values.
type InstanceDTO struct {
	Sources         []model.SourceReference `json:"sources"`
	CredentialSlots []string                `json:"credential_slots"`
	Runs            []model.RunSummary      `json:"runs"`
	Schedule        *ScheduleDTO            `json:"schedule"`
	ID              string                  `json:"id"`
	Name            string                  `json:"name"`
	ProviderKey     string                  `json:"provider_key"`
	Revision        uint64                  `json:"revision"`
	ETag            string                  `json:"etag"`
	Enabled         bool                    `json:"enabled"`
	Status          string                  `json:"status"`
	// Postcode retains the v1 wire name for the saved search value; interpretation belongs to the flow.
	Postcode             string           `json:"postcode"`
	OperationID          string           `json:"operation_id,omitempty"`
	CredentialConfigured bool             `json:"credential_configured"`
	Resources            []model.Resource `json:"resources"`
	Error                *model.Failure   `json:"last_error"`
	UpdatedAt            time.Time        `json:"updated_at"`
}

// OperationDTO exposes durable progress and retry timing, never private attempt payloads.
type OperationDTO struct {
	ID         string         `json:"id"`
	Status     string         `json:"status"`
	Kind       string         `json:"kind"`
	SetupID    string         `json:"setup_id,omitempty"`
	InstanceID string         `json:"instance_id,omitempty"`
	RetryAt    time.Time      `json:"retry_at"`
	ExpiresAt  time.Time      `json:"expires_at"`
	Error      *model.Failure `json:"last_error"`
}

// DefinitionDTO separates exact host bytes/hash from the active compiled revision.
// Status makes host changes, missing files and saved-but-not-applied revisions observable.
type DefinitionDTO struct {
	ShippedYAML     string           `json:"shipped_yaml,omitempty"`
	Key             string           `json:"key"`
	YAML            string           `json:"yaml"`
	Document        model.Definition `json:"document"`
	ContentHash     string           `json:"content_hash"`
	ActiveHash      string           `json:"active_hash"`
	Revision        uint64           `json:"revision"`
	ETag            string           `json:"etag"`
	Status          string           `json:"status"`
	DiskError       string           `json:"disk_error,omitempty"`
	UpdateAvailable bool             `json:"update_available"`
}

// ValidationDTO is a normalized unsaved definition plus the destination-review scope.
// Validation does not write the host file or send any provider request.
type ValidationDTO struct {
	YAML           string             `json:"yaml"`
	Document       model.Definition   `json:"document"`
	Hash           string             `json:"content_hash"`
	Origin         string             `json:"origin"`
	PreviousOrigin string             `json:"previous_origin"`
	Affected       []CredentialReview `json:"affected_instances"`
}

// CredentialReview pins the specific active slot and instance/credential revisions a user
// may reassociate. A stale acknowledgement cannot transfer a newly replaced or unrelated token.
type CredentialReview struct {
	CredentialSlot     string `json:"credential_slot"`
	InstanceID         string `json:"instance_id"`
	Revision           uint64 `json:"revision"`
	CredentialRevision uint64 `json:"credential_revision"`
}

// DestinationReview is explicit consent outside YAML to send private inputs and selected
// credentials to a changed origin. YAML alone cannot authorize credential reassociation.
type DestinationReview struct {
	Origin      string             `json:"origin"`
	Reassociate []CredentialReview `json:"reassociate"`
}
