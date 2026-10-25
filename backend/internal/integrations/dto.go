package integrations

import (
	model "node-herder/models/integrations"
	"time"
)

// Public DTOs deliberately exclude raw attempts, captured property tokens and secret values.
type CatalogDTO struct {
	Available bool           `json:"available"`
	Error     *model.Failure `json:"availability_error"`
	Types     []TypeDTO      `json:"types"`
	Providers []ProviderDTO  `json:"providers"`
}
type TypeDTO struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}
type ProviderDTO struct {
	Key             string `json:"key"`
	ID              string `json:"id"`
	Name            string `json:"name"`
	IntegrationType string `json:"integration_type"`
	Origin          string `json:"origin"`
	Available       bool   `json:"available"`
	DiskError       string `json:"disk_error,omitempty"`
}
type SetupDTO struct {
	Preview              *model.WasteSchedule `json:"preview,omitempty"`
	ID                   string               `json:"id"`
	Revision             uint64               `json:"revision"`
	ETag                 string               `json:"etag"`
	ProviderKey          string               `json:"provider_key"`
	Status               string               `json:"status"`
	Choices              []model.Choice       `json:"choices"`
	Postcode             string               `json:"postcode,omitempty"`
	InstanceID           string               `json:"instance_id,omitempty"`
	OperationID          string               `json:"operation_id,omitempty"`
	ExpiresAt            time.Time            `json:"expires_at"`
	CredentialConfigured bool                 `json:"credential_configured"`
	Error                *model.Failure       `json:"last_error"`
}
type InstanceDTO struct {
	CredentialSlots      []string              `json:"credential_slots"`
	Runs                 []model.RunSummary    `json:"runs"`
	ID                   string                `json:"id"`
	Name                 string                `json:"name"`
	ProviderKey          string                `json:"provider_key"`
	Revision             uint64                `json:"revision"`
	ETag                 string                `json:"etag"`
	Enabled              bool                  `json:"enabled"`
	Status               string                `json:"status"`
	Postcode             string                `json:"postcode"`
	OperationID          string                `json:"operation_id,omitempty"`
	CredentialConfigured bool                  `json:"credential_configured"`
	Resources            []model.WasteSchedule `json:"resources"`
	Error                *model.Failure        `json:"last_error"`
	UpdatedAt            time.Time             `json:"updated_at"`
}
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
type ValidationDTO struct {
	YAML           string             `json:"yaml"`
	Document       model.Definition   `json:"document"`
	Hash           string             `json:"content_hash"`
	Origin         string             `json:"origin"`
	PreviousOrigin string             `json:"previous_origin"`
	Affected       []CredentialReview `json:"affected_instances"`
}
type CredentialReview struct {
	CredentialSlot     string `json:"credential_slot"`
	InstanceID         string `json:"instance_id"`
	Revision           uint64 `json:"revision"`
	CredentialRevision uint64 `json:"credential_revision"`
}
type DestinationReview struct {
	Origin      string             `json:"origin"`
	Reassociate []CredentialReview `json:"reassociate"`
}
