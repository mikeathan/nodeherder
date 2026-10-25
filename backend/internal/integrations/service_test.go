package integrations

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"node-herder/internal/integrations/operations"
	"node-herder/mocks"
	model "node-herder/models/integrations"
	"node-herder/repository"
)

type providerFunc func(*http.Request) (*http.Response, error)

func (f providerFunc) Do(r *http.Request) (*http.Response, error) { return f(r) }
func reply(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Header: http.Header{}, Body: io.NopCloser(bytes.NewBufferString(body))}
}
func fixtureSchedule() string {
	return `{"property_id":"p:synthetic","address_match":"exact","evidence_granularity":"property","date_confidence":"published_calendar","date_completeness":"limited_horizon","retrieved_at":"2026-10-23T12:00:00Z","refresh_due_at":"2026-10-24T12:00:00Z","collections":[{"name":"Paper","waste_type":"paper_card","dates":["2026-10-27"],"dates_complete":true}]}`
}

// Domain tests drive queued work synchronously. Lifecycle tests separately exercise owned workers.
func serviceFixture(t *testing.T, client operations.Doer) (*Service, *mocks.MockClock) {
	t.Helper()
	repo, e := repository.OpenIntegrationRepository(t.TempDir(), repository.IntegrationKeyring{ActiveKeyID: "test", Keys: map[string]string{"test": base64.StdEncoding.EncodeToString(bytes.Repeat([]byte{7}, 32))}})
	if e != nil {
		t.Fatal(e)
	}
	files, e := repository.OpenIntegrationDefinitions(t.TempDir())
	if e != nil {
		t.Fatal(e)
	}
	clock := mocks.NewMockClock(func() time.Time { return time.Date(2026, 10, 23, 13, 0, 0, 0, time.UTC) })
	ctx, cancel := context.WithCancel(context.Background())
	s := &Service{repo: repo, files: files, http: client, clock: clock, ctx: ctx, cancel: cancel, queue: make(chan string, 64), running: map[string]context.CancelFunc{}, providerDeadlines: map[string]time.Time{}}
	if e = s.loadDefinitions(); e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() { cancel(); _ = repo.Close() })
	return s, clock
}
func runQueued(t *testing.T, s *Service) {
	t.Helper()
	select {
	case id := <-s.queue:
		s.execute(id)
	default:
		t.Fatal("no admitted operation")
	}
}
func configuredFixture(t *testing.T, s *Service) InstanceDTO {
	t.Helper()
	p, e := s.CreateSetup("waste_collection/whenisbins", "", "")
	if e != nil {
		t.Fatal(e)
	}
	p, e = s.SearchSetup(p.ID, "XX1 1XX", p.ETag)
	if e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	p, e = s.Setup(p.ID)
	if e != nil {
		t.Fatal(e)
	}
	if len(p.Choices) != 1 {
		t.Fatal(p)
	}
	p, e = s.SaveSetup(p.ID, "p:synthetic", "Bins", p.ETag)
	if e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	i, e := s.Instance(p.InstanceID)
	if e != nil {
		t.Fatal(e)
	}
	return i
}
func standardProvider(r *http.Request) (*http.Response, error) {
	switch r.URL.Path {
	case "/v1/addresses":
		return reply(200, `{"required_input":"property_id","candidates":[{"id":"p:synthetic","label":"Synthetic property"}]}`), nil
	case "/v1/lookups":
		return reply(201, `{"id":"lookup","status":"done","result":`+fixtureSchedule()+`}`), nil
	case "/v1/schedules/synthetic":
		return reply(200, fixtureSchedule()), nil
	default:
		return reply(404, `{"problem":"not_found"}`), nil
	}
}
func TestServiceManualLifecycle(t *testing.T) {
	calls := 0
	s, _ := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) { calls++; return standardProvider(r) }))
	i := configuredFixture(t, s)
	if len(i.Resources) != 1 || i.Resources[0].Services[0].Name != "Paper" {
		t.Fatal(i)
	}
	id := i.Resources[0].Services[0].ServiceID
	if calls != 2 {
		t.Fatal(calls)
	}
	_, _ = s.Catalog()
	_, _ = s.Instance(i.ID)
	_, _ = s.Instances()
	if calls != 2 {
		t.Fatal("read caused provider I/O")
	}
	op, e := s.Refresh(i.ID, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	i, e = s.Instance(i.ID)
	if e != nil || i.Resources[0].Services[0].ServiceID != id {
		t.Fatal(i, e)
	}
	if op, e = s.Operation(op.ID); e != nil || op.Status != "done" {
		t.Fatal(op, e)
	}
	oldETag := i.ETag
	i, e = s.UpdateInstance(i.ID, i.Name, false, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.Refresh(i.ID, i.ETag); e == nil {
		t.Fatal("disabled refresh")
	}
	if _, e = s.UpdateInstance(i.ID, i.Name, true, oldETag); e == nil {
		t.Fatal("lost edit")
	}
	if e = s.DeleteInstance(i.ID, i.ETag); e != nil {
		t.Fatal(e)
	}
	if _, e = s.Instance(i.ID); e == nil {
		t.Fatal("deleted data readable")
	}
}
func TestServicePendingAnd304(t *testing.T) {
	postCount := 0
	conditional := 0
	s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		switch r.URL.Path {
		case "/v1/lookups":
			postCount++
			return reply(201, `{"id":"lookup","status":"running"}`), nil
		case "/v1/lookups/lookup":
			return reply(200, `{"status":"done","result":`+fixtureSchedule()+`}`), nil
		case "/v1/schedules/synthetic":
			if r.Header.Get("If-None-Match") != "" {
				conditional++
				return reply(304, ``), nil
			}
			response := reply(200, fixtureSchedule())
			response.Header.Set("ETag", `"version"`)
			return response, nil
		}
		return standardProvider(r)
	}))
	i := configuredFixture(t, s)
	if len(i.Resources) != 0 {
		t.Fatal("fabricated pending resource")
	}
	if _, e := s.CheckOperation(i.OperationID, i.ETag); e == nil {
		t.Fatal("minimum poll bypass")
	}
	clock.Advance(5 * time.Second)
	if _, e := s.CheckOperation(i.OperationID, i.ETag); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	i, e := s.Instance(i.ID)
	if e != nil || len(i.Resources) != 1 || postCount != 1 {
		t.Fatal(i, e, postCount)
	}
	for n := 0; n < 2; n++ {
		if _, e = s.Refresh(i.ID, i.ETag); e != nil {
			t.Fatal(e)
		}
		runQueued(t, s)
		i, e = s.Instance(i.ID)
		if e != nil {
			t.Fatal(e)
		}
	}
	if conditional != 1 || i.Resources[0].Revision != 2 {
		t.Fatal("304 changed resource revision", conditional, i.Resources[0].Revision)
	}
}
func TestServiceSaveReceiptAndCandidateTampering(t *testing.T) {
	s, _ := serviceFixture(t, providerFunc(standardProvider))
	p, _ := s.CreateSetup("waste_collection/whenisbins", "", "")
	p, _ = s.SearchSetup(p.ID, "XX1 1XX", p.ETag)
	runQueued(t, s)
	p, _ = s.Setup(p.ID)
	if _, e := s.SaveSetup(p.ID, "p:tampered", "Bins", p.ETag); e == nil {
		t.Fatal("tampered choice")
	}
	saved, e := s.SaveSetup(p.ID, "p:synthetic", "Bins", p.ETag)
	if e != nil {
		t.Fatal(e)
	}
	again, e := s.SaveSetup(p.ID, "p:synthetic", "Bins", p.ETag)
	if e != nil || again.InstanceID != saved.InstanceID || again.OperationID != saved.OperationID {
		t.Fatal(again, e)
	}
	var private model.Setup
	e = s.repo.View(func(tx *repository.IntegrationTx) error { _, e := tx.Get("setups", p.ID, &private); return e })
	if e != nil || len(private.Inputs) > 0 || len(private.Credentials) > 0 || len(private.Choices) > 0 {
		t.Fatal("receipt retained setup PII", e)
	}
	runQueued(t, s)
	b, _ := json.Marshal(again)
	if bytes.Contains(b, []byte("p:synthetic")) {
		t.Fatal("receipt exposed provider property ID")
	}
}

func TestServiceCloseJoinsCancelledWork(t *testing.T) {
	started := make(chan struct{})
	s, _ := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		close(started)
		<-r.Context().Done()
		return nil, r.Context().Err()
	}))
	s.wg.Add(1)
	go s.worker()
	p, e := s.CreateSetup("waste_collection/whenisbins", "", "")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.SearchSetup(p.ID, "XX1 1XX", p.ETag); e != nil {
		t.Fatal(e)
	}
	<-started
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if e = s.Close(ctx); e != nil {
		t.Fatal(e)
	}
	if e = s.Close(ctx); e != nil {
		t.Fatal("second close", e)
	}
}

func TestServiceCloseDeadlineKeepsDBOpenUntilWorkerStops(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	s, _ := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) { close(started); <-release; return standardProvider(r) }))
	s.wg.Add(1)
	go s.worker()
	p, e := s.CreateSetup("waste_collection/whenisbins", "", "")
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.SearchSetup(p.ID, "XX1 1XX", p.ETag); e != nil {
		t.Fatal(e)
	}
	<-started
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e = s.Close(ctx); e == nil {
		t.Fatal("stuck worker reported joined")
	}
	if e = s.repo.View(func(*repository.IntegrationTx) error { return nil }); e != nil {
		t.Fatal("DB closed underneath worker", e)
	}
	close(release)
	closeCtx, stop := context.WithTimeout(context.Background(), time.Second)
	defer stop()
	if e = s.Close(closeCtx); e != nil {
		t.Fatal(e)
	}
}

func TestServiceAmbiguousPOSTReusesBodyAndKey(t *testing.T) {
	attempts := 0
	var keys []string
	var bodies []string
	s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == "POST" {
			attempts++
			body, _ := io.ReadAll(r.Body)
			keys = append(keys, r.Header.Get("Idempotency-Key"))
			bodies = append(bodies, string(body))
			if attempts == 1 {
				return nil, context.DeadlineExceeded
			}
		}
		return standardProvider(r)
	}))
	i := configuredFixture(t, s)
	if i.OperationID == "" {
		t.Fatal("lost uncertain attempt")
	}
	clock.Advance(5 * time.Second)
	if _, e := s.CheckOperation(i.OperationID, i.ETag); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	if len(keys) != 2 || keys[0] == "" || keys[0] != keys[1] || bodies[0] != bodies[1] {
		t.Fatal("ambiguous replay changed request", keys, bodies)
	}
}

func TestServiceRenewalOnlyForMatchedProblem(t *testing.T) {
	for _, problem := range []string{"no_schedule", "unknown_endpoint"} {
		t.Run(problem, func(t *testing.T) {
			posts := 0
			s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
				if r.Method == "POST" {
					posts++
				}
				if strings.Contains(r.URL.Path, "/schedules/") {
					return reply(404, `{"problem":"`+problem+`"}`), nil
				}
				return standardProvider(r)
			}))
			i := configuredFixture(t, s)
			clock.Advance(25 * time.Hour)
			if _, e := s.Refresh(i.ID, i.ETag); e != nil {
				t.Fatal(e)
			}
			runQueued(t, s)
			expected := 1
			if problem == "no_schedule" {
				expected = 2
			}
			if posts != expected {
				t.Fatal("unbounded/unmatched renewal", posts)
			}
		})
	}
}

func TestServicePendingDisableEnableKeepsLookup(t *testing.T) {
	posts := 0
	s, clock := serviceFixture(t, providerFunc(func(r *http.Request) (*http.Response, error) {
		if r.Method == "POST" {
			posts++
			return reply(201, `{"id":"lookup","status":"running"}`), nil
		}
		if r.URL.Path == "/v1/lookups/lookup" {
			return reply(200, `{"status":"done","result":`+fixtureSchedule()+`}`), nil
		}
		return standardProvider(r)
	}))
	i := configuredFixture(t, s)
	operationID := i.OperationID
	var e error
	i, e = s.UpdateInstance(i.ID, i.Name, false, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	if _, e = s.CheckOperation(operationID, i.ETag); e == nil {
		t.Fatal("disabled continuation")
	}
	i, e = s.UpdateInstance(i.ID, i.Name, true, i.ETag)
	if e != nil {
		t.Fatal(e)
	}
	clock.Advance(5 * time.Second)
	if _, e = s.CheckOperation(operationID, i.ETag); e != nil {
		t.Fatal(e)
	}
	runQueued(t, s)
	i, e = s.Instance(i.ID)
	if e != nil || len(i.Resources) != 1 || posts != 1 {
		t.Fatal(i, e, posts)
	}
}
