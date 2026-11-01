// Package deviceproviders defines shared boundaries for physical device
// integrations.
//
// A driver validates one encrypted enrollment and constructs an instance without
// network I/O. An instance owns its own lifecycle and command transport; the
// manager starts at most one instance per enrollment and never routes device
// state through a second authoritative store. Files are split by responsibility:
//
//	endpoint.go              validated private numeric endpoint
//	enrollments.go           encrypted enrollment repository
//	enrollment_bootstrap.go  keyring and storage startup
//	manager.go               worker lifecycle, identity and command routing
//	constants.go             shared status values and bounds
package deviceproviders
