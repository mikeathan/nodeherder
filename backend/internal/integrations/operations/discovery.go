package operations

// Resolves one same-origin Atom link immediately before a declared request.

import (
	"context"
	"encoding/xml"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"node-herder/internal/integrations/destination"
	model "node-herder/models/integrations"
)

const discoveryMaxBytes = 256 << 10

type atomLink struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}
type atomFeed struct {
	Links []atomLink `xml:"link"`
}

// executeDiscovery fetches the declared feed and replaces only this invocation's URL.
func executeDiscovery(ctx context.Context, client Doer, attempt Attempt, now time.Time) (Attempt, error) {
	request, _, err := Build(ctx, attempt)
	if err != nil {
		return attempt, err
	}
	response, err := client.Do(request)
	if err != nil {
		return attempt, model.Fail("provider_unavailable", "The warning feed could not be reached. Try again later.")
	}
	if response == nil || response.Body == nil {
		return attempt, model.Fail("invalid_response", "The warning feed returned no response.")
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusUnauthorized || response.StatusCode == http.StatusForbidden {
		return attempt, model.Fail("reconnect_required", "The provider rejected the warning feed credential.")
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return attempt, model.Fail("provider_unavailable", "The warning feed is unavailable. Try again later.")
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, discoveryMaxBytes+1))
	if err != nil || len(body) > discoveryMaxBytes {
		return attempt, model.Fail("invalid_response", "The warning feed exceeded its size limit.")
	}
	var feed atomFeed
	if err = xml.Unmarshal(body, &feed); err != nil {
		return attempt, model.Fail("invalid_response", "The warning feed was not valid Atom XML.")
	}
	rel := attempt.Definition.Requests[attempt.RequestID].Discover.Rel
	var href string
	for _, link := range feed.Links {
		if link.Rel != rel {
			continue
		}
		if href != "" {
			return attempt, model.Fail("invalid_response", "The warning feed published more than one snapshot link.")
		}
		href = strings.TrimSpace(link.Href)
	}
	if href == "" {
		return attempt, model.Fail("invalid_response", "The warning feed did not publish a snapshot link.")
	}
	target, err := url.Parse(href)
	if err != nil {
		return attempt, model.Fail("invalid_response", "The warning feed published an invalid snapshot link.")
	}
	if !target.IsAbs() {
		target = request.URL.ResolveReference(target)
	}
	base, _ := url.Parse(attempt.Definition.BaseURL)
	if target.Scheme != "https" || target.Hostname() == "" || target.User != nil || target.Fragment != "" ||
		(target.Port() != "" && target.Port() != "443") || destination.Origin(target) != destination.Origin(base) {
		return attempt, model.Fail("destination_blocked", "The warning feed published a link outside the provider origin.")
	}
	attempt.DiscoveredURL = target
	return attempt, nil
}
