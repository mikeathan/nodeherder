package operations

import (
	"context"
	"io"
	"net/http"
	model "node-herder/models/integrations"
	"strings"
	"testing"
	"time"
)

type doerFunc func(*http.Request) (*http.Response, error)

func (f doerFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }
func TestBuildAndExecute(t *testing.T) {
	d := model.Definition{BaseURL: "https://example.com", Auth: model.AuthConfig{Mode: "optional_bearer"}, Requests: map[string]model.Request{"arbitrary": {Method: "GET", Path: []model.Value{{Literal: "v1"}, {Ref: "state.id"}}, Query: map[string]model.Value{"postcode": {Ref: "input.postcode"}}}}}
	a := Attempt{Definition: d, RequestID: "arbitrary", ApprovedOrigin: "https://example.com", Inputs: map[string]string{"postcode": "S8 9ED"}, State: map[string]any{"id": "a/b?x"}}
	r, _, e := Build(context.Background(), a)
	if e != nil {
		t.Fatal(e)
	}
	if r.URL.EscapedPath() != "/v1/a%2Fb%3Fx" || r.URL.Query().Get("postcode") != "S8 9ED" {
		t.Fatal(r.URL)
	}
	client := doerFunc(func(*http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 202, Header: http.Header{"Retry-After": []string{"12"}}, Body: io.NopCloser(strings.NewReader(`{"status":"running"}`))}, nil
	})
	now := time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC)
	response, e := Execute(context.Background(), client, a, now)
	if e != nil || response.Status != 202 || response.RetryAt == nil || !response.RetryAt.Equal(now.Add(12*time.Second)) {
		t.Fatal(response, e)
	}
	a.Credential = &model.Credential{Value: "secret", Origin: "https://elsewhere.test"}
	if _, _, e = Build(context.Background(), a); e == nil {
		t.Fatal("credential origin bypass")
	}
	a.Credential = nil
	a.ApprovedOrigin = "https://other.test"
	if _, _, e = Build(context.Background(), a); e == nil {
		t.Fatal("input origin bypass")
	}
}
func TestCapture(t *testing.T) {
	q := model.Request{Captures: map[string]model.Capture{"token": {Pointer: "/result/id", Type: "string", When: "complete", StripPrefix: "p:"}}}
	root := map[string]any{"result": map[string]any{"id": "p:opaque"}}
	early, e := Capture(root, q, nil, false)
	if e != nil || len(early) != 0 {
		t.Fatal(early, e)
	}
	done, e := Capture(root, q, early, true)
	if e != nil || done["token"] != "opaque" {
		t.Fatal(done, e)
	}
	root["result"].(map[string]any)["id"] = "uprn"
	if _, e = Capture(root, q, nil, true); e == nil {
		t.Fatal("namespace change accepted")
	}
}
