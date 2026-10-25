package operations

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
func origin(u *url.URL) string {
	port := u.Port()
	host := strings.ToLower(u.Hostname())
	if strings.Contains(host, ":") {
		host = "[" + host + "]"
	}
	if port != "" && port != "443" {
		host += ":" + port
	}
	return strings.ToLower(u.Scheme) + "://" + host
}

// Build constructs a request once; persist its returned body and key before any provider I/O.
func Build(ctx context.Context, a Attempt) (*http.Request, []byte, error) {
	q, ok := a.Definition.Requests[a.RequestID]
	if !ok {
		return nil, nil, model.Fail("reconfigure_needed", "The configured request no longer exists.")
	}
	u, e := url.Parse(a.Definition.BaseURL)
	if e != nil || u.Scheme != "https" || u.Hostname() == "" || u.User != nil || u.RawQuery != "" || u.Fragment != "" || u.Port() != "" && u.Port() != "443" {
		return nil, nil, model.Fail("destination_blocked", "Use a public HTTPS provider on port 443.")
	}
	if origin(u) != a.ApprovedOrigin {
		return nil, nil, model.Fail("reconnect_required", "Review the provider destination before sending saved inputs.")
	}
	path := strings.TrimRight(u.Path, "/")
	escaped := strings.TrimRight(u.EscapedPath(), "/")
	for _, v := range q.Path {
		raw, e := resolve(v, a.Inputs, a.State)
		if e != nil {
			return nil, nil, e
		}
		s, e := scalar(raw)
		if e != nil {
			return nil, nil, e
		}
		if s == "" || s == "." || s == ".." {
			return nil, nil, model.Fail("invalid_input", "A request path value is missing or invalid.")
		}
		path += "/" + s
		escaped += "/" + url.PathEscape(s)
	}
	u.Path = path
	u.RawPath = escaped
	query := url.Values{}
	for name, v := range q.Query {
		raw, e := resolve(v, a.Inputs, a.State)
		if e != nil {
			return nil, nil, e
		}
		s, e := scalar(raw)
		if e != nil {
			return nil, nil, e
		}
		query.Set(name, s)
	}
	u.RawQuery = query.Encode()
	body := a.Body
	if body == nil && q.Method == http.MethodPost {
		values := map[string]any{}
		for name, v := range q.Body {
			values[name], e = resolve(v, a.Inputs, a.State)
			if e != nil {
				return nil, nil, e
			}
		}
		body, e = json.Marshal(values)
		if e != nil {
			return nil, nil, model.Fail("invalid_input", "Invalid request body.")
		}
	}
	if len(body) > 64<<10 {
		return nil, nil, model.Fail("invalid_input", "Request body is too large.")
	}
	request, e := http.NewRequestWithContext(ctx, q.Method, u.String(), bytes.NewReader(body))
	if e != nil {
		return nil, nil, model.Fail("invalid_input", "Could not construct the provider request.")
	}
	request.Header.Set("Accept", "application/json")
	if q.Method == "POST" {
		request.Header.Set("Content-Type", "application/json")
	}
	if q.Idempotency {
		if a.IdempotencyKey == "" {
			return nil, nil, model.Fail("invalid_operation", "An idempotency key is required.")
		}
		request.Header.Set("Idempotency-Key", a.IdempotencyKey)
	}
	if q.Conditional && a.ETag != "" {
		request.Header.Set("If-None-Match", a.ETag)
	}
	mode := a.Definition.Auth.Mode
	if mode == "bearer" && a.Credential == nil {
		return nil, nil, model.Fail("credential_required", "Add the provider credential in Setup.")
	}
	if mode != "none" && a.Credential != nil {
		if a.Credential.Origin != a.ApprovedOrigin {
			return nil, nil, model.Fail("reconnect_required", "The credential belongs to another provider destination.")
		}
		if a.Credential.Value == "" || len(a.Credential.Value) > 8192 || strings.ContainsAny(a.Credential.Value, "\r\n") {
			return nil, nil, model.Fail("invalid_credential", "Enter a valid provider token.")
		}
		request.Header.Set("Authorization", "Bearer "+a.Credential.Value)
	}
	return request, body, nil
}

// Execute retains status/selected headers independently from JSON. No automatic retry or polling.
func Execute(ctx context.Context, client Doer, a Attempt, now time.Time) (Response, error) {
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
	b, e := io.ReadAll(io.LimitReader(response.Body, MaxResponseBytes+1))
	if e != nil || len(b) > MaxResponseBytes {
		return out, model.Fail("invalid_response", "The provider response exceeded its size or time limit.")
	}
	if !utf8.Valid(b) {
		return out, model.Fail("invalid_response", "The provider response is not valid UTF-8.")
	}
	if len(b) == 0 {
		return out, model.Fail("invalid_response", "The provider response was empty.")
	}
	decoder := json.NewDecoder(bytes.NewReader(b))
	decoder.UseNumber()
	if e = decoder.Decode(&out.Body); e != nil {
		return out, model.Fail("invalid_response", "The provider did not return valid JSON.")
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		return out, model.Fail("invalid_response", "The provider returned more than one JSON value.")
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
		valid := false
		switch c.Type {
		case "string":
			s, ok := v.(string)
			valid = ok && len(s) > 0 && len(s) <= 2048
			if valid && c.StripPrefix != "" {
				if !strings.HasPrefix(s, c.StripPrefix) {
					return nil, model.Fail("invalid_response", "A provider identifier has an unexpected format.")
				}
				v = strings.TrimPrefix(s, c.StripPrefix)
				valid = v != ""
			}
		case "boolean":
			_, valid = v.(bool)
		case "number":
			n, ok := v.(json.Number)
			valid = ok && len(n.String()) <= 128
		case "local-date":
			s, ok := v.(string)
			if ok {
				t, e := time.Parse("2006-01-02", s)
				valid = e == nil && t.Format("2006-01-02") == s
			}
		case "utc-timestamp":
			s, ok := v.(string)
			if ok {
				t, e := time.Parse(time.RFC3339Nano, s)
				valid = e == nil
				if valid {
					v = t.UTC().Format(time.RFC3339Nano)
				}
			}
		}
		if !valid {
			return nil, model.Fail("invalid_response", fmt.Sprintf("The provider returned an invalid saved value (%s).", name))
		}
		next[name] = v
	}
	return next, nil
}
