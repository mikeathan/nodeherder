package operations

// Validates and executes a bounded pair of provider requests as one result.

import (
	"context"
	"net/http"
	"time"

	model "node-herder/models/integrations"
)

// ValidateCombination checks the bounded two-GET recipe used for atomic publication.
// It rejects nested, conditional, self and search combinations before provider I/O.
// The same validator is used by configuration compilation and operation execution.
func ValidateCombination(definition model.Definition, requestID string) error {
	primary, ok := definition.Requests[requestID]
	if !ok {
		return model.InvalidDefinition("requests", "Unknown primary request.")
	}
	if primary.Combine == "" {
		return nil
	}
	invalid := func(message string) error { return model.InvalidDefinition("requests."+requestID+".combine", message) }
	if primary.Combine == requestID {
		return invalid("A request cannot combine itself.")
	}
	additional, ok := definition.Requests[primary.Combine]
	if !ok {
		return invalid("The additional request does not exist.")
	}
	if additional.Combine != "" {
		return invalid("Nested request combinations are not supported.")
	}
	search := definition.Setup.Search
	if requestID == search.Request || primary.Combine == search.Request || (search.Postcode != nil && (requestID == search.Postcode.Request || primary.Combine == search.Postcode.Request)) {
		return invalid("Location search cannot be combined with forecast requests.")
	}
	for _, request := range []model.Request{primary, additional} {
		if request.Method != http.MethodGet {
			return invalid("Combined requests must use GET.")
		}
		if request.Conditional {
			return invalid("Combined requests cannot use conditional caching.")
		}
	}
	return nil
}

func successful(response Response) bool {
	return response.Status >= http.StatusOK && response.Status < http.StatusMultipleChoices
}

// combineResponse performs only the declared additional GET using the original deadline,
// credential binding and input snapshot. A failure remains a failure of the whole operation.
func combineResponse(ctx context.Context, client Doer, attempt Attempt, primary Response, now time.Time) (Response, error) {
	additionalAttempt := attempt
	additionalAttempt.RequestID = attempt.Definition.Requests[attempt.RequestID].Combine
	additionalAttempt.ETag = ""
	additionalAttempt.Body = nil
	additional, err := executeSingle(ctx, client, additionalAttempt, now)
	if err != nil {
		return primary, err
	}
	if !successful(additional) {
		return additional, nil
	}
	combined := map[string]any{"primary": primary.Root, "additional": additional.Root}
	primary.Root = combined
	primary.Body = combined
	primary.ETag = ""
	return primary, nil
}
