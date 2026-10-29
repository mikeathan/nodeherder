package operations

// Builds provider requests, decodes bounded responses and captures declared state.

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	model "node-herder/models/integrations"
)

// MaxResponseBytes bounds the decoded provider body, including decompressed responses.
const MaxResponseBytes = 2 << 20

// Doer permits isolated HTTP fixtures; production uses the policy-bound client.
type Doer interface {
	Do(*http.Request) (*http.Response, error)
}

// Attempt supplies a pinned definition, private input/state and approved credential context.
// Input/state references resolve saved values, never executable templates. The caller owns
// durable checkpoints and retries; Execute never silently repeats a provider submission.
// Body is nil before first construction; a persisted non-nil body must be replayed unchanged.
type Attempt struct {
	Definition           model.Definition
	RequestID            string
	Inputs               map[string]string
	State                map[string]any
	Credential           *model.Credential
	ApprovedOrigin       string
	Body                 []byte
	IdempotencyKey, ETag string
}

// Response retains status and selected headers separately from parsed JSON. Body is the
// original envelope, Root is the configured result object; 304 legitimately has no JSON body.
type Response struct {
	Status  int
	Body    any
	Root    any
	ETag    string
	RetryAt *time.Time
}

// resolve reads a declared literal or a saved input/state reference without evaluating templates.
func resolve(v model.Value, inputs map[string]string, state map[string]any) (any, error) {
	if v.Ref == "" {
		return v.Literal, nil
	}
	parts := strings.SplitN(v.Ref, ".", 2)
	if len(parts) != 2 {
		return nil, model.Fail("reconfigure_needed", "A request value is invalid.")
	}
	var value any
	var ok bool
	switch parts[0] {
	case "input":
		value, ok = inputs[parts[1]]
	case "state":
		value, ok = state[parts[1]]
	}
	if !ok || value == nil {
		return nil, model.Fail("reconfigure_needed", "A saved request value is missing. Open Setup to reconnect.")
	}
	return value, nil
}

// scalar converts supported request values to bounded text for URLs and headers.
func scalar(v any) (string, error) {
	switch x := v.(type) {
	case string:
		if len(x) > 2048 {
			return "", model.Fail("invalid_input", "A request value is too long.")
		}
		return x, nil
	case bool:
		return strconv.FormatBool(x), nil
	case json.Number:
		if len(x.String()) > 128 {
			return "", model.Fail("invalid_input", "A numeric request value is too long.")
		}
		return x.String(), nil
	case int:
		return strconv.Itoa(x), nil
	case int64:
		return strconv.FormatInt(x, 10), nil
	case float64:
		return strconv.FormatFloat(x, 'g', -1, 64), nil
	case nil:
		return "", nil
	default:
		return "", model.Fail("invalid_input", "Expected a simple request value.")
	}
}

// Build constructs a request without changing the pinned attempt or definition. The
// returned body is persisted by the caller before I/O so POST retries retain exact bytes.
func Build(ctx context.Context, attempt Attempt) (*http.Request, []byte, error) {
	definition, ok := attempt.Definition.Requests[attempt.RequestID]
	if !ok {
		return nil, nil, model.Fail("reconfigure_needed", "The configured request no longer exists.")
	}
	destination, forwardCredentials, err := requestDestination(attempt)
	if err != nil {
		return nil, nil, err
	}
	if err = configureRequestURL(destination, definition, attempt); err != nil {
		return nil, nil, err
	}
	body, err := requestBody(definition, attempt)
	if err != nil {
		return nil, nil, err
	}
	request, err := http.NewRequestWithContext(ctx, definition.Method, destination.String(), bytes.NewReader(body))
	if err != nil {
		return nil, nil, model.Fail("invalid_input", "Could not construct the provider request.")
	}
	if err = setRequestHeaders(request, definition, attempt); err != nil {
		return nil, nil, err
	}
	if err = authenticateRequest(request, attempt, forwardCredentials); err != nil {
		return nil, nil, err
	}
	return request, body, nil
}

// configureRequestURL mutates only the newly parsed destination owned by Build.
func configureRequestURL(destination *url.URL, definition model.Request, attempt Attempt) error {
	path := strings.TrimRight(destination.Path, "/")
	escaped := strings.TrimRight(destination.EscapedPath(), "/")
	for _, value := range definition.Path {
		segment, err := resolveScalar(value, attempt)
		if err != nil {
			return err
		}
		if segment == "" || segment == "." || segment == ".." {
			return model.Fail("invalid_input", "A request path value is missing or invalid.")
		}
		path += "/" + segment
		escaped += "/" + url.PathEscape(segment)
	}
	query := url.Values{}
	for name, value := range definition.Query {
		resolved, err := resolveScalar(value, attempt)
		if err != nil {
			return err
		}
		query.Set(name, resolved)
	}
	destination.Path = path
	destination.RawPath = escaped
	destination.RawQuery = query.Encode()
	return nil
}

// resolveScalar resolves a request value and checks that it is a supported scalar.
func resolveScalar(value model.Value, attempt Attempt) (string, error) {
	resolved, err := resolve(value, attempt.Inputs, attempt.State)
	if err != nil {
		return "", err
	}
	return scalar(resolved)
}

// requestBody constructs a bounded POST body or reuses the exact persisted retry body.
func requestBody(definition model.Request, attempt Attempt) ([]byte, error) {
	body := attempt.Body
	if body == nil && definition.Method == http.MethodPost {
		values := make(map[string]any, len(definition.Body))
		for name, value := range definition.Body {
			resolved, err := resolve(value, attempt.Inputs, attempt.State)
			if err != nil {
				return nil, err
			}
			values[name] = resolved
		}
		encoded, err := json.Marshal(values)
		if err != nil {
			return nil, model.Fail("invalid_input", "Invalid request body.")
		}
		body = encoded
	}
	if len(body) > 64<<10 {
		return nil, model.Fail("invalid_input", "Request body is too large.")
	}
	return body, nil
}

// setRequestHeaders applies content, idempotency and conditional request headers.
func setRequestHeaders(request *http.Request, definition model.Request, attempt Attempt) error {
	request.Header.Set("Accept", "application/json")
	if definition.Method == http.MethodPost {
		request.Header.Set("Content-Type", "application/json")
	}
	if definition.Idempotency {
		if attempt.IdempotencyKey == "" {
			return model.Fail("invalid_operation", "An idempotency key is required.")
		}
		request.Header.Set("Idempotency-Key", attempt.IdempotencyKey)
	}
	if definition.Conditional && attempt.ETag != "" {
		request.Header.Set("If-None-Match", attempt.ETag)
	}
	return nil
}

// Execute performs one admitted request and its optional validated companion. The caller
// owns publication, failure persistence and retry policy; no request is retried here.
func Execute(ctx context.Context, client Doer, attempt Attempt, now time.Time) (Response, error) {
	if err := ValidateCombination(attempt.Definition, attempt.RequestID); err != nil {
		return Response{}, err
	}
	primary, err := executeSingle(ctx, client, attempt, now)
	if err != nil {
		return primary, err
	}
	request := attempt.Definition.Requests[attempt.RequestID]
	if request.Combine == "" || !successful(primary) {
		return primary, nil
	}
	return combineResponse(ctx, client, attempt, primary, now)
}

// executeSingle sends one request and retains status and headers separately from JSON.
func executeSingle(ctx context.Context, client Doer, a Attempt, now time.Time) (Response, error) {
	var out Response
	request, _, e := Build(ctx, a)
	if e != nil {
		return out, e
	}
	response, e := client.Do(request)
	if e != nil {
		return out, model.Fail("provider_unavailable", "The provider could not be reached. Try again later.")
	}
	if response == nil || response.Body == nil {
		return out, model.Fail("invalid_response", "The provider returned no response.")
	}
	defer response.Body.Close()
	out.Status = response.StatusCode
	out.ETag = response.Header.Get("ETag")
	if len(out.ETag) > 512 {
		out.ETag = ""
	}
	out.RetryAt = RetryAfter(response.Header.Get("Retry-After"), now)
	if out.Status >= 300 && out.Status < 400 && out.Status != 304 {
		return out, model.Fail("provider_moved", "The provider returned a redirect. Review its address in Configure operations.")
	}
	if out.Status == 304 {
		return out, nil
	}
	out.Body, e = decodeProviderJSON(response.Body)
	if e != nil {
		return out, e
	}
	out.Root = out.Body
	if out.Status >= 200 && out.Status < 300 {
		p := a.Definition.Requests[a.RequestID].ResponseRoot
		if p != "" {
			out.Root, _ = Pointer(out.Body, p)
		}
	}
	return out, nil
}

// decodeProviderJSON enforces body size, UTF-8 and single-value JSON constraints.
func decodeProviderJSON(body io.Reader) (any, error) {
	b, e := io.ReadAll(io.LimitReader(body, MaxResponseBytes+1))
	if e != nil || len(b) > MaxResponseBytes {
		return nil, model.Fail("invalid_response", "The provider response exceeded its size or time limit.")
	}
	if !utf8.Valid(b) {
		return nil, model.Fail("invalid_response", "The provider response is not valid UTF-8.")
	}
	if len(b) == 0 {
		return nil, model.Fail("invalid_response", "The provider response was empty.")
	}
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	var decoded any
	if e = decoder.Decode(&decoded); e != nil {
		return nil, model.Fail("invalid_response", "The provider did not return valid JSON.")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return nil, model.Fail("invalid_response", "The provider returned more than one JSON value.")
	}
	return decoded, nil
}

// RetryAfter interprets either Retry-After seconds or an HTTP-date relative to now.
// Malformed values provide no deadline; common admission/backoff policy remains the caller's job.
func RetryAfter(raw string, now time.Time) *time.Time {
	if raw == "" {
		return nil
	}
	if seconds, e := strconv.ParseInt(raw, 10, 64); e == nil && seconds >= 0 {
		if seconds > 7*86400 {
			seconds = 7 * 86400
		}
		t := now.Add(time.Duration(seconds) * time.Second)
		return &t
	}
	if t, e := http.ParseTime(raw); e == nil && t.After(now) {
		if t.After(now.Add(7 * 24 * time.Hour)) {
			t = now.Add(7 * 24 * time.Hour)
		}
		return &t
	}
	return nil
}

// Capture applies validated selectors only at their configured completion phase.
func Capture(body any, request model.Request, state map[string]any, complete bool) (map[string]any, error) {
	next := map[string]any{}
	for k, v := range state {
		next[k] = v
	}
	for name, c := range request.Captures {
		if c.When == "complete" && !complete {
			continue
		}
		v, ok := Pointer(body, c.Pointer)
		if !ok || v == nil {
			if c.Optional {
				continue
			}
			return nil, model.Fail("invalid_response", "The provider omitted a required saved value.")
		}
		v, valid, err := normalizeCapture(v, c)
		if err != nil {
			return nil, err
		}
		if !valid {
			return nil, model.Fail("invalid_response", fmt.Sprintf("The provider returned an invalid saved value (%s).", name))
		}
		next[name] = v
	}
	return next, nil
}

// normalizeCapture validates and normalizes one declared value without changing saved state.
func normalizeCapture(value any, capture model.Capture) (any, bool, error) {
	switch capture.Type {
	case "string":
		return captureString(value, capture.StripPrefix)
	case "boolean":
		_, valid := value.(bool)
		return value, valid, nil
	case "number":
		number, valid := value.(json.Number)
		return value, valid && len(number.String()) <= 128, nil
	case "local-date":
		text, valid := value.(string)
		if !valid {
			return value, false, nil
		}
		date, err := time.Parse("2006-01-02", text)
		return value, err == nil && date.Format("2006-01-02") == text, nil
	case "utc-timestamp":
		text, valid := value.(string)
		if !valid {
			return value, false, nil
		}
		instant, err := time.Parse(time.RFC3339Nano, text)
		if err != nil {
			return value, false, nil
		}
		return instant.UTC().Format(time.RFC3339Nano), true, nil
	default:
		return value, false, nil
	}
}

// captureString validates a bounded text value and removes its declared identifier prefix.
func captureString(value any, prefix string) (any, bool, error) {
	text, valid := value.(string)
	if !valid || len(text) == 0 || len(text) > 2048 {
		return value, false, nil
	}
	if prefix == "" {
		return value, true, nil
	}
	if !strings.HasPrefix(text, prefix) {
		return nil, false, model.Fail("invalid_response", "A provider identifier has an unexpected format.")
	}
	text = strings.TrimPrefix(text, prefix)
	return text, text != "", nil
}
