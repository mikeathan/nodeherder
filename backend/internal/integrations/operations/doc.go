// Package operations executes the small reusable HTTP primitives declared in provider YAML.
// It does not own scheduling, persistence, setup navigation or waste-specific validation.
//
// References such as input.postcode and state.lookup_id resolve typed saved values, not executable
// templates. JSON Pointers select response fields: /result/collections/0/name means the name on the
// first collection inside result. These paths describe JSON locations, not Go memory pointers.
// The caller persists an Attempt before I/O and decides how to handle status, retry deadlines and
// capability publication; Execute itself never silently retries a provider request.
package operations
