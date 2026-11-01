package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"node-herder/internal/auth"
	"node-herder/internal/deviceproviders"
	"node-herder/internal/deviceproviders/esphome"
)

type fakeESPHomeEnrollments struct {
	items   []deviceproviders.Enrollment
	update  deviceproviders.EnrollmentUpdate
	removed string
}

func TestESPHomeDiscoveryRequiresAuthorizationAndDoesNotEnroll(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "isolated-discovery-test")
	t.Setenv("FRONTEND_ALLOWED_ORIGINS", "https://ui.example.test")
	jwt := auth.NewJWTService(auth.WithDefaultJWTConfig())
	token, err := jwt.GenerateJWT("user", "user@example.test")
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	handler := auth.Auth(jwt, nil)(newESPHomeDiscoveryHandler(func(context.Context) ([]esphome.Candidate, error) {
		calls++
		return []esphome.Candidate{{Name: "fixture", Address: "192.168.1.2", Port: 6053}}, nil
	}))
	if w := integrationRequest(handler, http.MethodPost, esphomeDiscoveryPath, "", false, "", nil); w.Code != 401 || calls != 0 {
		t.Fatalf("unauthenticated: %d", w.Code)
	}
	if w := integrationRequest(handler, http.MethodPost, esphomeDiscoveryPath, token, true, "", nil); w.Code != 403 || calls != 0 {
		t.Fatalf("CSRF: %d", w.Code)
	}
	w := integrationRequest(handler, http.MethodPost, esphomeDiscoveryPath, token, false, "", nil)
	if w.Code != 200 || calls != 1 || !bytes.Contains(w.Body.Bytes(), []byte("192.168.1.2")) {
		t.Fatalf("discovery: %d %s", w.Code, w.Body.String())
	}
}

func (f *fakeESPHomeEnrollments) List() ([]deviceproviders.Enrollment, error) { return f.items, nil }
func (f *fakeESPHomeEnrollments) Add(item deviceproviders.Enrollment) (deviceproviders.Enrollment, error) {
	item.ID = "fixture"
	f.items = append(f.items, item)
	return item, nil
}
func (f *fakeESPHomeEnrollments) Remove(_ context.Context, id string) error {
	f.removed = id
	f.items = nil
	return nil
}
func (f *fakeESPHomeEnrollments) Update(_ context.Context, id string, value deviceproviders.EnrollmentUpdate) (deviceproviders.Enrollment, error) {
	f.update = value
	return deviceproviders.Enrollment{ID: id}, nil
}

func TestESPHomeMutationRoutesPreserveOmittedCredential(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "isolated-esphome-update-test")
	jwt := auth.NewJWTService(auth.WithDefaultJWTConfig())
	token, err := jwt.GenerateJWT("user", "user@example.test")
	if err != nil {
		t.Fatal(err)
	}
	service := &fakeESPHomeEnrollments{}
	router := NewRouter()
	router.UseProtected(auth.Auth(jwt, nil))
	handler := newESPHomeHandler(service)
	router.PUT(esphomeDevicePath, handler)
	router.DELETE(esphomeDevicePath, handler)
	id := uuid.NewString()
	path := esphomeDevicesPath + "/" + id
	w := integrationRequest(router, http.MethodPut, path, token, false, `{"address":"192.168.1.2","port":6053}`, nil)
	if w.Code != 200 || service.update.Credential != nil {
		t.Fatalf("update: %d %s", w.Code, w.Body.String())
	}
	w = integrationRequest(router, http.MethodDelete, path, token, false, "", nil)
	if w.Code != 200 || service.removed != id {
		t.Fatalf("delete: %d %s", w.Code, w.Body.String())
	}
}

func TestESPHomeOnboardingAuthenticationAndRedaction(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "isolated-esphome-test")
	t.Setenv("FRONTEND_ALLOWED_ORIGINS", "https://ui.example.test")
	jwt := auth.NewJWTService(auth.WithDefaultJWTConfig())
	token, err := jwt.GenerateJWT("user", "user@example.test")
	if err != nil {
		t.Fatal(err)
	}
	service := &fakeESPHomeEnrollments{}
	handler := auth.Auth(jwt, nil)(newESPHomeHandler(service))
	path := esphomeDevicesPath
	key := base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{9}, 32))
	body := `{"name":"Radar","address":"192.168.50.160","port":6053,"encryption_key":"` + key + `"}`
	if w := integrationRequest(handler, http.MethodPost, path, "", false, body, nil); w.Code != 401 {
		t.Fatalf("unauthenticated: %d", w.Code)
	}
	if w := integrationRequest(handler, http.MethodPost, path, token, true, body, nil); w.Code != 403 {
		t.Fatalf("CSRF: %d", w.Code)
	}
	if w := integrationRequest(handler, http.MethodPost, path, token, false, `{"address":"127.0.0.1","port":6053}`, nil); w.Code != 400 {
		t.Fatalf("unsafe target: %d", w.Code)
	}
	if w := integrationRequest(handler, http.MethodPost, path, token, false, `{"address":"192.168.1.2","port":6053,"encryption_key":"bad-secret"}`, nil); w.Code != 400 || bytes.Contains(w.Body.Bytes(), []byte("bad-secret")) {
		t.Fatalf("invalid key: %d %s", w.Code, w.Body.String())
	}
	w := integrationRequest(handler, http.MethodPost, path, token, true, body, map[string]string{"Origin": "https://ui.example.test", "X-NodeHerder-Request": "1"})
	if w.Code != 201 || bytes.Contains(w.Body.Bytes(), []byte(key)) || len(service.items) != 1 {
		t.Fatalf("create: %d %s", w.Code, w.Body.String())
	}
	w = integrationRequest(handler, http.MethodGet, path, token, false, "", nil)
	if w.Code != 200 || bytes.Contains(w.Body.Bytes(), []byte(key)) {
		t.Fatalf("list: %d %s", w.Code, w.Body.String())
	}
}
