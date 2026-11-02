package operations

// Checks bounded same-origin Atom snapshot discovery.

import (
	"context"
	"net/http"
	"testing"
	"time"

	model "node-herder/models/integrations"
)

func TestAtomDiscoveryFollowsOneSameOriginLink(t *testing.T) {
	definition := model.Definition{
		BaseURL: "https://alerts.example",
		Auth:    model.AuthConfig{Mode: "none"},
		Requests: map[string]model.Request{
			"warnings": {
				Method:   "GET",
				Path:     []model.Value{{Literal: "feed"}},
				Discover: &model.DiscoverConfig{Format: "atom", Rel: "related", SameOrigin: true},
			},
		},
	}
	requests := 0
	client := doerFunc(func(request *http.Request) (*http.Response, error) {
		requests++
		if request.URL.Path == "/feed" {
			return testResponse(http.StatusOK, `<feed><link rel="related" href="https://alerts.example/snapshot"/></feed>`), nil
		}
		if request.URL.Path == "/snapshot" {
			return testResponse(http.StatusOK, `{"ok":true}`), nil
		}
		t.Fatalf("unexpected URL %s", request.URL)
		return nil, nil
	})
	response, err := Execute(context.Background(), client, Attempt{
		Definition: definition, RequestID: "warnings", ApprovedOrigin: "https://alerts.example",
	}, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if requests != 2 || response.Root == nil {
		t.Fatalf("requests = %d, response = %+v", requests, response)
	}
}

func TestAtomDiscoveryRejectsForeignOrigin(t *testing.T) {
	definition := model.Definition{
		BaseURL: "https://alerts.example",
		Auth:    model.AuthConfig{Mode: "none"},
		Requests: map[string]model.Request{
			"warnings": {
				Method:   "GET",
				Path:     []model.Value{{Literal: "feed"}},
				Discover: &model.DiscoverConfig{Format: "atom", Rel: "related", SameOrigin: true},
			},
		},
	}
	client := doerFunc(func(*http.Request) (*http.Response, error) {
		return testResponse(http.StatusOK, `<feed><link rel="related" href="https://elsewhere.example/snapshot"/></feed>`), nil
	})
	if _, err := Execute(context.Background(), client, Attempt{
		Definition: definition, RequestID: "warnings", ApprovedOrigin: "https://alerts.example",
	}, time.Now()); err == nil {
		t.Fatal("foreign discovery origin accepted")
	}
}
