package integrations

import (
	"node-herder/internal/integrations/definitions"
	"node-herder/internal/integrations/runtime"
	model "node-herder/models/integrations"
)

// CatalogDTO exposes installed integration categories and provider availability.
type CatalogDTO = runtime.CatalogDTO

// TypeDTO describes an installed integration category.
type TypeDTO = runtime.TypeDTO

// ProviderDTO describes a host provider definition without private inputs.
type ProviderDTO = runtime.ProviderDTO

// SetupDTO exposes redacted setup/preview progress.
type SetupDTO = runtime.SetupDTO

// InstanceDTO exposes a saved integration and versioned public resources.
type InstanceDTO = runtime.InstanceDTO

// OperationDTO exposes durable operation progress without request bytes or credentials.
type OperationDTO = runtime.OperationDTO

// DefinitionDTO exposes editable host definition source and active revision.
type DefinitionDTO = runtime.DefinitionDTO

// ValidationDTO describes validation and destination-review requirements.
type ValidationDTO = runtime.ValidationDTO

// CredentialReview binds destination consent to exact credential and instance revisions.
type CredentialReview = runtime.CredentialReview

// DestinationReview records explicit consent outside provider YAML.
type DestinationReview = runtime.DestinationReview

// BoardEntry identifies one integration's order/visibility preference.
type BoardEntry = runtime.BoardEntry

// BoardDTO exposes the shared board's preferences and conditional edit token.
type BoardDTO = runtime.BoardDTO

// DefinitionFiles is the host definition repository boundary.
type DefinitionFiles = runtime.DefinitionFiles

// CompiledDefinition is an immutable validated provider revision.
type CompiledDefinition = definitions.Compiled

// MaxDefinitionBytes bounds host YAML input to 64 KiB.
const MaxDefinitionBytes = definitions.MaxDefinitionBytes

// DefinitionKey derives a stable type/provider identifier from validated metadata.
func DefinitionKey(d model.Definition) string { return definitions.DefinitionKey(d) }
