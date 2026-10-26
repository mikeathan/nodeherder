# General-purpose add-ons: investigation and recommendation

Status: revised investigation; no application code changed
Date: 2026-10-23
Repository baseline: `986b923c`
Scope: Go backend, browser configuration surface, add-on definitions and the bins/WhenIsBins example

This report does not claim approval, ratification, or implementation authorization. Behaviour
changes still require a spec, plan, and tasks under the project SDD workflow.

## 0. What changed after architecture review

The first draft proposed projecting add-on results into virtual devices to reuse the existing
device pipeline. That was a technical shortcut, not a correct domain model. A bin collection,
weather reading or energy price is **not a device**. Treating it as one would:

- give add-on data a device identity and availability lifecycle it does not have;
- put add-on data into the public device context and public MCP surface;
- force arbitrary data into device entity types and automation semantics;
- make add-on config, provider IDs and addresses look like device metadata;
- make provider/definition changes look like device changes.

The revised recommendation is:

- **Add-ons are a first-class feature**, with their own definitions, instances, variables,
  secrets, data, status, API, UI and storage.
- **Provider behaviour is declarative data**, not a Go adapter per provider. Default definitions
  ship with the application, and an operator can edit/import a definition when an endpoint,
  header, request shape or response mapping changes. No backend release is required for those
  changes.
- **The runtime is a small, bounded HTTP/JSON execution engine** for definitions. It is not a
  general scripting language. Code changes are only needed for new primitives (a new auth flow,
  a new transform function, a new transport), not for ordinary endpoint/schema changes.
- **Manual setup and manual refresh come first.** Scheduled refresh is a later phase that reuses
  the same execution path.
- **Automations, assistant/MCP and dashboards consume add-on resources through add-on-specific
  interfaces**, not by pretending they are devices. This is a later integration phase if the
  first manual add-on is useful.

The rest of this report reflects that architecture.

## 1. Current codebase architecture and reusable components

### 1.1 Startup and composition

`backend/main.go` loads environment configuration, creates the store, then calls `hub.Register`
(`backend/main.go:24-45`). `hub.Register` constructs the WebSocket hub, metrics query service,
MCP server, MQTT client, controller and HTTP router (`backend/internal/hub.go:83-112`). The
router is a custom router with protected and public groups (`backend/internal/api/routes.go:41-230`).

The store is composed in `backend/store/initialiser.go:11-37`:

- devices: in-memory repository;
- metrics: BoltDB `metrics.db`;
- assistant history: BoltDB `assistant.db`;
- settings: BoltDB `settings.db`;
- settings cache with background cleanup/logging tasks.

This composition is a useful pattern for a new add-on repository, but add-ons should not be added
to `AppConfig` or to the device repository.

### 1.2 What can be reused

| Component | Reuse for add-ons? | Notes |
| --- | --- | --- |
| Custom HTTP router and auth middleware | Yes | Protected add-on routes; no public add-on execution endpoints |
| `lanes.Executor` (`backend/internal/lanes/lanes.go`) | Yes | Generic keyed FIFO executor; ideal for serialising manual refresh per add-on instance |
| `utils.Clock` / `Timer` (`backend/utils/time.go`) | Yes | Testable time, retries, later scheduler |
| `utils/storage.KeyValueDatabase` (`backend/utils/storage/key-value.go`) | Yes | BoltDB buckets for definitions, instances, data, secrets, setup sessions |
| WebSocket event hub (`backend/internal/ws/eventhub.go`) | Yes | Add `addons`/`addonUpdated` events |
| Frontend router/settings/UI controls/contracts | Yes | New Add-ons page, wizard, definition editor, store module |
| Automation type registries (`internal/automations/*`) | Pattern only | Add-on triggers/conditions should be first-class, not converted to devices |
| Metric storage/query | Optional later | Do not force add-on data through device metrics; add a separate time-series namespace if needed |
| MCP server/tool registration | Pattern only | Add `list_addons`/`query_addon`, not device projection |

### 1.3 Why the device pipeline is not the add-on model

The existing device pipeline is excellent for MQTT/Zigbee2MQTT devices:

- ordered per-device lanes in `HubController`;
- `DeviceProcessor` / `DeviceLifetimeService` state, availability, automation and metrics;
- `Device` / `Entity` types with numeric, binary and enum values;
- device snapshots broadcast to the browser.

It is not a generic data model. `createExpose` for non-bridge devices is numeric-only and
whitelisted (`backend/models/devices/device.go:470-487`); device identity is derived from the
friendly name (`backend/store/store.go:150-170`); and the device repository is in-memory
(`backend/store/initialiser.go:13`). More importantly, putting add-on data into devices would
couple add-on lifecycle and exposure to device semantics.

The device pipeline remains a useful reference for concurrency, lifecycle and cleanup, but it is
not reused as the storage or consumption model.

### 1.4 Existing automation and MCP boundaries

Automations are device-centric. `AutomationEngine.HandleDevice` loads a recipe by device ID and
evaluates triggers against the device (`backend/internal/automations/engine.go:88-95`). The
scheduler is an automation enable/disable scheduler, not a general refresh scheduler
(`backend/internal/automations/scheduler.go:174-294`).

MCP has one tool, `query_device`, which reads device metrics (`backend/internal/mcp/tools/intent.go:54-63`).
The public routes are `/api/context/devices` and `/api/mcp` (`backend/internal/hub.go:64-74`).

Therefore:

- add-on data must have its own read API and MCP tool/resource;
- automation support needs new add-on trigger/condition types (or a generic signal trigger);
- scheduled refresh needs a new scheduler, not the automation scheduler;
- add-on data must not be inserted into the public device context.

### 1.5 Persistence, configuration, secrets

- `AppConfig`/`HubConfig` (`backend/models/settings/config.go:150-202`) is hub settings, cloned and
  returned by protected APIs. It is not a place for endpoint definitions, provider variables or
  credentials.
- `FileSettingsRepo` stores one JSON blob in BoltDB (`backend/repository/settings.go:12-57`).
- Automations use JSON files and a type-registry serializer
  (`backend/internal/automations/automation.go:150-171`, `backend/utils/storage/serialiser.go`).
- There is no dedicated encrypted secret store. Existing secrets are environment variables
  (`backend/utils/enviroment.go:32-65`, `:103-120`).
- There is no general outbound HTTP abstraction. The assistant proxy and Google OAuth each create
  their own `http.Client` (`backend/internal/api/routes.go:636-731`,
  `backend/internal/auth/auth.go:149-175`).

Add-ons need their own repository and secret store.

### 1.6 Frontend extension points

- Settings sections: `frontend/src/mixins/useTabComponents.ts:9-45`,
  `frontend/src/components/hub/settings/Settings.vue:15-45`.
- Routes/auth guards: `frontend/src/router/index.ts:21-116`.
- WebSocket event map: `frontend/src/store/modules/ws/index.ts:49-87`.
- Authenticated HTTP helpers: `frontend/src/contracts/api.ts:1-47`.
- Shared UI controls: `frontend/src/components/ui/*`.

A new Add-ons page can reuse the router, UI controls, Vuex conventions and WS event mapping
without touching device state.

## 2. Problem definition and requirements

### 2.1 Problem

Users want to add a capability to NodeHerder by supplying ordinary household information:

1. Select “Bin Collections”.
2. Enter a postcode.
3. Pick their address from the returned options.
4. Save.
5. See and refresh collection dates manually.

They should not understand API endpoints, HTTP methods, request bodies, response structures or
provider IDs. But the operator must be able to repair a definition when a provider changes its
endpoint or response shape **without waiting for a backend release**.

### 2.2 Functional requirements

| ID | Requirement | Priority |
| --- | --- | --- |
| FR-01 | Add-ons are independent of devices and device state | P1 |
| FR-02 | Catalog of shipped add-on definitions and configured instances | P1 |
| FR-03 | Server-driven multi-step setup with dependent fields | P1 |
| FR-04 | Persist variables from one request and use them in later requests | P1 |
| FR-05 | Declarative provider definition for URL, method, headers, body, extraction and outputs | P1 |
| FR-06 | Edit/import/export/reset definitions in an advanced UI without backend code changes | P1 |
| FR-07 | Manual refresh/poll for async providers; no scheduler required in v1 | P1 |
| FR-08 | Secure credential storage and redaction | P1 |
| FR-09 | Add-on data/status UI and protected read API | P1 |
| FR-10 | Scheduled refresh/resume later, reusing the manual execution path | P2 |
| FR-11 | Automations/MCP/dashboard consumption through add-on-specific interfaces | P2 |
| FR-12 | Optional add-on history/time series | P3 |
| FR-13 | User-authored arbitrary scripts or unrestricted workflows | Out of scope |

### 2.3 Definition requirements implied by the feedback

Because endpoint changes must be repairable in the UI:

- Default definitions are **data**, not Go constants.
- The definition must describe:
  - user inputs and setup steps;
  - request(s), including method, URL template, query, headers and JSON body;
  - extraction from responses into named variables;
  - which variables are persisted for later requests;
  - async polling behaviour;
  - refresh request(s) using stored variables;
  - output fields and their types/labels;
  - auth requirements and where secrets may be used.
- The runtime must validate definitions before saving and against a sample response before use.
- The UI needs two levels:
  - **normal**: only the inputs required by the definition;
  - **advanced**: definition editor/import/export/test for the operator.

### 2.4 Assumptions and open decisions

Assumptions:

- Single-household deployment; authenticated users can use add-ons.
- Manual setup/refresh is the first delivery; scheduling is explicitly deferred.
- Cloud HTTPS providers are in scope; local-network providers need an explicit egress policy.
- UK dates use `Europe/London`.
- Definitions are trusted operator content, but still validated and resource-limited.

Open decisions:

1. Who may edit definitions? Recommendation: an operator/admin capability, not every
   authenticated user. Current auth has no roles (`backend/internal/auth/tokens.go:19-24`).
2. Is a raw JSON definition editor acceptable for v1, with a guided builder later?
3. Should add-on data be available to MCP/automations in v1 or only the protected Add-ons UI/API?
4. Are add-on outputs recorded historically, or only current data?
5. Do we need a generic “check status” action for async lookups, or is bounded inline polling
   enough for the first version?

## 3. Research findings

### 3.1 Home Assistant

Home Assistant separates:

- **integration manifest** — identity, dependencies, discovery, integration type;
- **config flow** — server-driven UI steps with validation, dynamic forms, progress, external
  steps and reauth/reconfigure;
- **config entry** — stored data, options, unique ID, version and runtime data;
- **coordinator** — one polling loop for a data source, with failure/backoff/availability;
- **entities** — stable typed outputs that automations and dashboards consume.

Sources: Home Assistant contributor docs, “Building integrations”, “Config flow”, “Config
entries”, “Fetching data”, “The integration manifest” (accessed 2026-10-23).

**Lesson:** the useful separation is definition → setup flow → instance/variables → data →
outputs. NodeHerder should copy that separation, not Python plugin loading. An add-on output is
like an entity/resource, not a device.

Home Assistant also offers YAML REST/template integrations for long-tail HTTP APIs. Those expose
URLs, request options, JSONPath/templates and auth to the user. They prove that endpoint/schema
changes can be fixed in configuration rather than code, but they are advanced YAML and not a
friendly add-on experience. A structured definition schema with a normal input UI is the better
product.

### 3.2 Airbyte low-code/declarative connectors

Airbyte’s connector builder and low-code CDK describe HTTP APIs through declarative manifests:
requests, auth, pagination, extraction and transformations. The builder is positioned as enough
for most connectors; the low-code CDK handles more complex cases; the full Python CDK handles the
rest.

**Lesson:** a declarative HTTP/JSON runtime is viable and directly addresses “change the endpoint
without a release”. The manifest schema grows around auth variants, pagination, dependencies,
async polling and transformations. Keep the DSL deliberately small; add code hooks only where the
DSL cannot express a provider safely.

### 3.3 Node-RED, n8n, openHAB, ioBroker

These systems demonstrate breadth through workflow editors and adapter ecosystems:

- Node-RED: event-driven low-code flows; very flexible, but a workflow editor rather than a
  curated add-on catalog.
- n8n: visual workflows plus code/npm packages and a large connector catalog; powerful, but a
  separate runtime and deployment surface.
- openHAB: add-ons/things/channels/items; strong stable model, heavyweight OSGi ecosystem.
- ioBroker: adapter processes and admin UI; broad, but process/packaging overhead.

**Lesson:** NodeHerder does not need a general workflow runtime or plugin marketplace for the
first add-ons. A small declarative runtime covers standard HTTP/JSON providers. A workflow engine
would be justified only if users need arbitrary multi-branch automations across many services.

### 3.4 WhenIsBins API

Observed OpenAPI 1.1.0 contract and live checks (2026-10-23):

- `GET /v1/addresses?postcode=...` returns the council, `required_input`, `candidates_source`,
  `candidates[]` (`id`, `label`), `postcode_representative` and optional `input_options`.
- `POST /v1/lookups` is the single write. It requires `postcode` plus the input discovery asked
  for, and an `Idempotency-Key`. `201` means created, not completed.
- Statuses: `queued`, `running`, `partial`, `done`, `failed`; `done`/`failed` are terminal.
  `Retry-After` is supplied while non-final. `GET /v1/lookups/{id}` and
  `GET /v1/lookups/{id}/wait` provide polling/long-polling.
- `GET /v1/schedules/{property_token}` is the stable, address-free schedule resource. It supports
  `If-None-Match`/`ETag`; `304` means unchanged, not re-verified. Stale data is served while a
  background refresh occurs. `404 no_schedule` means no answer held.
- The schedule contains collections, by-date grouping, address match, date confidence,
  completeness, evidence granularity, retrieval/refresh timestamps, optional `calendar_url`,
  `uprn` and `council_id`.
- Free anonymous use has limits; regular use should request a token. 429/503 include
  `Retry-After`.

Live observations:

- address discovery returned the Sheffield council and candidate property IDs;
- the stable schedule endpoint returned the same core schedule data and an `ETag`;
- `If-None-Match` returned `304` with the same ETag;
- the `calendar_url` returned a valid iCalendar feed with a 12-hour publication/refresh hint and
  an explicit limited-coverage note.

**Conclusion:** the first lookup is a one-time setup operation per property. Manual refresh
should use the stable schedule endpoint, not repeat the lookup. Store the property token/UPRN
from the initial result for later requests.

### 3.5 Sheffield council direct endpoint

With a genuine UPRN (`100051030107`), `POST /api/getCalendarData` returned `200` with dated
records, service names and colours. The response also contains the full address in every record.

This is a future provider definition, not the default:

- it needs a UPRN the user normally does not know;
- it is undocumented and council-specific;
- it carries PII in every record;
- it has no documented versioning, retry or rate-limit contract.

### 3.6 iCalendar feed

The feed is simple and valid, but loses:

- address match;
- date confidence;
- date completeness;
- evidence granularity;
- provider refresh metadata;
- collection metadata beyond the summary.

It is useful as an export/subscription URL or an optional fallback, not as the primary add-on
data source.

### 3.7 Go libraries

| Option | Use | Cost/risk | Verdict |
| --- | --- | --- | --- |
| Stdlib `net/http` | Runtime HTTP | Must implement SSRF/retry/limits | Recommended |
| `golang.org/x/time/rate` | Per-definition rate limits | Small dependency | Optional/recommended |
| `tidwall/gjson` or `ohler55/ojg/jp` | JSONPath/GJSON extraction | Untyped mappings; need validation | Recommended for definition extractors |
| `text/template` with restricted funcs | URL/body/header templates | Must prevent secret leakage into URLs/logs | Recommended with constraints |
| `expr-lang/expr` | Safe mapping expressions | Adds a language/validation surface | Optional for advanced transforms |
| `cel-go` | Sandboxed expressions | Additional runtime/language | Defer |
| `goja` / Starlark | Embedded scripting | Sandbox/resource/debug/supply-chain burden; current goja README requires Go 1.26 while this repo uses Go 1.24 | Avoid |
| `arran4/golang-ical` + RRULE | iCalendar parsing | Loses metadata | Optional |
| stdlib `plugin` / `hashicorp/go-plugin` | Dynamic Go plugins | Build/version/process/security burden | Avoid |

The project already uses `golang.org/x/oauth2` (`backend/go.mod`), so OAuth definitions can reuse
it. JSON is already the dominant configuration format; YAML can be supported later if desired.

## 4. Alternative approaches and trade-offs

| Approach | Endpoint changes without code | Extensibility | Security | Complexity | Verdict |
| --- | --- | --- | --- | --- | --- |
| Code-only provider adapters | No | Medium | High | Medium | Rejected given the repair requirement |
| Device projection | No | Low/medium | Lower (public device exposure) | Medium | Rejected: wrong domain |
| Declarative definitions + generic runtime | Yes | High for HTTP/JSON | Medium; must constrain templates/egress | Medium | **Recommended** |
| Declarative definitions + Go hooks | Yes for common changes; code for special auth/transforms | High | Medium/high | Medium | **Recommended evolution** |
| Embedded scripting | Yes | Very high | Low unless heavily sandboxed | High operational | Rejected for v1 |
| Workflow engine (n8n/Node-RED/Temporal) | Yes | Very high | Separate boundary | High infrastructure | Rejected for v1 |
| External Home Assistant add-on bridge | Depends on HA | Limited | Depends on HA | Requires HA | Rejected as core |

### Why declarative now?

The user explicitly needs endpoint/schema changes without a Go release. A code-only adapter does
not satisfy that. A declarative runtime satisfies it while keeping the execution model bounded and
reviewable, provided that:

- definitions are validated data, not arbitrary code;
- only a small set of operations is supported;
- egress and secrets are constrained;
- definition editing is restricted to an operator/admin path;
- the runtime is covered by fixture and security tests.

### Why not a full workflow/scripting engine?

The bins flow needs two requests, variables, async polling and output mapping. It does not need
loops, arbitrary branching, user code, package installation, or a visual canvas. A small
declarative schema covers the requirement with far less operational and security surface.

## 5. Recommended architecture

### 5.1 Naming

- **Add-on** — the user-facing feature (“Bin Collections”, “Weather”, “Energy Prices”).
- **Add-on definition** — the declarative provider/request/output package shipped by default or
  edited/imported by the operator.
- **Add-on instance** — one configured use of a definition (one property, one location, one
  tariff).
- **Variable** — a value captured from a response and persisted for later requests (e.g.
  `property_id`, `property_token`, `uprn`, `etag`).
- **Setup session** — the temporary server-side state for a multi-step add-on setup.
- **Add-on data** — the latest canonical output document, with provenance and timestamps.
- **Output** — a typed field in add-on data that other features can display/consume.

Do not call these devices.

### 5.2 Components

```text
Browser
  │  protected REST: catalog/setup/instances/refresh/definition editor
  │  WS: addons/addonUpdated
  ▼
Add-on API
  │
  ├── Definition registry: shipped defaults + user overrides/imports
  ├── Setup manager: steps, validation, variables, secrets, async progress
  ├── Instance repository: config, variables, definition version, status
  ├── Secret store: encrypted values, refs only in config/data
  ├── Execution engine: HTTP requests, extraction, templates, polling
  ├── Data repository: latest canonical output + run metadata
  └── (later) Scheduler / automation / MCP adapters
```

Suggested paths:

```text
backend/models/addons/{definition,instance,data,setup}.go
backend/repository/addons.go
backend/repository/addon_secrets.go
backend/internal/addons/
  registry.go        // default + overridden definitions
  definition.go      // schema, validation, versioning
  engine.go          // request execution, extraction, templates
  setup.go           // setup sessions and step transitions
  instances.go       // lifecycle, manual refresh, polling
  data.go            // canonical data/provenance
  secrets.go
  http.go            // bounded outbound client/egress policy
backend/internal/api/addons.go
frontend/src/components/addons/...
frontend/src/store/modules/addons/...
frontend/src/contracts/addons.ts
frontend/src/types/addons.type.ts
```

### 5.3 Add-on definition schema

The exact schema is an implementation decision. It must at least express:

```go
type Definition struct {
    ID          string
    Name        string
    Version     int
    Description string
    Category    string // bins, weather, energy, ...
    Auth        AuthSpec
    Inputs      []InputField
    Setup       []SetupStep
    Refresh     *RequestPlan
    Outputs     []OutputField

    // Optional:
    Poll         *PollPlan
    ScheduleHint *ScheduleHint
}

type SetupStep struct {
    ID          string
    Title       string
    Fields      []Field
    Request     *RequestPlan
    Extract     map[string]string // variable -> JSONPath/GJSON expression
    Store       []string          // variables persisted to the instance
    Next        string            // next step ID
    Complete    bool              // final step
}

type RequestPlan struct {
    Method  string
    URL     string
    Query   map[string]string
    Headers map[string]string
    Body    any
}

type OutputField struct {
    Key         string
    Label       string
    Description string
    Type        string // number, string, boolean, date, datetime, list
    Unit        string
    // Optional presentation hints; no arbitrary HTML/JS.
}
```

The definition must be able to express conditional next steps (for example, if a council returns
`required_input: property`, show a free-text property field instead of a candidate select). Keep
the condition vocabulary small and validated.

### 5.4 Execution semantics

For a setup step:

1. Render the step fields from the definition, using known variables/options.
2. Validate submitted values against the field type and the current setup session.
3. Persist any secret values to the secret store immediately.
4. Render the request using only the approved variable/secret context.
5. Execute with the bounded HTTP client.
6. Extract response values with JSONPath/GJSON into named variables.
7. Merge non-secret variables into the setup session; persist variables marked `Store`.
8. Decide the next step, completion, or a sanitized error.
9. If the definition has a poll plan, either poll within the manual action’s bounded wait or
   persist the poll cursor and show “Check status”.

For a manual refresh:

1. Load the instance and pinned definition version.
2. Build the refresh request from stored variables.
3. Execute and extract.
4. Validate required outputs.
5. Atomically replace the latest data document and run metadata.
6. Broadcast `addonUpdated`.
7. On `304`, update the check timestamp only; keep original retrieval time.
8. On failure, keep last good data, record sanitized status/error, show it in the UI.

No background scheduler is required for v1. Manual async lookup may poll while the user waits for
a bounded period; if still pending, store the lookup ID/cursor and let the user press “Check
status”.

### 5.5 What “endpoint change without code” means

Default definitions are embedded in the backend at build time, but are stored/overridden as data.
The advanced editor can change:

- URL and version/path;
- method, query parameters, headers;
- JSON body templates;
- JSONPath/GJSON extraction paths;
- output labels/types/units;
- auth placement (header/query/body);
- polling endpoints/cursors;
- refresh endpoint/request.

No Go release is needed for these changes. Validate and test the definition before activation.

Code is still needed when:

- a new auth type is required (e.g. a new OAuth variant/signature scheme);
- a new transform primitive is required (e.g. a date rule the definition DSL cannot express);
- a provider requires a non-HTTP transport;
- the definition schema itself needs a new capability.

That is the right boundary: common provider repairs are data changes; genuinely new mechanics are
reviewed code.

### 5.6 Definition versioning and defaults

- Shipped definitions have an `id`, `version`, and content hash.
- An instance pins the definition version it was created/upgraded with.
- The UI shows “default” vs “customized” and lets the operator reset to the shipped version.
- Updating a definition is explicit: validate, test, show a diff, then apply. Instances remain on
  the old version until upgraded, or the operator can apply to all.
- Import/export supports sharing a repaired definition without a release.
- The runtime supports the current and previous definition schema versions; unknown newer
  versions are rejected with a clear error rather than partially executed.

### 5.7 Instance and data model

```go
type Instance struct {
    ID                string
    DefinitionID      string
    DefinitionVersion int
    DisplayName       string
    Config            map[string]any // non-secret user inputs
    Variables         map[string]any // persisted provider IDs/tokens/etag
    SecretRefs        []string
    Enabled           bool
    CreatedAt         time.Time
    UpdatedAt         time.Time
    LastRun           *RunInfo
}

type Data struct {
    InstanceID        string
    DefinitionVersion int
    CapturedAt        time.Time
    RetrievedAt       time.Time
    Values            map[string]any
    Provenance        map[string]any
    ETag              string
    RefreshDueAt      *time.Time
}

type RunInfo struct {
    StartedAt   time.Time
    FinishedAt  time.Time
    Status      string // success, not_modified, pending, failed
    HTTPStatus  int
    ErrorCode   string
    ErrorDetail string // sanitized
}
```

`Config` holds user-supplied non-secret values such as postcode and display preferences.
`Variables` holds captured provider values such as `property_id`, `property_token`, `uprn`,
`lookup_id` and `etag`. If a variable is personal (UPRN, property ID, address, calendar URL),
treat it as sensitive in logs/API; encrypt it if the storage design supports per-field encryption,
or keep the whole instance record private and redacted.

### 5.8 Persistence

Use a dedicated BoltDB file, e.g. `data/addons.db`, with buckets:

- `definitions` — imported/overridden definitions and hashes;
- `instances` — config, variables and definition version;
- `data` — latest canonical data and run metadata per instance;
- `setup_sessions` — temporary setup state, owner, expiry, cursor/idempotency key;
- `secrets` — encrypted secret values by ref;
- optional `runs` — bounded recent run history for diagnostics.

Why separate from `AppConfig`:

- definitions are large, versioned and operator-editable;
- provider variables and PII should not enter general hub settings;
- data is per-instance and can be replaced independently;
- secrets must never be serialized with hub settings or sent to the browser.

### 5.9 API

Protected routes:

```text
GET    /api/addons/catalog
GET    /api/addons/definitions
PUT    /api/addons/definitions/{id}
POST   /api/addons/definitions/import
POST   /api/addons/definitions/{id}/test

POST   /api/addons/setup
GET    /api/addons/setup/{sessionID}
POST   /api/addons/setup/{sessionID}/step
POST   /api/addons/setup/{sessionID}/cancel

GET    /api/addons/instances
GET    /api/addons/instances/{id}
GET    /api/addons/instances/{id}/data
POST   /api/addons/instances/{id}/refresh
POST   /api/addons/instances/{id}/poll
PATCH  /api/addons/instances/{id}
DELETE /api/addons/instances/{id}
```

Responses are sanitized DTOs:

- never return secret values;
- redact/omit provider variables and PII unless the caller is authorized and the field is needed;
- return definition input schemas and options, not the raw provider response;
- return sanitized run errors and last HTTP status.

WebSocket:

```text
addons        // list/summary after connect or change
addonUpdated  // instance status/data metadata changed
```

No device events are emitted for add-ons.

### 5.10 Frontend

Proposed components:

```text
frontend/src/components/addons/
  AddonsPage.vue
  AddonCatalog.vue
  AddonSetupWizard.vue
  AddonDetail.vue
  AddonDataView.vue
  AddonDefinitionEditor.vue // advanced operator view
  AddonHealth.vue
frontend/src/store/modules/addons/
frontend/src/contracts/addons.ts
frontend/src/services/addons.service.ts
frontend/src/types/addons.type.ts
```

Normal setup is fully schema-driven:

- text/postcode/number/date/select/checkbox fields;
- dynamic options from setup state;
- consent fields;
- progress and pending status;
- sanitized errors with retry/back/check-status actions.

The advanced definition editor is operator-facing:

- JSON/YAML editor or structured form;
- schema validation;
- “Test request” with a sample input and redacted preview;
- diff against the shipped default;
- import/export;
- reset to default;
- warning that definitions perform outbound requests.

### 5.11 Consumption by other features

Do not project add-ons into devices. Add explicit adapters when needed:

**Add-on data API** — protected UI/API first.

**MCP/assistant** — add `list_addons` and `query_addon` tools/resources. The assistant can read
add-on outputs and provenance directly. Do not require add-on metrics or device context.

**Automations** — add a first-class add-on event source and trigger/condition types:

```text
AddonTrigger { instance_id, output, conditions }
AddonCondition { instance_id, output, operator, value }
```

The automation engine needs an add-on evaluation path alongside the device path. This should be
scoped as a separate change because it touches engine routing, dependency indexing and the recipe
UI. It is not a reason to call add-ons devices.

**Dashboard/tiles** — add-on cards can render `OutputField`s generically. A capability-specific
summary component is presentation, not setup logic.

**History** — optional later. If numeric add-on outputs need history, add a separate add-on
time-series repository/namespace rather than forcing them through device metrics.

### 5.12 Manual first, scheduled later

V1:

- setup is user-triggered;
- refresh is a button;
- async lookup can be polled inline for a bounded period, then resumed by a “Check status”
  button using the persisted lookup ID/cursor;
- ETag/refresh metadata is stored but not scheduled;
- UI shows last run and data age.

Later:

- add a scheduler that calls the same `Refresh`/`Poll` methods;
- add backoff/jitter, stale marking and automatic resume;
- reuse the same definition/runtime, not a second implementation;
- add scheduler tests with `utils.Clock`.

This directly matches the requested sequence and keeps the first delivery small.

## 6. User experience

### 6.1 Add-ons page

A top-level or Settings entry called **Add-ons** shows:

- catalog cards for shipped definitions: Bin Collections, Weather, Energy Prices;
- configured instances with data preview, last updated, status and refresh action;
- “Add add-on” and (for operators) “Manage definitions”.

### 6.2 Bins setup

1. Select **Bin Collections**.
2. Enter postcode.
3. The definition’s first request runs. The response’s candidate list is stored in the setup
   session and rendered as a select field.
4. Choose the address. The selected candidate ID is validated against the session and stored as
   `property_id`.
5. The second request uses `postcode` + `property_id` and the persisted idempotency key.
6. If the lookup is done, extract `collections`, `property_token`, `uprn`, `refresh_due_at`, etc.
   and save the instance/data.
7. If pending, show progress and a “Check status” action. No background scheduler is required.
8. The instance detail shows next collection, all dates, source/confidence/horizon, last updated
   and a Refresh button.

The user never sees endpoints, JSON paths, HTTP methods, or provider IDs. The operator can open
the advanced definition editor to repair a URL or extraction path.

### 6.3 Instance detail

Show:

- output fields (next collection, bins, days until, etc.);
- full add-on data/list where appropriate;
- source/provenance and any quality labels (address match, date confidence, limited horizon);
- last run, last successful retrieval, data age;
- sanitized error and actions: Retry, Check status, Edit inputs, Edit definition (operator),
  Remove;
- no device card/device lifecycle claims.

### 6.4 Definition editor experience

- Start from a shipped default.
- Show inputs, requests, extractors and outputs in a structured view; raw JSON/YAML available.
- Validate schema and required variables.
- Test against the live provider only after explicit confirmation; show a redacted preview.
- Save as an override; never silently mutate the shipped default.
- Export/import for sharing fixes.
- Reset to shipped version.

This is how an endpoint change is handled without code: edit the definition, test it, apply it.

## 7. Detailed walkthrough: bins variables and requests

### 7.1 First request — postcode discovery

Definition step (illustrative):

```json
{
  "id": "postcode",
  "fields": [
    { "key": "postcode", "type": "postcode", "label": "Postcode", "required": true }
  ],
  "request": {
    "method": "GET",
    "url": "https://whenisbins.com/v1/addresses",
    "query": { "postcode": "{{ postcode }}" }
  },
  "extract": {
    "council": "$.council",
    "required_input": "$.required_input",
    "candidates": "$.candidates",
    "postcode_representative": "$.postcode_representative"
  },
  "next": "address"
}
```

Runtime:

- validates/normalises the postcode;
- executes the request with timeout and response-size limits;
- stores `council`, `required_input`, `candidates` and consent policy in the setup session;
- renders the next step. If `required_input` is `property_id`, show a select of
  `candidates[].label`; if it is `property`, show a text field; if `candidates_source` is
  `unavailable`, show the free-text fallback.

The candidate list is **server-side session state**, not something the browser can invent.

### 7.2 Second request — create lookup

Definition step:

```json
{
  "id": "address",
  "fields": [
    { "key": "property_id", "type": "select", "label": "Address",
      "options": "{{ candidates }}" }
  ],
  "request": {
    "method": "POST",
    "url": "https://whenisbins.com/v1/lookups",
    "headers": {
      "Content-Type": "application/json",
      "Idempotency-Key": "{{ idempotency_key }}"
    },
    "body": {
      "postcode": "{{ postcode }}",
      "property_id": "{{ property_id }}"
    }
  },
  "extract": {
    "lookup_id": "$.id",
    "status": "$.status",
    "property_token": "$.result.property_id",
    "uprn": "$.result.uprn",
    "collections": "$.result.collections",
    "by_date": "$.result.by_date",
    "date_confidence": "$.result.date_confidence",
    "date_completeness": "$.result.date_completeness",
    "retrieved_at": "$.result.retrieved_at",
    "refresh_due_at": "$.result.refresh_due_at"
  },
  "store": ["property_id", "property_token", "uprn", "lookup_id"]
}
```

The runtime:

- verifies the selected `property_id` was one of the session candidates;
- generates and persists the idempotency key before sending;
- maps `201`/`done` to completed setup and saves the extracted values as instance variables;
- maps `queued`/`running`/`partial` to a pending state with `lookup_id`, `Retry-After` and a
  cursor; the user can retry/check status;
- maps `failed` to a sanitized error and allows address reselection where applicable;
- never resubmits a new lookup because the first one is slow.

Stored instance variables after success:

| Variable | Source | Later use |
| --- | --- | --- |
| `postcode` | user input | display/refresh fallback |
| `property_id` | selected candidate | re-lookup if needed |
| `property_token` | `result.property_id` | stable schedule refresh |
| `uprn` | `result.uprn` | diagnostics/direct-provider future |
| `lookup_id` | `id` | manual status check |
| `etag` | first schedule response | conditional refresh |

This is the exact “store the code from the first response and use it in the query” behaviour.

### 7.3 Refresh request

Definition refresh plan (illustrative):

```json
{
  "request": {
    "method": "GET",
    "url": "https://whenisbins.com/v1/schedules/{{ property_token }}",
    "headers": { "If-None-Match": "{{ etag }}" }
  },
  "extract": {
    "collections": "$.collections",
    "by_date": "$.by_date",
    "date_confidence": "$.date_confidence",
    "date_completeness": "$.date_completeness",
    "retrieved_at": "$.retrieved_at",
    "refresh_due_at": "$.refresh_due_at",
    "calendar_url": "$.calendar_url"
  },
  "store": ["etag"]
}
```

Manual behaviour:

- `200` replaces data; update ETag/retrieval/refresh metadata.
- `304` updates “last checked” only; keep `retrieved_at`.
- `404 no_schedule` offers a controlled re-lookup/check action.
- `422` address mismatch requires confirmation/re-setup.
- `429`/`503` show retry-after; no automatic scheduler in v1.
- `401` shows “credentials required” (even if WhenIsBins currently needs none).

### 7.4 Polling plan

The definition schema should include an optional poll block:

```json
{
  "poll": {
    "id": "lookup_id",
    "request": {
      "method": "GET",
      "url": "https://whenisbins.com/v1/lookups/{{ lookup_id }}"
    },
    "until": "$.status in ['done','failed']",
    "retry_after": "$.retry_after",
    "bounded_wait_seconds": 120
  }
}
```

V1 may implement only the bounded inline wait plus manual “Check status”; the later scheduler can
execute the same poll plan automatically. Do not build a separate polling implementation.

### 7.5 Output fields

Suggested bins outputs:

| Key | Type | Meaning |
| --- | --- | --- |
| `next_collection_date` | date/string | next collection date |
| `next_collection_types` | string/list | bin/service names for that date |
| `days_until_collection` | number | relative value for display/automation |
| `collections` | list | full collection list with dates/colours/completeness |
| `source_council` | string | council name/id |
| `date_confidence` | string | provider confidence label |
| `date_completeness` | string | limited/full/next-only label |
| `data_age` | number/seconds | display/diagnostic |
| `calendar_url` | string/link | optional subscription link (protected, not public) |

Output fields are definition-owned and stable. A later definition update can add fields without
changing the add-on runtime.

## 8. Reliability and API evolution

### 8.1 Provider/definition changes

| Change | Handling without Go code |
| --- | --- |
| URL/version/path | Edit definition request URL; validate/test; apply |
| Query/header/body shape | Edit templates; validate against test response |
| Response field rename/move | Edit extractor JSONPath/GJSON; validate required outputs |
| New output field | Add output definition and extractor |
| Poll endpoint/cursor | Edit poll plan |
| Auth placement (same auth kind) | Edit definition auth spec |
| Provider shutdown | Disable/deprecate definition; configure another; preserve data |
| Response date/number format | Use supported transform functions or code if new primitive needed |
| New auth protocol/transport | Code change (runtime primitive) |

The definition editor must show validation errors before save and a redacted test result. There
must be a “reset to default” path and an audit record of definition changes.

### 8.2 Definition/instance versioning

- Definition version and content hash stored with each instance.
- Instance data records the definition version used.
- Upgrade is explicit; failed validation leaves the old definition/data active.
- Runtime supports current and previous definition schema versions; unknown newer versions are
  rejected.
- Definitions may be exported/imported; import validates schema and safety constraints.
- Data migration is definition-specific: if output keys/types change, the operator sees a preview
  and chooses to apply; old data is retained until replaced.

### 8.3 Failure handling

| Condition | Behaviour |
| --- | --- |
| DNS/TLS/connection/timeout | Keep last data, sanitized error, manual retry |
| 429/503 with `Retry-After` | Show wait; no tight retry |
| 5xx | Keep last data; retry manually; for POST use same idempotency key/body |
| 401/403 | Mark auth required; edit credentials/definition; no retry loop |
| 404 endpoint | Mark definition incompatible; offer edit/import/reset |
| 404 `no_schedule` | Offer re-lookup/check action; no loop |
| 422 address mismatch | Require confirmation/re-setup |
| Required output missing | Treat as definition/schema error; keep last good data |
| Invalid JSON/oversized response | Sanitized error; keep last data |
| Secret decryption failure | Fail closed; request credential re-entry |
| Persistence write failure | Keep previous active data; report error |
| Shutdown | Cancel in-flight request; no background goroutines in v1 |

### 8.4 Scheduled phase

When scheduling is added:

- one bounded worker pool, per-instance serialisation, jitter/backoff;
- respect `Retry-After` and stored `refresh_due_at`;
- conditional requests/ETag;
- resume pending lookups from persisted cursors;
- mark data age/stale in UI;
- use `utils.Clock` and fake clocks in tests;
- reuse the exact manual `Refresh`/`Poll` methods;
- no second definition/execution implementation.

## 9. Security assessment

### 9.1 The new risk: editable definitions

Exposing endpoint/request definitions to the UI turns configuration into a powerful outbound
request capability. This is the price of “fix the endpoint without a release”. It must be
controlled:

- Restrict definition editing to an operator/admin role or an explicit advanced mode.
  Current auth has no roles (`backend/internal/auth/tokens.go:19-24`), so this is a prerequisite
  for safe multi-user deployment.
- Treat definitions as code-equivalent for review; validate before save and show a warning.
- Audit definition changes (who, when, old/new hash).
- Do not allow arbitrary scripts, filesystem access, shell commands or package imports.

### 9.2 Egress/SSRF

Even with editable definitions:

- HTTPS only for cloud definitions; allow an explicit local-host exception later.
- Validate URL templates after rendering.
- Resolve DNS and reject loopback/private/link-local/metadata addresses for cloud definitions;
  validate the resolved address in a custom dialer to mitigate DNS rebinding.
- Limit redirects; redirects must remain within the same allowed host/scheme policy.
- Allow only GET/POST (and perhaps PUT) initially; no CONNECT, arbitrary methods, or raw sockets.
- Bound request time, response body size, decompression size, redirects and concurrency.
- Rate-limit manual tests/refreshes per instance and per definition.
- Do not allow definitions to set arbitrary `Host`, proxy, socket or TLS-skip options.

### 9.3 Secrets and templates

- Secrets stored encrypted (AES-256-GCM) in the dedicated secret bucket.
- Secret refs only in definition/instance config; values resolved only during execution.
- Templates may reference secrets only in designated auth/header/body fields, never in URLs or
  log-visible query strings. If a provider requires a query token, treat it as a special,
  redacted field and document the risk.
- Redact `Authorization`, cookies, tokens, postcode, property ID, UPRN, address and
  `calendar_url` from logs/diagnostics.
- Never return secret values from the API, definition test preview, or error messages.
- Decryption failure fails closed; do not fall back to plaintext.

### 9.4 Definition editor validation

Before saving/activating:

- schema validation (types, required fields, supported operations);
- template validation (known variables only);
- URL/host/egress validation;
- extractor syntax validation;
- output type validation;
- optional live test with explicit confirmation and redacted response preview;
- size/complexity limits on definitions (max steps, fields, outputs, response paths).

### 9.5 User-created add-ons

Do not allow arbitrary user-authored scripts or unrestricted workflows. A declarative definition
is already close to code; restrict it to trusted operators. If broader sharing is desired later:

- signed/bundled definitions;
- import review and hash pinning;
- an isolated execution service/egress proxy;
- per-definition quotas and audit logs;
- no access to host environment/secrets beyond explicit refs.

### 9.6 Privacy

- Setup captures postcode/address/UPRN/property IDs. Store only what is needed; keep them out of
  public APIs/logs and redact diagnostics.
- Add-on data API is protected. Do not add add-on data to public device context or public MCP
  without an explicit exposure decision.
- Display source/quality labels so users understand address match, confidence and limited
  horizons.
- Removal must state whether data/history is retained.

## 10. Phased delivery

### Phase 0 — spec and decisions

- Write a spec/ADR for first-class add-ons and declarative definitions.
- Decide who can edit definitions; add an operator/admin permission if needed.
- Decide v1 output scope (UI/API only vs MCP/automation).
- Define the definition schema v1 and validation rules.
- Define the whenisbins default definition and variable model.

### Phase 1 — manual bins add-on

- Add-on definition registry with embedded default definitions.
- Definition validation, versioning, import/export/reset.
- Add-on instance/data/setup/secret repositories.
- Bounded HTTP execution engine: templates, JSONPath/GJSON extraction, poll plan.
- Setup session with server-side candidates and persisted variables.
- Manual refresh/check-status.
- Protected API and WS events.
- Vue Add-ons page, schema-driven setup wizard, detail/data view, advanced definition editor.
- WhenIsBins default definition (address discovery, lookup with idempotency, stable schedule
  refresh, poll plan).
- Fixture/httptest tests for every response/status branch; fake clock; no live household devices.

### Phase 2 — scheduled refresh

- Bounded scheduler/worker pool using the same `Refresh`/`Poll` methods.
- Backoff/jitter, ETag/304, stale/age display, resume persisted lookups.
- Health/last-run history and diagnostics.
- Restart replay and shutdown tests.

### Phase 3 — consumption

- Add-on MCP tools/resources (`list_addons`, `query_addon`) without device projection.
- Add-on automation trigger/condition types and engine routing.
- Generic add-on dashboard cards/summary components.
- Optional add-on time-series history.

### Phase 4 — breadth

- Additional default definitions (weather, energy prices) to prove the schema.
- Guided definition builder on top of the JSON/YAML editor.
- Optional local-network providers, OAuth variants, signed definition packs.

## 11. Verification plan (not yet run)

No application tests were run for this investigation.

Backend:

- definition schema validation, template validation and extractor validation;
- setup session: dependent fields, candidate membership, variable persistence, expiry,
  owner-scoping;
- idempotency: timeout/5xx reuse key/body; no duplicate lookup; pending lookup resume;
- manual refresh: 200/304/404/422/429/503/5xx/401; keep-last-good-data semantics;
- secret encryption/decryption/fail-closed/redaction;
- definition import/export/reset/version compatibility;
- HTTP egress policy: host allowlist, private-IP rejection, redirects, size/time bounds;
- concurrency: per-instance manual refresh serialisation, no duplicate runs;
- add-on data/API does not appear in device context/MCP by accident;
- race detection on affected packages; isolated storage/httptest only.

Frontend:

- contract validation for definitions/catalog/setup/instance/data/status;
- wizard dynamic fields, candidate select, errors, pending/check-status, offline behavior;
- advanced definition editor validation/test/diff/reset;
- no secrets/PII in browser state, URLs, logs or bundles;
- keyboard/focus/accessibility and 390px/1280px layout;
- Jest/unit tests and Playwright against a mock backend only.

Manual:

- one live WhenIsBins check only when explicitly authorized; no MQTT/household device commands;
- record exact URLs, status and redaction behavior.

## 12. Risks and decisions

| Risk/decision | Impact | Handling |
| --- | --- | --- |
| Editable definitions are powerful | SSRF/data exfiltration | Operator-only editing, egress policy, validation, audit |
| No role model | Any authenticated user could edit definitions | Add admin/operator permission before exposing editor |
| Provider response changes | Definitions break | JSONPath/output edits without code; validation/test/diff |
| Endpoint changes | Definitions break | URL/method/request edits without code |
| Definition DSL grows | Becomes a programming language | Keep primitives small; code only for new mechanics |
| Async lookup | User waits/duplicates | Persist idempotency key/lookup ID; manual check status; later scheduler |
| No scheduler in v1 | Data only refreshes manually | Explicit product choice; later phase reuses methods |
| PII in variables | Address/UPRN/property ID exposure | Private storage, redaction, protected API, no public device context |
| Add-on output history | Not available initially | Optional separate add-on time series later |
| Automations absent initially | Add-on data not in rules | Phase 3 add-on trigger/condition, not device projection |
| Provider terms/limits | 429/blocked | Honour limits, request token, document terms |
| Definition import | Malicious/code-like content | Validation, review, signed packs later |
| Multiple providers | Fallback ambiguity | Manual selection/import first; explicit provenance later |

## Appendix A — Key file references

Backend:

- `backend/main.go`, `backend/internal/hub.go` — startup/wiring and public routes to avoid.
- `backend/internal/api/routes.go`, `backend/internal/api/middleware.go`, `backend/internal/auth/*`
  — protected routes/auth.
- `backend/internal/lanes/lanes.go` — generic per-key serialisation for manual refresh.
- `backend/utils/storage/key-value.go` — BoltDB bucket interface for add-on storage.
- `backend/utils/time.go`, `backend/mocks/types.go` — clock/testability.
- `backend/internal/ws/eventhub.go` — event hub extension.
- `backend/internal/automations/engine.go`, `trigger.go`, `condition.go`, `scheduler.go` —
  device-centric automation boundary; add-on triggers need a separate path.
- `backend/internal/mcp/server/server.go`, `internal/mcp/tools/intent.go` — MCP tool registration;
  add-on tools should not use device metrics.
- `backend/models/settings/*`, `backend/repository/settings.go` — why not to store add-ons in
  `AppConfig`.
- `backend/models/devices/device.go`, `backend/internal/services/*` — device pipeline; reference
  only, not the add-on model.

Frontend:

- `frontend/src/router/index.ts`, `frontend/src/mixins/useTabComponents.ts`,
  `frontend/src/components/hub/settings/Settings.vue` — navigation/settings extension.
- `frontend/src/store/modules/ws/index.ts` — add-on WS event mapping.
- `frontend/src/contracts/api.ts`, `frontend/src/contracts/hub.ts` — authenticated HTTP/runtime
  validation patterns.
- `frontend/src/components/ui/*` — shared controls for the wizard/editor.

External evidence (accessed 2026-10-23):

- WhenIsBins OpenAPI `https://whenisbins.com/v1/openapi.yaml` and guide
  `https://whenisbins.com/v1/guide`.
- Home Assistant contributor docs: “Building integrations”, “Config flow”, “Config entries”,
  “Fetching data”, “The integration manifest”.
- Home Assistant RESTful/template integration docs.
- Airbyte Python/low-code CDK README and connector-builder docs.
- Go library pages: `net/http`, `golang.org/x/time/rate`, `tidwall/gjson`, `ohler55/ojg`,
  `expr-lang/expr`, `cel-go`, `goja`, `golang-ical`, `go-plugin`.

## Appendix B — Glossary

- **Add-on:** user-facing capability configured independently of devices.
- **Definition:** declarative provider/request/extraction/output package, shipped or operator-edited.
- **Instance:** one configured add-on target (property/location/tariff).
- **Variable:** stored value captured from a response and reused in later requests.
- **Setup session:** temporary, owner-scoped state for multi-step setup.
- **Add-on data:** latest canonical output document plus provenance/timestamps.
- **Output:** typed field in add-on data for UI/API/later automations.
- **Manual mode:** user-triggered setup/refresh; no scheduler.
