package runtime

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"gopkg.in/yaml.v3"

	defaults "node-herder/configs/integrations"
	"node-herder/internal/integrations/capabilities"
	"node-herder/internal/integrations/capabilities/waste"
	"node-herder/internal/integrations/definitions"
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

// sampleCapability deliberately has no waste metadata, services, dates or timezone. It proves
// runtime orchestration cannot accidentally rely on the shape of the first shipped capability.
type sampleCapability struct{}

type sampleReading struct {
	ContractVersion string    `json:"contract_version"`
	ResourceID      string    `json:"resource_id"`
	Revision        uint64    `json:"revision"`
	Reading         int       `json:"reading"`
	CheckedAt       time.Time `json:"checked_at"`
	EvaluatedAt     time.Time `json:"evaluated_at"`
}

func (sampleCapability) Descriptor() capabilities.Descriptor {
	return capabilities.Descriptor{Capability: "sample_reading/v1", IntegrationType: "sample_information", Name: "Sample readings"}
}
func (sampleCapability) ValidateMapping(mapping map[string]any) error {
	if len(mapping) != 1 || mapping["value"] != "/reading" {
		return model.InvalidDefinition("mapping", "Choose the reading field.")
	}
	return nil
}
func (sampleCapability) Normalize(root any, _ map[string]any, c capabilities.Context) (capabilities.Result, error) {
	var out capabilities.Result
	value, ok := operations.Pointer(root, "/reading")
	if !ok {
		return out, model.Fail("invalid_response", "Missing reading.")
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		return out, err
	}
	reading := sampleReading{ContractVersion: "sample_reading/v1", ResourceID: c.ResourceID, Revision: 1, CheckedAt: c.Now}
	if err = json.Unmarshal(encoded, &reading.Reading); err != nil {
		return out, model.Fail("invalid_response", "Invalid reading.")
	}
	if c.Previous != nil {
		reading.Revision = c.Previous.Revision + 1
	}
	if reading.ResourceID == "" {
		reading.ResourceID, err = c.NewIdentity("resource")
		if err != nil {
			return out, err
		}
	}
	out.Resource, err = model.NewResource(reading)
	out.Reservations = []capabilities.Identity{{ID: reading.ResourceID, Kind: "resource"}}
	return out, err
}
func (sampleCapability) Project(resource model.Resource, now time.Time) (model.Resource, error) {
	var reading sampleReading
	if err := resource.Decode(&reading); err != nil {
		return model.Resource{}, err
	}
	reading.EvaluatedAt = now
	return model.NewResource(reading)
}
func (sampleCapability) Checked(resource model.Resource, now time.Time) (model.Resource, error) {
	var reading sampleReading
	if err := resource.Decode(&reading); err != nil {
		return model.Resource{}, err
	}
	reading.CheckedAt = now
	return model.NewResource(reading)
}

func TestServiceExecutesNonWasteCapability(t *testing.T) {
	conditional := false
	broken := false
	s, clock := serviceFixture(t, providerFunc(func(request *http.Request) (*http.Response, error) {
		switch request.URL.Path {
		case "/v1/addresses":
			return reply(200, `{"candidates":[{"id":"p:synthetic","label":"Sample meter"}]}`), nil
		case "/v1/lookups":
			return reply(201, `{"id":"lookup","status":"done","result":{"property_id":"p:synthetic","reading":42}}`), nil
		default:
			if broken {
				return reply(200, `{"reading":"invalid"}`), nil
			}
			if conditional && request.Header.Get("If-None-Match") != "" {
				return reply(304, ""), nil
			}
			response := reply(200, `{"reading":43}`)
			response.Header.Set("ETag", `"sample"`)
			return response, nil
		}
	}))
	registry, err := capabilities.NewRegistry(waste.Adapter{}, sampleCapability{})
	if err != nil {
		t.Fatal(err)
	}
	s.capabilities = registry
	s.compiler = definitions.NewCompiler(registry)
	compiled, err := s.compiler.CompileYAML(seed(t))
	if err != nil {
		t.Fatal(err)
	}
	d := compiled.Definition
	d.IntegrationType = "sample_information"
	d.ID = "meter"
	d.Name = "Sample meter"
	d.Capability = "sample_reading/v1"
	d.Mapping = map[string]any{"value": "/reading"}
	d.Setup.Search.Expect = nil
	search := d.Inputs[d.Setup.Search.Input]
	search.Type = "text"
	d.Inputs[d.Setup.Search.Input] = search
	source, err := yaml.Marshal(d)
	if err != nil {
		t.Fatal(err)
	}
	key := "sample_information/meter"
	if err = s.files.Write(key, string(source), ""); err != nil {
		t.Fatal(err)
	}
	catalog, err := s.ReloadDefinitions()
	if err != nil || len(catalog.Types) != 2 {
		t.Fatal(catalog, err)
	}
	setup, err := s.CreateSetup(key, "", "")
	if err != nil {
		t.Fatal(err)
	}
	setup, err = s.SearchSetup(setup.ID, " Meter Alpha ", setup.ETag)
	if err != nil {
		t.Fatal(err)
	}
	runQueued(t, s)
	setup, err = s.Setup(setup.ID)
	if err != nil {
		t.Fatal(err)
	}
	if setup.Postcode != "Meter Alpha" {
		t.Fatal("text input was changed", setup)
	}
	setup, err = s.SaveSetup(setup.ID, "p:synthetic", "Sample", setup.ETag)
	if err != nil {
		t.Fatal(err)
	}
	runQueued(t, s)
	instance, err := s.Instance(setup.InstanceID)
	if err != nil || len(instance.Resources) != 1 {
		t.Fatal(instance, err)
	}
	id := instance.Resources[0].ResourceID
	for turn := 0; turn < 2; turn++ {
		conditional = turn == 1
		_, err = s.Refresh(instance.ID, instance.ETag)
		if err != nil {
			t.Fatal(err)
		}
		runQueued(t, s)
		instance, err = s.Instance(instance.ID)
		if err != nil {
			t.Fatal(err)
		}
		var reading sampleReading
		if err = instance.Resources[0].Decode(&reading); err != nil {
			t.Fatal(err)
		}
		if reading.Reading != 43 || reading.ResourceID != id || reading.Revision != 2 || !reading.EvaluatedAt.Equal(clock.Now()) {
			t.Fatal(reading)
		}
	}
	// Offline and live tests use the same codec while leaving the saved instance untouched.
	var lastPreview string
	for _, mode := range []string{"offline", "live"} {
		preview, err := s.TestDefinition(key, string(source), instance.ID, mode, json.RawMessage(`{"reading":99}`), nil)
		if err != nil {
			t.Fatal(mode, err)
		}
		if mode == "live" {
			runQueued(t, s)
			preview, err = s.Setup(preview.ID)
			if err != nil {
				t.Fatal(err)
			}
		}
		lastPreview = preview.ID
		if preview.Preview == nil || preview.Preview.ContractVersion != "sample_reading/v1" {
			t.Fatal(mode, preview)
		}
	}
	after, err := s.Instance(instance.ID)
	if err != nil || after.Resources[0].Revision != 2 {
		t.Fatal(after, err)
	}
	// A failed capability result must retain the last-good content and its resource identity.
	broken = true
	if _, err = s.Refresh(instance.ID, after.ETag); err != nil {
		t.Fatal(err)
	}
	runQueued(t, s)
	after, err = s.Instance(instance.ID)
	if err != nil || after.Error == nil || after.Error.Code != "invalid_response" || after.Resources[0].ResourceID != id || after.Resources[0].Revision != 2 {
		t.Fatal(after, err)
	}
	// Removing a capability must expose an error, never pretend its saved data is empty/ready.
	s.capabilities = fixtureRegistry(t)
	after, err = s.Instance(instance.ID)
	if err != nil || after.Status != "unsupported_capability" || after.Error == nil {
		t.Fatal(after, err)
	}
	preview, err := s.Setup(lastPreview)
	if err != nil || preview.Status != "error" || preview.Preview != nil || preview.Error == nil || preview.Error.Code != "unsupported_capability" {
		t.Fatal(preview, err)
	}
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
	registry := fixtureRegistry(t)
	s := &Service{shipped: defaults.Files, capabilities: registry, compiler: definitions.NewCompiler(registry), repo: repo, files: files, http: client, clock: clock, ctx: ctx, cancel: cancel, queue: make(chan string, 64), running: map[string]context.CancelFunc{}, providerDeadlines: map[string]time.Time{}}
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
	if len(i.Resources) != 1 || wasteSchedule(t, i.Resources[0]).Services[0].Name != "Paper" {
		t.Fatal(i)
	}
	id := wasteSchedule(t, i.Resources[0]).Services[0].ServiceID
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
	if e != nil || wasteSchedule(t, i.Resources[0]).Services[0].ServiceID != id {
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

func TestScheduleConsumersReceiveChronologicalServices(t *testing.T) {
	var providerSchedule map[string]any
	if err := json.Unmarshal([]byte(fixtureSchedule()), &providerSchedule); err != nil {
		t.Fatal(err)
	}
	providerSchedule["collections"] = []map[string]any{
		{"name": "Blue bin", "waste_type": "paper_card", "dates": []string{"2026-11-17"}},
		{"name": "Black bin", "waste_type": "refuse", "dates": []string{"2026-10-27"}},
		{"name": "Brown bin", "waste_type": "glass", "dates": []string{"2026-11-03"}},
	}
	response, err := json.Marshal(map[string]any{"id": "lookup", "status": "done", "result": providerSchedule})
	if err != nil {
		t.Fatal(err)
	}
	service, _ := serviceFixture(t, providerFunc(func(request *http.Request) (*http.Response, error) {
		if request.URL.Path == "/v1/lookups" {
			return reply(201, string(response)), nil
		}
		return standardProvider(request)
	}))
	saved := configuredFixture(t, service)
	listed, err := service.Instances()
	if err != nil {
		t.Fatal(err)
	}
	current, err := service.Instance(saved.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, instance := range []InstanceDTO{saved, current, listed[0]} {
		var names []string
		for _, collection := range wasteSchedule(t, instance.Resources[0]).Services {
			names = append(names, collection.Name)
		}
		if !reflect.DeepEqual(names, []string{"Black bin", "Brown bin", "Blue bin"}) {
			t.Fatal(names)
		}
		if instance.Resources[0].ResourceID != saved.Resources[0].ResourceID || instance.Resources[0].Revision != saved.Resources[0].Revision {
			t.Fatal("read changed resource identity")
		}
	}
}

func fixtureRegistry(t *testing.T) *capabilities.Registry {
	t.Helper()
	registry, err := capabilities.NewRegistry(waste.Adapter{})
	if err != nil {
		t.Fatal(err)
	}
	return registry
}
func wasteSchedule(t *testing.T, resource model.Resource) model.WasteSchedule {
	t.Helper()
	var schedule model.WasteSchedule
	if err := resource.Decode(&schedule); err != nil {
		t.Fatal(err)
	}
	return schedule
}

func seed(t *testing.T) string {
	t.Helper()
	b, err := defaults.Files.ReadFile("waste_collection/whenisbins.yaml")
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
