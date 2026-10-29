package api

import (
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"node-herder/internal/auth"
	"node-herder/internal/integrations"
	"node-herder/models/devices"
	model "node-herder/models/integrations"
	"node-herder/models/settings"
	"node-herder/repository"
)

type homeDeviceFixture struct{}

func (homeDeviceFixture) FindDeviceById(id string) (*devices.Device, error) {
	if id != "lamp" {
		return nil, errors.New("missing")
	}
	return &devices.Device{Id: id, Exposes: map[string]*devices.Entity{"state": {Name: "state"}}}, nil
}

type homeIntegrationFixture struct {
	reads int
}

func (f *homeIntegrationFixture) Instances() ([]integrations.InstanceDTO, error) {
	f.reads++
	return []integrations.InstanceDTO{{ID: "bins", Sources: []model.SourceReference{{Kind: "waste_summary", ResourceID: "calendar"}, {Kind: "waste_service", ResourceID: "calendar", ServiceID: "black"}}}}, nil
}

func homeAPIFixture(t *testing.T) (*Router, *settings.AppConfigCache, string, *homeIntegrationFixture) {
	t.Helper()
	t.Setenv("JWT_SECRET_KEY", "home-isolated-test")
	t.Setenv("FRONTEND_ALLOWED_ORIGINS", "https://ui.example.test,")
	repo, err := repository.NewFileSettingsRepoFromFile(t.TempDir() + "/settings.db")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	cache, err := settings.NewAppConfigCache(repo, nil)
	if err != nil {
		t.Fatal(err)
	}
	jwt := auth.NewJWTService(auth.WithDefaultJWTConfig())
	token, err := jwt.GenerateJWT("home-test", "home@example.test")
	if err != nil {
		t.Fatal(err)
	}
	sources := &homeIntegrationFixture{}
	handler := NewHomeLayoutHandler(cache, homeDeviceFixture{}, sources)
	router := NewRouter()
	router.UseGlobal(CORS)
	router.UseProtected(auth.Auth(jwt, nil))
	router.GET("/api/home/layout", handler)
	router.PUT("/api/home/layout", handler)
	return router, cache, token, sources
}

func TestHomeAPIAuthCSRFSourceAdmissionAndConflicts(t *testing.T) {
	router, cache, token, sources := homeAPIFixture(t)
	const path = "/api/home/layout"
	if got := integrationRequest(router, "GET", path, "", false, "", nil); got.Code != 401 {
		t.Fatal("unauthenticated read", got.Code)
	}
	read := integrationRequest(router, "GET", path, token, false, "", nil)
	if read.Code != 200 || read.Header().Get("Cache-Control") != "private, no-store" || read.Header().Get("ETag") == "" || sources.reads != 0 {
		t.Fatal("GET must remain private/read-only", read.Code, read.Body.String())
	}
	initial, _ := cache.LoadHomeLayout()
	proposal := initial.Clone()
	proposal.Blocks = []settings.HomeBlock{{ID: "group", Kind: "group", Name: "Kitchen", Cards: []settings.HomeCard{
		{ID: "light", Kind: "device", DeviceID: "lamp", Expose: "state", Panel: true},
		{ID: "black-tile", Kind: "waste_service", InstanceID: "bins", ResourceID: "calendar", ServiceID: "black", Panel: true},
	}}}
	body, _ := json.Marshal(proposal)
	for _, origin := range []string{"", "null", "https://evil.test"} {
		got := integrationRequest(router, "PUT", path, token, true, string(body), map[string]string{"If-Match": initial.ETag, "Origin": origin, "X-NodeHerder-Request": "1"})
		if got.Code != 403 {
			t.Fatal("cookie CSRF accepted", origin, got.Code, got.Body.String())
		}
	}
	if got := integrationRequest(router, "PUT", path, token, false, string(body), nil); got.Code != 428 {
		t.Fatal("missing fence", got.Code)
	}
	saved := integrationRequest(router, "PUT", path, token, true, string(body), map[string]string{"If-Match": initial.ETag, "Origin": "https://ui.example.test", "X-NodeHerder-Request": "1"})
	if saved.Code != 200 || sources.reads != 1 {
		t.Fatal("valid save", saved.Code, saved.Body.String(), sources.reads)
	}
	if got := integrationRequest(router, "PUT", path, token, false, string(body), map[string]string{"If-Match": initial.ETag}); got.Code != 409 {
		t.Fatal("lost concurrent write", got.Code)
	}
	current, _ := cache.LoadHomeLayout()
	proposal = current.Clone()
	proposal.Blocks[0].Cards[1].ServiceID = "unknown-service"
	body, _ = json.Marshal(proposal)
	got := integrationRequest(router, "PUT", path, token, false, string(body), map[string]string{"If-Match": current.ETag})
	if got.Code != 400 {
		t.Fatal("unknown source admitted", got.Code, got.Body.String())
	}
	after, _ := cache.LoadHomeLayout()
	if after.ETag != current.ETag {
		t.Fatal("invalid proposal partially published")
	}
	got = integrationRequest(router, "OPTIONS", path, "", false, "", map[string]string{"Origin": "https://ui.example.test", "Access-Control-Request-Method": "PUT"})
	if !strings.Contains(got.Header().Get("Access-Control-Allow-Headers"), "If-Match") {
		t.Fatal("missing Home preflight headers", got.Header())
	}
}

func TestHomeAPIMalformedBodiesAndAcceptedMissingSources(t *testing.T) {
	router, cache, token, _ := homeAPIFixture(t)
	layout, _ := cache.LoadHomeLayout()
	for _, body := range []string{`null`, `{}`, `{"version":1,"blocks":null}`, `{"version":1,"blocks":[],"unexpected":"private"}`, `{} {}`} {
		got := integrationRequest(router, "PUT", "/api/home/layout", token, false, body, map[string]string{"If-Match": layout.ETag})
		if got.Code != 400 {
			t.Fatal("malformed body accepted", body, got.Code)
		}
	}
	layout.Blocks = []settings.HomeBlock{{ID: "missing", Kind: "card", Cards: []settings.HomeCard{{ID: "accepted-missing", Kind: "waste_service", InstanceID: "deleted", ResourceID: "old", ServiceID: "gone"}}}}
	current, err := cache.SaveHomeLayout(layout, layout.ETag)
	if err != nil {
		t.Fatal(err)
	}
	// No optional integration runtime is required to move/hide already accepted references.
	handler := NewHomeLayoutHandler(cache, nil, nil)
	router.PUT("/api/home/layout", handler)
	current.Blocks[0].Cards[0].Panel = true
	body, _ := json.Marshal(current)
	got := integrationRequest(router, "PUT", "/api/home/layout", token, false, string(body), map[string]string{"If-Match": current.ETag})
	if got.Code != 200 {
		t.Fatal("accepted missing ref cannot be retained", got.Code, got.Body.String())
	}
}

func TestHomeIntegrationSourceIdentity(t *testing.T) {
	instances := []integrations.InstanceDTO{{ID: "external", Sources: []model.SourceReference{{Kind: "sample_capability", ResourceID: "reading", ServiceID: "channel"}}}}
	card := settings.HomeCard{Kind: "sample_capability", InstanceID: "external", ResourceID: "reading", ServiceID: "channel"}
	if !homeIntegrationExists(card, instances) {
		t.Fatal("opaque capability source unavailable")
	}
	card.ServiceID = "other"
	if homeIntegrationExists(card, instances) {
		t.Fatal("wrong source identity accepted")
	}
	card.ServiceID = "channel"
	card.Kind = "other_capability"
	if homeIntegrationExists(card, instances) {
		t.Fatal("wrong capability source accepted")
	}
}
