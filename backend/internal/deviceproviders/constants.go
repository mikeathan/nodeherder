package deviceproviders

// Connection status values shared with provider health reporting and the UI.
const (
	StatusConnecting = "connecting"
	StatusOnline     = "online"
	StatusOffline    = "offline"
)

// DiagnosticProviderStopped reports that a configured worker is no longer running.
const DiagnosticProviderStopped = "provider_stopped"

// Port and enrollment limits. These bounds are shared by HTTP validation, the
// repository and provider drivers so a value cannot be accepted at one layer
// and rejected at another.
const (
	MinPort = 1
	MaxPort = 65535

	maxEnrollments          = 32
	maxProviderNameBytes    = 64
	providerNameCharacters  = "abcdefghijklmnopqrstuvwxyz0123456789_-"
	maxCredentialBytes      = 1024
	maxEnrollmentNameBytes  = 256
	maxPhysicalIDBytes      = 128
	maxEntityBindings       = 1024
	maxEntityIDBytes        = 256
	maxEntityKindBytes      = 64
	enrollmentRecordVersion = 1
)
