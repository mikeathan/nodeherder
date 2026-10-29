package operations

// Checks combined provider requests and their validation.

import (
	"context"
	"net/http"
	"testing"
	"time"

	model "node-herder/models/integrations"
)

func TestInvalidCombinationDoesNotContactProvider(t *testing.T) {
	for _, scenario := range []string{"self", "missing", "nested", "conditional", "post", "search"} {
		t.Run(scenario, func(t *testing.T) {
			definition := model.Definition{BaseURL: "https://example.com", Auth: model.AuthConfig{Mode: "none"}, Requests: map[string]model.Request{"primary": {Method: "GET", Path: []model.Value{{Literal: "primary"}}, Combine: "additional"}, "additional": {Method: "GET", Path: []model.Value{{Literal: "additional"}}}}}
			primary, additional := definition.Requests["primary"], definition.Requests["additional"]
			switch scenario {
			case "self":
				primary.Combine = "primary"
			case "missing":
				primary.Combine = "missing"
			case "nested":
				additional.Combine = "primary"
			case "conditional":
				additional.Conditional = true
			case "post":
				additional.Method = "POST"
			case "search":
				definition.Setup.Search.Request = "additional"
			}
			definition.Requests["primary"] = primary
			definition.Requests["additional"] = additional
			calls := 0
			client := doerFunc(func(*http.Request) (*http.Response, error) { calls++; return nil, nil })
			_, err := Execute(context.Background(), client, Attempt{Definition: definition, RequestID: "primary", ApprovedOrigin: definition.BaseURL}, time.Now())
			if err == nil || calls != 0 {
				t.Fatal("invalid recipe contacted provider", err, calls)
			}
		})
	}
}
