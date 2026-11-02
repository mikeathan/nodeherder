package operations

// Checks OAuth client-credentials token requests, caching and one-time retry.

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	model "node-herder/models/integrations"
)

func TestOAuthClientCredentialsJSONAndRetry(t *testing.T) {
	definition := model.Definition{
		BaseURL: "https://api.fuel.example",
		Auth: model.AuthConfig{
			Mode:              "oauth2_client_credentials",
			TokenURL:          "https://auth.fuel.example/token",
			ClientIDSlot:      "client_id",
			ClientSecretSlot:  "client_secret",
			TokenStyle:        "json_body",
			TokenResponseRoot: "/data",
		},
		Requests: map[string]model.Request{"prices": {Method: "GET", Path: []model.Value{{Literal: "prices"}}}},
	}
	tokenRequests, apiRequests := 0, 0
	client := doerFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Host == "auth.fuel.example" {
			tokenRequests++
			body, _ := io.ReadAll(request.Body)
			if !strings.Contains(string(body), `"grant_type":"client_credentials"`) ||
				!strings.Contains(string(body), `"client_id":"client"`) ||
				!strings.Contains(string(body), `"client_secret":"secret"`) {
				t.Fatalf("token body = %s", body)
			}
			return testResponse(http.StatusOK, `{"data":{"access_token":"token","token_type":"Bearer","expires_in":3600}}`), nil
		}
		apiRequests++
		if request.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("authorization = %q", request.Header.Get("Authorization"))
		}
		if apiRequests == 1 {
			return testResponse(http.StatusUnauthorized, `{"error":"expired"}`), nil
		}
		return testResponse(http.StatusOK, `{"ok":true}`), nil
	})
	response, err := Execute(context.Background(), client, Attempt{
		Definition:     definition,
		RequestID:      "prices",
		ApprovedOrigin: "https://api.fuel.example",
		Credentials: map[string]model.Credential{
			"client_id":     {Value: "client", Origin: "https://auth.fuel.example"},
			"client_secret": {Value: "secret", Origin: "https://auth.fuel.example"},
		},
	}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if tokenRequests != 2 || apiRequests != 2 || response.Status != http.StatusOK {
		t.Fatalf("tokens=%d api=%d status=%d", tokenRequests, apiRequests, response.Status)
	}
}
