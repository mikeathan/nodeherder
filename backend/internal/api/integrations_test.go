package api

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"node-herder/internal/auth"
	connections "node-herder/internal/integrations"
	"node-herder/repository"
)

func integrationAPIFixture(t *testing.T) (*Router, string, string) {
	t.Helper()
	t.Setenv("JWT_SECRET_KEY", "isolated-integrations-test")
	t.Setenv("FRONTEND_ALLOWED_ORIGINS", "https://ui.example.test,")
	keys := repository.IntegrationKeyring{ActiveKeyID: "test", Keys: map[string]string{"test": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{2}, 32))}}
	repo, e := repository.OpenIntegrationRepository(t.TempDir(), keys)
	if e != nil {
		t.Fatal(e)
	}
	files, e := repository.OpenIntegrationDefinitions(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	service, e := connections.NewService(context.Background(), connections.Options{Repository: repo, Definitions: files})
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if e := service.Close(ctx); e != nil {
			t.Error(e)
		}
	})
	jwt := auth.NewJWTService(auth.WithDefaultJWTConfig())
	one, e := jwt.GenerateJWT("one", "one@example.test")
	if e != nil {
		t.Fatal(e)
	}
	two, e := jwt.GenerateJWT("two", "two@example.test")
	if e != nil {
		t.Fatal(e)
	}
	r := NewRouter()
	r.UseGlobal(CORS)
	r.UseProtected(auth.Auth(jwt, nil))
	RegisterIntegrationRoutes(r, service)
	return r, one, two
}
func integrationRequest(router http.Handler, method, path, token string, cookie bool, body string, headers map[string]string) *httptest.ResponseRecorder {
	r := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if token != "" {
		if cookie {
			r.AddCookie(&http.Cookie{Name: auth.CookieNameSession, Value: token})
		} else {
			r.Header.Set("Authorization", "Bearer "+token)
		}
	}
	for key, value := range headers {
		r.Header.Set(key, value)
	}
	w := httptest.NewRecorder()
	router.ServeHTTP(w, r)
	return w
}
func TestIntegrationAuthorizationAndRevisions(t *testing.T) {
	r, one, two := integrationAPIFixture(t)
	path := "/api/integrations/catalog"
	w := integrationRequest(r, "GET", path, "", false, "", nil)
	if w.Code != 401 || !bytes.Contains(w.Body.Bytes(), []byte(`"data":null`)) {
		t.Fatal(w.Code, w.Body.String())
	}
	w = integrationRequest(r, "POST", "/api/integrations/setup", one, false, `{"provider_key":"waste_collection/whenisbins"}`, nil)
	if w.Code != 201 {
		t.Fatal(w.Code, w.Body.String())
	}
	var envelope struct{ Data connections.SetupDTO }
	if e := json.Unmarshal(w.Body.Bytes(), &envelope); e != nil {
		t.Fatal(e)
	}
	setup := envelope.Data
	w = integrationRequest(r, "GET", "/api/integrations/setup/"+setup.ID, two, false, "", nil)
	if w.Code != 200 {
		t.Fatal("second user cannot continue shared setup", w.Code)
	}
	for _, headers := range []map[string]string{nil, {"X-NodeHerder-Request": "1"}, {"Origin": "https://evil.test", "X-NodeHerder-Request": "1"}, {"Origin": "https://ui.example.test"}} {
		w = integrationRequest(r, "DELETE", "/api/integrations/setup/"+setup.ID, two, true, "", headers)
		if w.Code != 403 {
			t.Fatal("cookie mutation admitted", w.Code)
		}
	}
	w = integrationRequest(r, "DELETE", "/api/integrations/setup/"+setup.ID, two, false, "", nil)
	if w.Code != 428 {
		t.Fatal("missing revision", w.Code, w.Body.String())
	}
	w = integrationRequest(r, "DELETE", "/api/integrations/setup/"+setup.ID, two, false, "", map[string]string{"If-Match": `"wrong"`})
	if w.Code != 409 {
		t.Fatal(w.Code)
	}
	w = integrationRequest(r, "DELETE", "/api/integrations/setup/"+setup.ID, two, true, "", map[string]string{"Origin": "https://ui.example.test", "X-NodeHerder-Request": "1", "If-Match": setup.ETag})
	if w.Code != 200 {
		t.Fatal(w.Code, w.Body.String())
	}
}
func TestIntegrationCORSAndDirectHandler(t *testing.T) {
	r, one, _ := integrationAPIFixture(t)
	w := integrationRequest(r, "OPTIONS", "/api/integrations/definitions/waste_collection/whenisbins", one, false, "", map[string]string{"Origin": "https://ui.example.test"})
	if w.Code != 204 || !bytes.Contains([]byte(w.Header().Get("Access-Control-Allow-Headers")), []byte("If-Match")) {
		t.Fatal(w.Code, w.Header())
	}
	w = integrationRequest(r, "OPTIONS", "/other", one, false, "", map[string]string{"Origin": "https://ui.example.test"})
	if bytes.Contains([]byte(w.Header().Get("Access-Control-Allow-Headers")), []byte("If-Match")) {
		t.Fatal("broadened unrelated CORS")
	}
	w = httptest.NewRecorder()
	NewIntegrationHandler(connections.DegradedService("key_unavailable"), "catalog").ServeHTTP(w, httptest.NewRequest("GET", "/api/integrations/catalog", nil))
	if w.Code != 401 {
		t.Fatal("direct handler lacks authentication")
	}
}

func TestIntegrationCredentialsAreWriteOnlyAndCookieWins(t *testing.T) {
	r, one, two := integrationAPIFixture(t)
	w := integrationRequest(r, "POST", "/api/integrations/setup", one, false, `{"provider_key":"waste_collection/whenisbins"}`, nil)
	var envelope struct{ Data connections.SetupDTO }
	if e := json.Unmarshal(w.Body.Bytes(), &envelope); e != nil {
		t.Fatal(e)
	}
	p := envelope.Data
	path := "/api/integrations/setup/" + p.ID + "/credentials/api_token"
	w = integrationRequest(r, "PUT", path, two, false, `{"value":"synthetic-private-token"}`, map[string]string{"If-Match": p.ETag})
	if w.Code != 200 || bytes.Contains(w.Body.Bytes(), []byte("synthetic-private-token")) {
		t.Fatal("secret echoed", w.Code)
	}
	w = integrationRequest(r, "GET", "/api/integrations/setup/"+p.ID, one, false, "", nil)
	if w.Code != 200 || bytes.Contains(w.Body.Bytes(), []byte("synthetic-private-token")) {
		t.Fatal("secret read back", w.Code)
	}
	w = integrationRequest(r, "PUT", path, "invalid-cookie", true, `{"value":"another-token"}`, map[string]string{"Authorization": "Bearer " + two, "If-Match": p.ETag, "Origin": "https://ui.example.test", "X-NodeHerder-Request": "1"})
	if w.Code != 401 {
		t.Fatal("bearer overrode invalid cookie", w.Code)
	}
}

func TestIntegrationBoardAuthorizationAndContract(t *testing.T) {
	router, one, two := integrationAPIFixture(t)
	const path = "/api/integrations/board"
	if w := integrationRequest(router, "GET", path, "", false, "", nil); w.Code != 401 {
		t.Fatal(w.Code)
	}
	read := integrationRequest(router, "GET", path, one, false, "", nil)
	var envelope struct{ Data connections.BoardDTO }
	if err := json.Unmarshal(read.Body.Bytes(), &envelope); err != nil {
		t.Fatal(err)
	}
	if read.Code != 200 || read.Header().Get("ETag") != envelope.Data.ETag || read.Header().Get("Cache-Control") != "private, no-store" {
		t.Fatal(read.Code, read.Header(), read.Body.String())
	}
	for _, tc := range []struct {
		body    string
		etag    string
		cookie  bool
		headers map[string]string
		status  int
	}{
		{`{"entries":[]}`, "", false, nil, 428},
		{`{"entries":[]}`, `"stale"`, false, nil, 409},
		{`{}`, envelope.Data.ETag, false, nil, 400},
		{`{"entries":null}`, envelope.Data.ETag, false, nil, 400},
		{`{"entries":[{"instance_id":"x"}]}`, envelope.Data.ETag, false, nil, 400},
		{`{"entries":[{"instance_id":"x","hidden":null}]}`, envelope.Data.ETag, false, nil, 400},
		{`{"entries":[{"instance_id":"x","hidden":false,"extra":true}]}`, envelope.Data.ETag, false, nil, 400},
		{`{"entries":[]}`, envelope.Data.ETag, true, nil, 403},
		{`{"entries":[]}`, envelope.Data.ETag, true, map[string]string{"Origin": "https://evil.test", "X-NodeHerder-Request": "1"}, 403},
		{`{"entries":[]}`, envelope.Data.ETag, true, map[string]string{"Origin": "https://ui.example.test", "X-NodeHerder-Request": "1"}, 200},
	} {
		headers := map[string]string{"If-Match": tc.etag}
		for k, v := range tc.headers {
			headers[k] = v
		}
		w := integrationRequest(router, "PUT", path, two, tc.cookie, tc.body, headers)
		if w.Code != tc.status {
			t.Fatalf("%s: want %d got %d %s", tc.body, tc.status, w.Code, w.Body.String())
		}
	}
}
