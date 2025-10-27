package auth

import (
	"net/http"
)


to be refactored later
// SwitchOAuth is a tiny delegator that routes to online or offline auth
// based on a simple predicate. If decide(r) returns true, offline is used;
// otherwise online is used.
type SwitchOAuth struct {
	online  OAuth
	offline OAuth
	decide  func(*http.Request) bool
}

func NewSwitchOAuth(online, offline OAuth, decide func(*http.Request) bool) *SwitchOAuth {
	return &SwitchOAuth{online: online, offline: offline, decide: decide}
}

func (s *SwitchOAuth) pick(r *http.Request) OAuth {
	if s.decide != nil && s.decide(r) {
		return s.offline
	}
	return s.online
}

func (s *SwitchOAuth) HandleLogin(w http.ResponseWriter, r *http.Request) {
	s.pick(r).HandleLogin(w, r)
}
func (s *SwitchOAuth) HandleCallback(w http.ResponseWriter, r *http.Request) {
	s.pick(r).HandleCallback(w, r)
}
func (s *SwitchOAuth) HandleLogout(w http.ResponseWriter, r *http.Request) {
	s.pick(r).HandleLogout(w, r)
}
func (s *SwitchOAuth) HandleMe(w http.ResponseWriter, r *http.Request) { s.pick(r).HandleMe(w, r) }

func (s *SwitchOAuth) RegisterRoutes(registrar RouteRegistrar) {
	base := "/api/auth"
	registrar.PublicPOST(base+"/login", http.HandlerFunc(s.HandleLogin))
	registrar.PublicPOST(base+"/logout", http.HandlerFunc(s.HandleLogout))
	registrar.PublicGET(base+"/callback", http.HandlerFunc(s.HandleCallback))
	registrar.GET(base+"/me", http.HandlerFunc(s.HandleMe))
}
