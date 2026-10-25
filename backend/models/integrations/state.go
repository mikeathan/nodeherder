package integrations

import "time"

// These records are private repository aggregates; controllers explicitly construct public DTOs.
type Choice struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}
type Failure struct {
	ContentHash string       `json:"content_hash,omitempty"`
	ActiveHash  string       `json:"active_hash,omitempty"`
	RetryAction string       `json:"retry_action,omitempty"`
	Code        string       `json:"code"`
	Message     string       `json:"message"`
	RetryAt     *time.Time   `json:"retry_at,omitempty"`
	Fields      []FieldError `json:"field_errors,omitempty"`
}
type FieldError struct {
	Path    string `json:"path"`
	Message string `json:"message"`
	Line    int    `json:"line,omitempty"`
	Column  int    `json:"column,omitempty"`
}

func (e *Failure) Error() string         { return e.Code + ": " + e.Message }
func Fail(code, message string) *Failure { return &Failure{Code: code, Message: message} }

type Credential struct {
	Value    string `json:"value"`
	Origin   string `json:"origin"`
	Revision uint64 `json:"revision"`
}
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
type RunSummary struct {
	OperationID string    `json:"operation_id"`
	Kind        string    `json:"kind"`
	Outcome     string    `json:"outcome"`
	Message     string    `json:"message"`
	At          time.Time `json:"at"`
}
type Instance struct {
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
	Snapshot        *WasteSchedule        `json:"snapshot,omitempty"`
	ServiceIDs      map[string]string     `json:"service_ids"`
	ETag            string                `json:"etag,omitempty"`
	CacheScope      string                `json:"cache_scope,omitempty"`
	OperationID     string                `json:"operation_id,omitempty"`
	LastRenewal     *time.Time            `json:"last_renewal,omitempty"`
	LastError       *Failure              `json:"last_error,omitempty"`
	CreatedAt       time.Time             `json:"created_at"`
	UpdatedAt       time.Time             `json:"updated_at"`
}
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
	Preview            *WasteSchedule        `json:"preview,omitempty"`
	ConnectionName     string                `json:"connection_name,omitempty"`
	SubmissionHash     string                `json:"submission_hash,omitempty"`
}
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
