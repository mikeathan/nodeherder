package auth

// AuthMode identifies the sign-in path offered for a request.
type AuthMode string

const (
	AuthModeGoogle AuthMode = "google"
	AuthModeLocal  AuthMode = "local"
)

type authModeResponse struct {
	Mode AuthMode `json:"mode"`
}
