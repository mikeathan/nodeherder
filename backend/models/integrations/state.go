package integrations

import "time"

// These records are private repository aggregates; controllers explicitly construct public DTOs.
type Choice struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

// Failure is a sanitized domain/HTTP error. It may be persisted or sent to the browser, so
// it must not contain credentials, raw provider bodies or internal filesystem paths.
type Failure struct {
	ContentHash string       `json:"content_hash,omitempty"`
	ActiveHash  string       `json:"active_hash,omitempty"`
	RetryAction string       `json:"retry_action,omitempty"`
	Code        string       `json:"code"`
	Message     string       `json:"message"`
	RetryAt     *time.Time   `json:"retry_at,omitempty"`
	Fields      []FieldError `json:"field_errors,omitempty"`
}

// FieldError identifies an invalid definition field and optional source position for editor feedback.
type FieldError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

// Error implements error using only the sanitized failure code/message.
func (e *Failure) Error() string { return e.Code + ": " + e.Message }

// Fail constructs a domain failure safe for transport after choosing a public message.
func Fail(code, message string) *Failure { return &Failure{Code: code, Message: message} }

// Credential is private encrypted state bound to a destination origin and revision.
// It must never be serialized into a public DTO or shareable YAML.
type Credential struct {
	Value    string `json:"value"`
	Origin   string `json:"origin"`
	Revision uint64 `json:"revision"`
}

// DefinitionRecord keeps the active host source and a bounded previous revision. Pending
// source/hash is the recovery journal for a file written before runtime activation commits.
type DefinitionRecord struct {
	SeedHash          string    `json:"seed_hash,omitempty"`
	Key               string    `json:"key"`
	YAML              string    `json:"yaml"`
	Hash              string    `json:"hash"`
	Revision          uint64    `json:"revision"`
	PreviousExpiresAt time.Time `json:"previous_expires_at"`
	PreviousYAML      string    `json:"previous_yaml,omitempty"`
	PendingYAML       string    `json:"pending_yaml,omitempty"`
	PendingHash       string    `json:"pending_hash,omitempty"`
	DiskError         string    `json:"disk_error,omitempty"`
}

// RunSummary is a bounded, sanitized check history entry; it contains no provider response body.
type RunSummary struct {
	OperationID string    `json:"operation_id"`
	Kind        string    `json:"kind"`
	Outcome     string    `json:"outcome"`
	Message     string    `json:"message"`
	At          time.Time `json:"at"`
}

// Instance is the encrypted aggregate for a saved connection and its last validated snapshot.
// Revision prevents lost edits; Generation fences in-flight work after lifecycle changes.
// BindingID/ResourceID survive compatible repairs; State contains private captured provider values.
type Instance struct {
	Board           *BoardPreference      `json:"board,omitempty"`
	Runs            []RunSummary          `json:"runs"`
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	DefinitionKey   string                `json:"definition_key"`
	Revision        uint64                `json:"revision"`
	Generation      uint64                `json:"generation"`
	BindingID       string                `json:"binding_id"`
	BindingRevision uint64                `json:"binding_revision"`
	ResourceID      string                `json:"resource_id"`
	Enabled         bool                  `json:"enabled"`
	SearchInput     string                `json:"search_input"`
	Inputs          map[string]string     `json:"inputs"`
	State           map[string]any        `json:"state"`
	Credentials     map[string]Credential `json:"credentials"`
	Origin          string                `json:"origin"`
	Snapshot        *Resource             `json:"snapshot,omitempty"`
	Identities      map[string]string     `json:"service_ids"`
	ETag            string                `json:"etag,omitempty"`
	CacheScope      string                `json:"cache_scope,omitempty"`
	OperationID     string                `json:"operation_id,omitempty"`
	LastRenewal     *time.Time            `json:"last_renewal,omitempty"`
	LastError       *Failure              `json:"last_error,omitempty"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}

// BoardPreference stores only presentation choices. Its independent revision prevents a
// layout save from invalidating provider work or conflicting with an ordinary refresh.
// A nil Instance.Board represents an older/default record: visible, with creation-time order.
type BoardPreference struct {
	Order    int    `json:"order"`
	Hidden   bool   `json:"hidden"`
	Revision uint64 `json:"revision"`
}

// Setup is an encrypted, expiring shared wizard or isolated test session. Its pinned definition
// and Generation prevent old choices/requests being reused after edits. A completed setup is
// reduced to a minimal receipt so a repeated Save returns the original instance instead of a new POST.
type Setup struct {
	ID                 string                `json:"id"`
	DefinitionKey      string                `json:"definition_key"`
	DefinitionRevision uint64                `json:"definition_revision"`
	DefinitionYAML     string                `json:"definition_yaml"`
	Revision           uint64                `json:"revision"`
	Generation         uint64                `json:"generation"`
	InstanceID         string                `json:"instance_id,omitempty"`
	TargetInstanceID   string                `json:"target_instance_id,omitempty"`
	TargetRevision     uint64                `json:"target_revision,omitempty"`
	Inputs             map[string]string     `json:"inputs"`
	State              map[string]any        `json:"state"`
	Choices            []Choice              `json:"choices"`
	Credentials        map[string]Credential `json:"credentials"`
	OperationID        string                `json:"operation_id,omitempty"`
	Status             string                `json:"status"`
	LastError          *Failure              `json:"last_error,omitempty"`
	CreatedAt          time.Time             `json:"created_at"`
	ExpiresAt          time.Time             `json:"expires_at"`
	AbsoluteExpiry     time.Time             `json:"absolute_expiry"`
	TestOnly           bool                  `json:"test_only,omitempty"`
	Preview            *Resource             `json:"preview,omitempty"`
	ConnectionName     string                `json:"connection_name,omitempty"`
	SubmissionHash     string                `json:"submission_hash,omitempty"`
}

// Operation is the durable identity/checkpoint of one logical lookup, refresh or test.
// Body and IdempotencyKey must be persisted together before POST and reused byte-for-byte after
// an ambiguous response. Generation and definition/credential revisions fence obsolete publication.
type Operation struct {
	ID                 string            `json:"id"`
	InstanceID         string            `json:"instance_id,omitempty"`
	SetupID            string            `json:"setup_id,omitempty"`
	Generation         uint64            `json:"generation"`
	DefinitionYAML     string            `json:"definition_yaml"`
	RequestFingerprint string            `json:"request_fingerprint"`
	ProfileRevision    uint64            `json:"profile_revision"`
	DefinitionHash     string            `json:"definition_hash"`
	Inputs             map[string]string `json:"inputs"`
	Origin             string            `json:"origin"`
	RequestID          string            `json:"request_id"`
	Kind               string            `json:"kind"`
	Status             string            `json:"status"`
	Body               []byte            `json:"body,omitempty"`
	IdempotencyKey     string            `json:"idempotency_key,omitempty"`
	State              map[string]any    `json:"state,omitempty"`
	CredentialRevision uint64            `json:"credential_revision"`
	RetryAt            time.Time         `json:"retry_at"`
	ExpiresAt          time.Time         `json:"expires_at"`
	CreatedAt          time.Time         `json:"created_at"`
	LastError          *Failure          `json:"last_error,omitempty"`
	ETag               string            `json:"etag,omitempty"`
	CacheScope         string            `json:"cache_scope,omitempty"`
	Renewed            bool              `json:"renewed"`
}
