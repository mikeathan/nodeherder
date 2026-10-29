package operations

// Selects a provider or search destination and decides whether credentials can be forwarded.

import (
	"net/url"

	"node-herder/internal/integrations/destination"
	model "node-herder/models/integrations"
)

// requestDestination keeps the definition and credential context immutable. A declared
// external search service is a public-data request, so credential forwarding is explicit.
func requestDestination(attempt Attempt) (*url.URL, bool, error) {
	provider, err := destination.Parse(attempt.Definition.BaseURL)
	if err != nil {
		return nil, false, err
	}
	if destination.Origin(provider) != attempt.ApprovedOrigin {
		return nil, false, model.Fail("reconnect_required", "Review the provider destination before sending saved inputs.")
	}
	search := attempt.Definition.Setup.Search
	baseURL := search.BaseURL
	if search.Postcode != nil && attempt.RequestID == search.Postcode.Request {
		baseURL = search.Postcode.BaseURL
	} else if attempt.RequestID != search.Request {
		return provider, true, nil
	}
	if baseURL == "" {
		return provider, true, nil
	}
	searchDestination, err := destination.Parse(baseURL)
	if err != nil {
		return nil, false, err
	}
	return searchDestination, destination.Origin(searchDestination) == attempt.ApprovedOrigin, nil
}
