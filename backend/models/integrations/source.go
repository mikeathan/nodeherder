package integrations

// SourceReference is an opaque, capability-owned display source within a resource.
// Kind and optional ServiceID retain persisted placement identities; consumers compare
// references without interpreting the capability's payload or source vocabulary.
type SourceReference struct {
	Kind       string `json:"kind"`
	ResourceID string `json:"resource_id"`
	ServiceID  string `json:"service_id,omitempty"`
}
