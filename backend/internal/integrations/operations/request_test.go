package operations

// Checks provider request construction, responses and captured values.

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	model "node-herder/models/integrations"
)

func TestExecuteQueryKeySanitizesTransportFailure(t *testing.T) {
	d := model.Definition{BaseURL: "https://example.com", Auth: model.AuthConfig{Mode: "apikey_query"}, Requests: map[string]model.Request{"forecast": {Method: "GET", Path: []model.Value{{Literal: "forecast"}}}}}
	a := Attempt{Definition: d, RequestID: "forecast", ApprovedOrigin: "https://example.com", Credential: &model.Credential{Origin: "https://example.com", Value: "private-key"}}
	r, _, err := Build(context.Background(), a)
	if err != nil || r.URL.Query().Get("key") != "private-key" || r.Header.Get("Authorization") != "" {
		t.Fatal("invalid query authentication", err)
	}
	client := doerFunc(func(request *http.Request) (*http.Response, error) {
		return nil, fmt.Errorf("request failed: %s", request.URL)
	})
	_, err = Execute(context.Background(), client, a, time.Now())
	if err == nil || strings.Contains(err.Error(), "private-key") {
		t.Fatal("transport failure leaked query credential")
	}
	a.Credential = nil
	if _, _, err = Build(context.Background(), a); err == nil {
		t.Fatal("missing key accepted")
	}
}

type doerFunc func(*http.Request) (*http.Response, error)

func (f doerFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }

func TestBuildAPIKeyHeader(t *testing.T) {
	d := model.Definition{BaseURL: "https://example.com", Auth: model.AuthConfig{Mode: "apikey_header"}, Requests: map[string]model.Request{"forecast": {Method: "GET", Path: []model.Value{{Literal: "forecast"}}}}}
	a := Attempt{Definition: d, RequestID: "forecast", ApprovedOrigin: "https://example.com"}
	if _, _, err := Build(context.Background(), a); err == nil {
		t.Fatal("missing key accepted")
	}
	a.Credential = &model.Credential{Value: "secret", Origin: "https://example.com"}
	r, _, err := Build(context.Background(), a)
	if err != nil || r.Header.Get("apikey") != "secret" || r.Header.Get("Authorization") != "" || strings.Contains(r.URL.String(), "secret") {
		t.Fatal("key was not restricted to its header", err)
	}
	a.Credential.Origin = "https://other.example"
	if _, _, err = Build(context.Background(), a); err == nil {
		t.Fatal("key sent to an unapproved origin")
	}
	a.Credential.Origin = a.ApprovedOrigin
	a.Credential.Value = "secret\r\nInjected: value"
	if _, _, err = Build(context.Background(), a); err == nil {
		t.Fatal("invalid header value accepted")
	}
}
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

func TestCombinedForecastAndPublicLocationSearch(t *testing.T) {
	d := model.Definition{BaseURL: "https://forecast.example", Auth: model.AuthConfig{Mode: "apikey_header"}, Setup: model.SetupRecipe{Search: model.SearchStep{Request: "search", BaseURL: "https://search.example"}}, Requests: map[string]model.Request{
		"search":   {Method: "GET", Path: []model.Value{{Literal: "locations"}}},
		"forecast": {Method: "GET", Path: []model.Value{{Literal: "hourly"}}, Combine: "daily"},
		"daily":    {Method: "GET", Path: []model.Value{{Literal: "daily"}}},
	}}
	a := Attempt{Definition: d, ApprovedOrigin: d.BaseURL, RequestID: "search", Credential: &model.Credential{Value: "secret", Origin: d.BaseURL}}
	r, _, err := Build(context.Background(), a)
	if err != nil || r.URL.Host != "search.example" || r.Header.Get("apikey") != "" {
		t.Fatal("cross-origin search exposed a credential", err)
	}
	a.RequestID = "forecast"
	calls := 0
	failDaily := false
	client := doerFunc(func(r *http.Request) (*http.Response, error) {
		calls++
		if r.URL.Host != "forecast.example" || r.Header.Get("apikey") != "secret" {
			t.Fatal("invalid forecast destination/auth")
		}
		status := 200
		if failDaily && strings.HasSuffix(r.URL.Path, "daily") {
			status = 429
		}
		return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"data":1}`))}, nil
	})
	result, err := Execute(context.Background(), client, a, time.Now())
	if err != nil || calls != 2 || result.Root.(map[string]any)["additional"] == nil {
		t.Fatal("combined response missing", err)
	}
	failDaily = true
	result, err = Execute(context.Background(), client, a, time.Now())
	if err != nil || result.Status != 429 {
		t.Fatal("child failure was hidden", err)
	}
}

func TestCaptureTypedValues(t *testing.T) {
	request := model.Request{Captures: map[string]model.Capture{
		"number":    {Pointer: "/number", Type: "number"},
		"date":      {Pointer: "/date", Type: "local-date"},
		"timestamp": {Pointer: "/timestamp", Type: "utc-timestamp"},
	}}
	body := map[string]any{"number": json.Number("42"), "date": "2026-03-29", "timestamp": "2026-03-29T13:00:00+01:00"}
	values, err := Capture(body, request, nil, true)
	if err != nil || values["timestamp"] != "2026-03-29T12:00:00Z" {
		t.Fatalf("capture: %#v, %v", values, err)
	}
	body["date"] = "2026-02-30"
	if _, err := Capture(body, request, nil, true); err == nil {
		t.Fatal("accepted invalid calendar date")
	}
}
