# External integrations: investigation, DeepSeek review and proposed plan

Spec: [spec.md](spec.md) | Status: M1 implemented and verified locally; Integrations board implemented and verified locally; later milestones deferred | Updated: 2026-10-26
Constitutions: parent/backend/frontend proposed v1.0.0; no ratification or approval claimed.
Repository inspected: `986b923c` on `main`.

## 1. Recommendation and evidence status

AC-49 is a small typography correction in `components/integrations/connections.css`:
explicit regular code weight, system monospace stack, 14px-equivalent size on desktop/mobile
and 1.6 line height. Use the theme page background with a subtle 3% accent tint, normal text,
strong theme border and accent caret. Preserve textarea behaviour, labels and configuration data.
Manually measure text contrast across all theme/mode/preset-accent combinations in the isolated UI.
Verify existing YAML editing browser flow, computed typography and desktop/mobile screenshots;
run required frontend lint/build/unit checks. No backend/API/state changes or new dependencies.

### Section-navigation follow-up (AC-47)

AC-48 additionally covers transient geometry: a shared feature page-header component reserves its
subtitle line and keeps mobile actions on the title row when they fit. Loading announcements are
out of flow. Reuse up to 32 accepted definition DTOs in session-only Vuex state, keyed by provider,
with logout cleanup and authoritative revalidation. Hydrate cached editor/draft state before its
first render; one loading interval covers both metadata and instance reads. Never let revalidation
overwrite a dirty buffer; draft etags stay pinned so concurrent changes still conflict safely.
Use a lightweight content placeholder on cold entry. Preserve guarded focus/scroll restoration.
Reset the actual `.nh-app-main` scroller for integration path changes only; preserve native Help
topic hashes and other views. Distinguish top-level Help tabs (return to the top) from contextual
Help anchors (restore the field position) so an asynchronous draft restore cannot move the header.
Existing AppShell route-key checks must recognize canonical `/integrations` as well as the old
alias, so changing instance/provider IDs remounts the page owner correctly. No other shell layout
or navigation behaviour changes. Regression checks delay reads and compare live header/nav/body
geometry, draft contents, cancellation/logout and distinct provider IDs at both viewports.

Replace detail-only tab markup with one presentational `IntegrationSections.vue`, driven by
typed routes from `domain/integration-navigation.ts`. Carry `instance` in provider/Help query
parameters only when explicitly entered from detail; do not select an arbitrary integration
when a provider has multiple instances. Results/Setup use detail URL hashes for reload and
browser history. Preserve provider editor `new/from` context across Help, existing in-memory
draft retention/focus restoration and confirmation before leaving an unsaved editor.
Use shared tab spacing rather than the detail override that sets bottom margin to zero.
No API/backend/storage/device changes. Source-owned domain tests and ConnectionsPage browser
checks cover all four pages, active links, history/direct entry/drafts and >=12px spacing at
1280x800 and 390x844. Run frontend suite/lint/build/E2E types and isolated browser fixtures.
NH-01/04/05 and FE-01/03/04 apply; NH-03/06 and FE-02/05 retain their existing boundaries;
NH-02 and backend rules are N/A to this frontend-only change. AC-47–48 passed 403 frontend unit
tests, build/lint/E2E types and 44 isolated desktop/mobile browser checks; see implementation-review.

### Current follow-up: generic execution and aligned naming

Implemented package layout:

```text
backend/internal/integrations/
  bootstrap.go, service.go, dto.go   # composition and compatible public API aliases
  runtime/                         # admission, coordinator, setup/definition actions, transactions
  definitions/                     # bounded YAML compiler
  flows/searchselect/              # input, choices, progress and recipe validation
  operations/                      # HTTP primitives and JSON field locations
  capabilities/                    # immutable registry and domain execution contract
    waste/                         # mapping, normalization, projection and renewal
  transport/httpclient/            # pooled HTTPS client and outbound/DNS policy
```

The coordinator is now 175 lines and imports no waste capability. Runtime is the single owner
of locks, queues, transactions and shutdown; its setup/definition action files intentionally
remain with that owner. Pure flow/definition/capability packages do not acquire runtime locks or
own repository writes. Root `Service`/DTO aliases preserve existing application callers.
The v1 `postcode` and `service_ids` JSON names remain compatibility fields; execution treats
search answers and capability identity maps generically, with input-type validation in the flow.

AC-43–46 supersede the older Connections naming below. Root `internal/integrations` is
the composition boundary; `runtime` owns service actions/lifecycle and repository transactions.
Extract YAML compilation to `definitions`, choice/status interpretation to `flows/searchselect`,
and restricted outbound transport to `transport/httpclient`. Retain `operations` for bounded
HTTP primitives. `capabilities` supplies an immutable injected registry and small execution
contract; `capabilities/waste` owns mapping validation, normalization, semantic identities and
coverage/renewal decisions. Child packages never import the composition root.

Store resources as opaque versioned JSON, validated/projected by capability codecs. Retain
the existing flat waste JSON and encrypted storage schema/AAD so old records need no rewrite.
Runtime/public previews, normal execution and catalogue use the same capability registry.
Domain mappings cross the generic YAML boundary as heterogeneous data, then undergo strict
capability-owned decoding. Unsupported capabilities fail closed. Existing HTTP clients in
OAuth/assistant lack integration egress restrictions; preserve the dedicated policy client.

Keep lock ownership, atomic publication, last-good data, revision fences, exact POST replay,
explicit status continuation, bounded renewal and idle-client shutdown. A synthetic capability
fixture proves extensibility without shipping another integration. Canonical UI paths become
`/integrations` with old URL aliases; current copy uses Integrations. No other view redesign.
Remove `doc.go`, keeping useful GoDoc at declaration/algorithm boundaries. See ADR-003.

T054–58 verify new boundaries, legacy-resource compatibility, failure isolation and lifecycle;
run Go suite/build/vet/affected races, frontend suite/lint/build and isolated desktop/mobile
browser navigation. NH-01/03–06, BE-01/03–05 and FE-01–05 apply; device semantics NH-02/BE-02
remain unchanged. Do not touch household data or run browser tests against the personal hub.
These are proposed checks until actual results are recorded in implementation-review.md.

### Integrations landing-page follow-up — implemented

The user approved the grouped Connection board on 2026-10-26. One data widget per connection,
a separate Add catalog, draft reordering and separate hide/delete actions are implemented.
Home-like cards reuse the shared bin SVGs and navigate to the existing full page. Refresh is
independent; layout editing disables navigation. There are no resize or Details-footer controls.
The [board contract](contracts/connections-board.md), [ADR-002](adr-002-connection-board-layout.md)
and [implementation review](implementation-review.md) record ownership and verification.
Home/Devices/Overview/Panel, manual refresh and equal signed-in access remain unchanged.
The [design investigation](designs/connections-dashboard-review.md) retains the alternatives;
its weather/energy examples do not represent implemented features.

### Existing M1 architecture

Build a first-class **integration subsystem** with server-driven setup, durable instances,
validated typed resources and credential isolation. The first release is user-confirmed personal Docker use
with manual setup/refresh (M1). Later roadmap phases cover automatic refresh (M2), assistant/
automation consumers (M3) and provider replacement/second-provider proof (M4). Manual M1 is useful but does not complete the original
request's automatic-refresh objective.

The user approved the name **Integrations** and narrowed its initial UI to its own section.
Add one main-navigation entry after Devices; discovery, setup, results, health, provider editing
and Help stay inside it. Home, Devices, Overview and Panel retain their existing content/counts/
layouts/controls. No pins, room tiles, cross-promotion shortcuts, notices or new global footer copy.
See the [navigation/page-boundary contract](contracts/connections-navigation.md) and
[configuration/help contract](contracts/operations-and-help.md). The
[full audit](review.md), [runtime lifecycle](contracts/runtime-lifecycle.md) and
[API/recovery contract](contracts/api-and-recovery.md) record the latest implementation-path review.

The user has now explicitly rejected a separate integration privilege tier: every signed-in user
must be able to view/edit/modify the shared feature. They require simple desktop/mobile use and
UI repair of provider API/response changes, favouring host-saved configurations similar to YAML.
These access/editability decisions are recorded as user instructions, not pending permission gates.
Manual refresh for the first release and personal Docker deployment were directly confirmed on
2026-10-25. Automatic updates remain a later phase, not a first-release requirement.

Recommend one YAML file per provider definition under the existing host-mounted configs directory,
with Form and editable YAML views over that same document. Validate → save host file → reload/apply;
Git commits are a separate intentional user action. Private household bindings/credentials/runtime
state stay in the protected DB. All authenticated users share the same controls and setup sessions.
No role model, extra feature allowlists or creator-only access checks are required.
Host layout is configs/integrations/<type>/<provider>.yaml, including waste_collection/whenisbins.yaml.

Generic operations execute a bounded search_select recipe with explicit YAML request links;
the common waste validator owns domain meaning. New definitions reuse supported recipes/capabilities.
Request names are arbitrary identifiers, not provider-specific Go methods. Arbitrary new
authentication/transports/transforms may still need code; a YAML file is not a universal runtime.
The [configuration contract](contracts/configuration.md) specifies source-of-truth, editor/mobile UX,
file/DB failure recovery, host Reload, version control and deployment behaviour.

Artifacts: [spec](spec.md), [tasks](tasks.md), [waste contract](contracts/waste_schedule.v1.md),
[synthetic examples](contracts/waste_schedule.v1.examples.json),
[execution/access contract](contracts/execution-and-security.md), and
[proposed ADR](adr-001-integration-boundaries.md), [operations/help](contracts/operations-and-help.md)
and [small typed format](contracts/definition-format.md). M1 is now implemented; see the [implementation review](implementation-review.md) for current evidence. The investigation below retains its original baseline context.

Evidence labels used throughout:

- **Verified locally:** source inspection at the baseline above and the current DeepSeek document.
- **Externally documented:** primary maintainer/provider documentation and public source reviewed
  during this investigation; these are not live end-to-end tests.
- **User evidence:** the two successful requests and failed council request supplied in the prompt.
- **Reported, not reproduced:** DeepSeek's live HTTP/ETag/calendar/council observations. The report
  contains summaries but no accompanying response artifacts to independently audit.
- **Proposed:** architecture, limits, estimates, contracts and delivery phases below.

During the original investigation no application tests or provider jobs were run. Current isolated application checks are recorded in the implementation review. Earlier browser access could not parse OpenAPI
YAML; the latest audit downloaded/parsed it successfully through a documentation-only GET after
sandbox DNS failure. Version 1.1.0, structural findings and SHA-256 are recorded in review.md.
This verifies documented shapes, not live lookup success or every council branch. Never report
DeepSeek's live checks as checks performed by this review.

Home Assistant calls service connectors **integrations**. Its separately managed programs,
formerly add-ons, are now **apps**. Packaging NodeHerder itself as an HA app would be a separate
deployment project. This feature most closely matches integrations.
[Home Assistant terminology](https://contributors.home-assistant.io/apps/)

## 2. Review of DeepSeek's current report

Reviewed: [external-integrations-investigation.md](../../docs/external-integrations-investigation.md),
1,238 lines when initially reviewed. The original report is tracked and remains unchanged.

| Priority / reference | Finding and consequence | Recommended correction |
| --- | --- | --- |
| Delivery decision — earlier report lines 143–169, 715–733 | The original request includes eventual automatic refresh, while the supplied review reports a later manual-first preference. The original plan incorrectly presented automatic-first as settled sequencing. | Manual M1 confirmed 2026-10-25; automatic M2 remains a later roadmap phase. D-01 is resolved. |
| Delivery decision — earlier report lines 152–190, 344–355 | Endpoint/schema repair without a release differs from installing a new provider. The user now directly requires UI-editable configuration. | M1 uses host YAML and Form/YAML editing for supported repairs; D-02 is recorded, without promising arbitrary new mechanics. |
| High — lines 848–851, 880, 895 | `property_token` is extracted unchanged from `result.property_id` and interpolated into the schedule URL. | Keep the full provider ID separately; derive the URL token by validating/removing its `p:` namespace. Use a named typed transform and fixture. The provider guide distinguishes those values. |
| High — lines 444–482, 898–907, 932–933 | Extraction describes JSON bodies, yet examples depend on response headers. ETag is stored without an extraction source; retry timing reads `$.retry_after`; the status expression is not an agreed JSONPath or GJSON operation. | Specify an execution response envelope with status, headers and optional body. Define exact selector and expression dialects. Handle 304 before JSON parsing. Retry-After is an HTTP header, not an assumed body field. |
| High — lines 567–590, 691–707, 943–959, 986–990 | “Canonical” data is an untyped map of definition-owned fields. There is no provider-independent capability/service identity or enforceable output contract. A preview does not protect existing rules when keys/types change. | Version capability schemas independently of provider packages. Keep local resource/service IDs through confirmed provider replacement and validate units, dates, quality and absent/null semantics. |
| High — lines 1030–1082 | Good SSRF controls are listed, but credential authority is under-specified. An editor could change a request to a public attacker host and send a stored token there; private-IP rejection would pass. An advanced UI switch is also not authorization. | Bind secrets to instance/provider and approved origins; enforce grants outside the editable definition. Changing origins requires a separate authorized action and credential reassociation. |
| Medium — lines 1052–1060 | AES-GCM is named without key provisioning, rotation, backup or missing-key recovery. | Define those operational contracts before writing encrypted data. Keep key material outside the DB and JWT signing key; use versioned ciphertext and authenticated record identity. |
| Medium — lines 594–601 | Separate buckets and atomic publication are proposed without a concrete transaction boundary. The existing KV wrapper opens its own DB handle and exposes no arbitrary cross-bucket transaction. | Use one integration repository owning one bbolt handle and explicit atomic operations; or keep each atomic aggregate in a single record. Do not open the same file independently for every bucket. |
| Medium — lines 57–63, 1006 | Keyed FIFO execution is useful, but it neither caps the number of active keys globally nor cancels started tasks by itself. “No background goroutines” is not a meaningful guarantee with lanes/HTTP. | Use global/provider concurrency caps, coalescing and cancellable I/O; define ownership and shutdown. Do not hold a lane through a two-minute polling sleep. |
| Medium — lines 293–295, 943–959, 1013 | Refresh metadata alone is insufficient for horizon exhaustion or derived countdowns. The supplied dates end before the supplied refresh due date. | Track future coverage per service; recompute relative dates locally and refresh/recover before useful data runs out. Never extrapolate collection dates. |
| Medium — lines 1095–1137 | No complexity estimates or acceptance traceability accompany the phases; key decisions are embedded as product requirements. | Add explicit outcomes, estimates, decision gates, failure tests and rule checks. |

**Where the report is strong:** first-class integration storage instead of device impersonation;
backend-owned candidate selection; secret references; preserving last good data; versioned definitions;
and identifying the public device/MCP boundary. Those recommendations are supported by the code.

**Where the disagreement is a trade-off:** declarative HTTP definitions are viable. The issue is
matching their scope and execution semantics to the actual repair requirement. Dynamic forms do not require
declarative provider execution. A small schema becomes a workflow language once it covers branches,
polling, retries, templates, array transforms, auth and migrations; that development cost must be explicit.

The report's current goja toolchain warning is supported: the upstream module and README currently
require Go 1.26, whereas this backend declares Go 1.24. Older pinned revisions would need separate
evaluation; this is not a reason to upgrade NodeHerder for this feature.
[goja go.mod](https://raw.githubusercontent.com/dop251/goja/master/go.mod)

## 3. Current architecture and actual reuse

All paths below exist today; proposed paths appear later.

| Area / code | Verified behaviour | Reuse / limitation |
| --- | --- | --- |
| `backend/main.go:39`, `backend/internal/hub.go:83` | Store and transport/controller composition with application context | Wire a separate integration service/repository here; initialization failure must not disable independent device management. |
| `backend/store/initialiser.go:11` | In-memory device repository; bbolt-backed metrics, assistant and settings repositories | Reuse storage technology and constructor patterns. There is no integration store. |
| `backend/utils/storage/key-value.go:13`, `:40`, `:100` | Owned bytes, 0600 DB file, single-key writes and some batch operations | Useful primitives, not a ready multi-record integration unit of work. |
| `backend/repository/settings.go:12`, `backend/models/settings/config.go:150`, `backend/models/hub/state.go:8` | Settings serialized as hub configuration and included in hub snapshots | Do not add credentials/provider-private state to AppConfig. |
| `backend/internal/api/routes.go:636`, `backend/internal/auth/auth.go:151` | Assistant and Google OAuth have individual bounded HTTP clients | No general provider client, outbound policy or retry framework. Reuse patterns, not those handlers. |
| `backend/internal/auth/middleware.go:59`, `tokens.go:24` | Cookie/bearer JWT auth; claims contain user ID/name but no roles | Use protected routing and the agreed equal signed-in access; no extra feature roles. Apply origin/credential/revision protections to definition mutations. Google login does not provide external integration token management. |
| `backend/internal/lanes/lanes.go:101`, `:139` | Per-key FIFO, queue bound per key; started Do calls wait for completion | Reuse only with integration-owned admission/cancellation policy; not a global worker pool or durable queue. |
| `backend/utils/time.go:20`, `backend/store/tasks.go:38` | Injectable clock/timers and context-owned cleanup task pattern | Reuse controllable time and explicit start/stop/join ownership. |
| `backend/internal/automations/scheduler.go:184`, `:268`, `handler.go:91` | In-memory timer jobs, London wall-time handling and enable/disable automation registration | Generic timer mechanics exist, but durable provider due times, retry quotas and async continuation do not. A small coordinator is simpler than stretching this API. |
| `backend/internal/automations/engine.go:88`, `context.go:11`, `device.go:20` | Recipes/evaluation contexts are device-centric | New integration event/condition consumer needs explicit engine and recipe UI work. Merely adding a trigger type is insufficient. |
| `backend/internal/services/device_lifetime.go:69`, `:147`, `backend/models/devices/device.go:470` | Device availability, seed-trigger behaviour and accepted-state-before-output pipeline; non-bridge discovery whitelist/numeric assumption | Do not inject schedules through that pipeline. This is a practical boundary, not a claim that virtual entities are always invalid. |
| `backend/internal/mcp/server/server.go:254`, `tools/intent.go:54` | Read-only query_device tool aimed at metrics; assistant histories persist messages | No generic agent integration runner, semantic memory store or external MCP-client scheduler was found. Prompt context overstates what is implemented. |
| `backend/internal/hub.go:64` | Device context and MCP routes explicitly public | New resources must stay out of these routes unless auth is deliberately redesigned and tested. Adding a new tool to the same public MCP server would expose it too. |
| `frontend/src/contracts/api.ts:15`, `contracts/hub.ts:8`, `services/assistant.service.ts:1` | Authenticated fetch and runtime parsing patterns | Add an integration service/contracts; strengthen validation for new schemas instead of relying on the broad isApiResponse guard. |
| `frontend/src/store/index.ts:20`, `store/modules/ws/index.ts`, `router/index.ts:21` | Vuex, WS event mapping and authenticated lazy routes | Add one Vuex module/page, existing PrimeVue/shared controls. Do not persist setup/credentials in localStorage. |

Home areas and Panel are currently device-oriented. Earlier samples explored extending their
models, cache writers and renderers; those changes are now outside scope. The existing dashboard
cache persistence issue is not touched by this feature. No dashboard schema/WS migration or core
Home/Panel refactoring is required for the contained Integrations UI.

The repository has no general plugin loader, provider catalog, durable external workflow runtime,
encrypted credential repository or integration lifecycle. MQTT, automation serializers and MCP
registrations are specific extension points, not an existing general integration framework.

## 4. Research and free alternatives

Home Assistant's reusable ideas are configuration entries, multi-step config flows, reauth/reconfigure,
typed entities and a shared data coordinator. Its manifest identifies an integration; provider
behaviour is not universally described by that manifest. Borrow the separation rather than infer
that friendly setup requires YAML execution.
[config flows](https://contributors.home-assistant.io/core/integrations/config-flow/),
[manifest](https://contributors.home-assistant.io/core/integrations/manifest/),
[coordinator](https://contributors.home-assistant.io/core/integrations/fetching-data/)

openHAB separates bindings, things, channels and items, including web services. This supports
separating provider protocol from consumer-visible state; it does not require adopting its runtime.
[openHAB concepts](https://www.openhab.org/docs/concepts/)

| Solution | What is actually reusable / free status | Assessment for NodeHerder |
| --- | --- | --- |
| Waste Collection Schedule for HA | MIT-licensed Python integration, council sources including Sheffield, calendar support | Strongest existing free bins ecosystem found. Best option if the household already runs HA. Otherwise adds HA/Python deployment and a bridge; not a Go library. Adapt selected provider knowledge with notices, not its whole runtime. [project](https://github.com/mampfes/hacs_waste_collection_schedule), [license](https://raw.githubusercontent.com/mampfes/hacs_waste_collection_schedule/master/LICENSE) |
| Direct Sheffield council endpoint | Existing connector POSTs councilId 1 and genuine UPRN, reads actual_scheduled_date/service | Plausible direct source with no intermediary subscription. Its access terms, discovery journey and ongoing stability still need validation; public reachability is not a guaranteed support contract. [connector source](https://raw.githubusercontent.com/mampfes/hacs_waste_collection_schedule/master/custom_components/waste_collection_schedule/waste_collection_schedule/source/sheffield_gov_uk.py) |
| Independent council/provider ICS feed | Standard calendar import where actually available | Often simplest for dates/reminders. Requires event mapping, bounded recurrence support and coverage rules. No independent Sheffield ICS endpoint was verified here. HA also has an existing remote-calendar integration. [Remote Calendar](https://www.home-assistant.io/integrations/remote_calendar/) |
| WhenIsBins | Matches supplied working setup journey | Sensible initial UK-wide candidate, subject to usage/access confirmation; do not select it solely because two requests worked. See provider notes below. |
| Node-RED | Apache-2.0 core, installable npm nodes; individual nodes can differ | Good optional external adapter/workflow service for technical users. Does not remove the need for NodeHerder's normal setup, stable data, credentials and lifecycle contracts. [license](https://nodered.org/about/license/), [installation](https://nodered.org/docs/user-guide/runtime/adding-nodes) |
| Airbyte | Declarative request/auth/extraction/pagination and a connector-builder UI | Useful design reference, primarily source-data synchronization. Not a small Go UI integration library; would introduce a separate platform. [builder](https://docs.airbyte.com/platform/connector-development/connector-builder-ui/overview), [OAuth](https://docs.airbyte.com/platform/connector-development/config-based/advanced-topics/oauth) |
| Benthos / Redpanda Connect | MIT core; free Apache-2.0 connectors plus enterprise features | Credible Go pipeline alternative if transformation/stream processing becomes a demonstrated need. Not a ready household setup/config-entry system. Check selected components rather than assume every connector is free. [product](https://www.redpanda.com/connect), [licensing](https://docs.redpanda.com/connect/get-started/licensing/) |
| n8n | Sustainable Use License, with enterprise distinctions | Available for certain free self-hosted uses; not interchangeable with an unrestricted permissive library. Adds a separate workflow platform. [license](https://raw.githubusercontent.com/n8n-io/n8n/master/LICENSE.md) |

Waste Collection Schedule is itself previewing typed retrieval/parsing/transform building blocks
and declarative sources. Its README labels this a 3.0 alpha and warns of changed labels affecting
automations. That supports a hybrid direction and stable canonical identities; it does not establish
an already proven drop-in Go runtime.
[maintainer architecture announcement](https://raw.githubusercontent.com/mampfes/hacs_waste_collection_schedule/master/README.md)

For a second capability, Open-Meteo is a realistic simple HTTP experiment: its free hosted API is
non-commercial, rate-limited and has no uptime guarantee; attribution applies. Do not advertise
all distributed/commercial deployments as free. Octopus documents public product/tariff discovery,
which would test dependent selection, pagination and time/units separately from account credentials.
[Open-Meteo pricing](https://open-meteo.com/en/pricing),
[Octopus API](https://docs.octopus.energy/rest/guides/endpoints/)

## 5. Architecture comparison

Costs and security ratings below are engineering judgments, not benchmark results. Typical
household HTTP integration performance is dominated by upstream requests, not map lookup speed.

| Approach | Complexity / maintenance / operational cost | Extensibility and security | Decision |
| --- | --- | --- | --- |
| Purpose-built bins feature only | Lowest initial scope; repeated infrastructure for future providers | Easy to reason about, limited breadth | Valid if scope shrinks to bins only; not enough for stated framework. |
| Code-only provider modules | Low/medium; compiler and fixture tests | New provider or endpoint/mapping repair can require release | Does not satisfy a confirmed no-release repair requirement by itself. |
| JSON/YAML HTTP packages | Medium/high language/tooling/security cost | No rebuild for supported operations | Defer unrestricted packages; implement only bounded YAML HTTP definitions in M1. |
| Typed adapters + host YAML definitions | Medium/high bounded YAML parser/editor cost, ordinary provider tests | Supported repairs without release; new mechanics remain code | Revised proposed M1; exact grammar/authority defined in the execution contract. |
| Embedded Goja/Starlark | Interpreter plus host API, resource and debugging design | Broad author flexibility, harder limits; not inherently safe by calling it scripting | Defer without a genuine authoring requirement. |
| CEL / Expr for transforms | Smaller expression boundary, still needs cost/input/output limits | Useful bounded mappings; not network/workflow lifecycle | Consider one language only if typed built-in operations become inadequate. [CEL](https://github.com/cel-expr/cel-go) |
| Wasm/wazero | Module ABI, host calls, limits, package distribution | Potential no-rebuild isolation; host capabilities remain security-critical | Later if third-party executable extensions become central. [wazero](https://wazero.io/) |
| stdlib plugin | Toolchain/dependency matching, platform limits, no unload | In-process executable authority | Reject for this use. [Go warnings](https://pkg.go.dev/plugin) |
| hashicorp/go-plugin | Processes, protocol/version/lifecycle, per-platform distribution | Mature local RPC; crash separation does not automatically sandbox filesystem/network | Viable later for independently distributed trusted connectors, not first bins integration. [maintainer README](https://raw.githubusercontent.com/hashicorp/go-plugin/main/README.md) |
| Workflow platform / Temporal | Separate deployment or substantial runtime concepts | Powerful orchestration, excessive operational scope here | No demonstrated need; avoid. |

Use stdlib `net/http`, `encoding/json`, crypto and existing bbolt/clock first. Existing oauth2 is
useful client machinery, but refresh tokens, scopes, callback state and encrypted persistence
still require application ownership. M1 fixes response locations as JSON Pointers with typed references; a selector library is optional,
not an architectural requirement. GJSON/JSONPath are alternatives, not additional M1 grammars.
[GJSON syntax](https://raw.githubusercontent.com/tidwall/gjson/master/SYNTAX.md),
[OjG](https://github.com/ohler55/ojg), [oauth2](https://pkg.go.dev/golang.org/x/oauth2)

Do not treat `text/template` as a restricted substitution mechanism merely by restricting functions:
it has control structures and assumes trusted authors. Prefer typed reference nodes and normal
JSON/query/path encoding. A literal `{{postcode}}` is not standard Go map-field template syntax.
[text/template](https://pkg.go.dev/text/template)

## 6. Concrete contracts and ownership

The linked contracts are normative within this draft; this plan summarizes rather than duplicates
field/authorization/key rules. Runtime bootstrap, missing API paths and recovery are fixed by the
linked audit contracts; implementations must not fill those gaps independently. The M1 source owners now implemented are:

```text
backend/models/integrations/                 domain/config/setup/capability types
backend/repository/integrations.go           one encrypted runtime DB, keyring/rotation/backup, atomic aggregates
backend/repository/integration_definitions.go host YAML, atomic writes/hash/journal recovery
backend/internal/integrations/
  bootstrap.go                              optional feature ownership/degraded init/Close join
  definitions.go                            definition catalog, validation, activation and repair
  service.go                                instance lifecycle, revision/generation fencing
  setup.go                                  shared revision-scoped setup and durable attempts
  configuration.go                          strict YAML compilation and fixed HTTP operations
  http.go                                   bounded transport, approved origins/IP policy
  dto.go                                    private/public response boundary
  controls.go                               credentials, enable/disable and deletion
  test_definition.go                        isolated offline/live previews
  publication.go                            persistence-failure checkpoint handling
  janitor.go                                expired private rows/orphan cleanup, no provider I/O
  coordinator.go                            manual admission in M1; due work/projections in M2
  capabilities/waste.go                      normalization and read-time projections
  identity.go                               stable resource/service identity registry
  operations/pointer.go                     bounded JSON Pointer extraction
  operations/request.go                     shared HTTP request/capture operation
backend/internal/api/integrations.go         protected DTOs and mutation-origin enforcement
frontend/src/contracts/integrations.ts        runtime ingress validation
frontend/src/services/integrations.service.ts
frontend/src/store/modules/integrations/index.ts
frontend/src/components/integrations/         Integrations catalog, wizard, health, Form/YAML editor, Help
```

The generic runner receives context, definition, state and injected policy-bound HTTP/clock/scoped
credentials. Its outcomes are validated result, unchanged, pending input/status or typed failure.
Request selection and setup/refresh wiring are in YAML. M1's reusable search/select recipe composes
HTTP, extraction, form/choice, finite status checking, bounded poll and validated save primitives;
no provider-named runtime adapter or arbitrary workflow language is needed. Capability validators
handle domain meaning. See the operations/help contract for exact references and schema boundaries.

[Waste contract](contracts/waste_schedule.v1.md) fixes exact wire fields/nullability, date/quality/
coverage/freshness enums, persisted UUID identities and provider service-key mappings. Its typed field descriptions
and five complete examples cover exact, partial, provisional, confirmed-empty and unavailable data.
Provisional/unavailable examples describe transient outcomes, not durable replacement snapshots.
Typed decoding and focused domain checks enforce those rules; no runtime JSON Schema engine is required.

One bbolt repository owns private runtime data and active compiled revisions. Its path must be
strict, without the generic GetDataDir/KV temporary fallback; failure is local to Integrations. Host YAML is the
single editable definition source; compiled snapshots/revision journals are recovery/runtime data,
not a second independently editable configuration store. A separate definition repository owns
safe file paths, content-hash conflicts, atomic rename and reload. There is no filesystem/DB atomic
transaction: saved_not_applied is explicit and startup/Reload reconciliation preserves last-good
behaviour. Never silently move definitions into a temporary folder when the configured host mount
is read-only. Current backend Compose already mounts ./backend/configs at /backend/configs.

## 7. User setup and protected API

### Integrations entry and page boundaries

Desktop/sidebar and mobile/drawer retain Overview → Home → Devices → **Integrations** →
Automations → Assistant → Console → Settings. Inside Integrations: Waste collection → WhenIsBins →
postcode → address → saved results. Tabs are Results, Setup, Configure operations and Help.
Provider configuration edits show their shared scope; ordinary setup requires no HTTP knowledge.

The [navigation contract](contracts/connections-navigation.md) fixes routes and page isolation.
Backend /api/integrations naming and configs/integrations remain. No Show on Home, room assignment,
Panel tile, connection-specific dashboard field or settings WS payload is introduced. Devices and
other existing pages have no links/notices promoting Integrations; users use the existing main menu.
All other existing navigation is retained; no connection data or extra product copy appears
outside Integrations.

### Setup and API

Normal flow: Integrations → Waste collection → WhenIsBins → postcode → exact address selection → save → current
collections, source/coverage and health. Unsupported required inputs show a clear unsupported journey,
not a guessed property. Dynamic options and step/revision state remain server-owned. Generic Vue
controls render form/choice/progress/complete/error steps. No raw provider HTML or arbitrary JS.

Every authenticated user has the same Configure/Data/Definition controls, including refresh,
editing, live test, restore and removal. Definition has a form for supported repairs and editable
YAML for detailed changes, both representing the same host document. Users can continue each
other's setups; revisions protect shared edits. No separate feature role/allowlist migration.
Existing login, cookie Origin/header checks, bounded egress and secret handling remain in place.

The configuration contract requires simple mobile forms, 44px touch targets, clear save/error
feedback, a contained YAML pane and preserved unsaved buffers. Test normal setup and an actual
endpoint/mapping repair at 390px/1280px. Do not make raw YAML necessary for ordinary phone use.
M1 uses authenticated REST polling for integration data. No integration payload, private data or
resource references are added to the global settings WS hub or device snapshots.
Configure operations includes an in-app Help tab/linkable child page. It explains format versions,
requests/references, recipe wiring, mappings and repair examples; contextual links return to the
correct field without losing the draft. Help is bundled, accessible and responsive, not remote AI
content. Add IntegrationsHelp.vue and an authenticated help route using existing router conventions.

All routes are protected and use existing :id parameters:

```text
GET    /api/integrations/catalog
POST   /api/integrations/setup
GET    /api/integrations/setup/:id
POST   /api/integrations/setup/:id/step
DELETE /api/integrations/setup/:id
GET    /api/integrations/instances
GET    /api/integrations/instances/:id
GET    /api/integrations/instances/:id/resources
PUT    /api/integrations/setup/:id/credentials/:slot
PUT    /api/integrations/instances/:id/credentials/:slot
DELETE /api/integrations/instances/:id/credentials/:slot
POST   /api/integrations/instances/:id/refresh
POST   /api/integrations/instances/:id/check-status
PUT    /api/integrations/instances/:id
DELETE /api/integrations/instances/:id
GET    /api/integrations/definitions/:type/:provider
PUT    /api/integrations/definitions/:type/:provider
DELETE /api/integrations/definitions/:type/:provider
POST   /api/integrations/definitions/:type/:provider/validate
POST   /api/integrations/definitions/:type/:provider/test
POST   /api/integrations/definitions/:type/:provider/reload
POST   /api/integrations/definitions/:type/:provider/restore
POST   /api/integrations/definitions/reload
```

Definition PUT saves/applies YAML with expected content hash; response reports file/active hashes.
Destination/credential reconnection appears in the same user's Save review, not a separate privilege
workflow. Credential submission remains write-only for secret values. The API/recovery contract defines the exact envelope/preconditions, 410 setup expiry and 428
missing revisions. API success/errors include 200 reads/
completed saves, 201 persisted creation, 202 admitted async work, 400 invalid input/YAML, 401 invalid
session, 403 unsafe origin/destination, 404 missing record, 409 stale file/session revision, 429 busy
and explicit non-success saved_not_applied/config_write_failed states where persistence/activation
fails. GET status never performs provider I/O. Never hide disk/activation failures as success.

Actual router supports PUT but no PATCH helper. CORS lacks the custom mutation header **and**
If-Match/If-None-Match; all must be added only for integration paths, with safe exposed headers. Scope the
header change and actual-request origin enforcement explicitly. Existing CORS alone does not reject
disallowed requests. Do not change unrelated public/device routes implicitly.

Keyboard/focus, visible non-colour error states, cancellation/deletion, and 390px/1280px layouts are
required. No secrets, setup candidate lists or provider-private bindings in localStorage/URLs/logs.

## 8. Supplied WhenIsBins flow and profile behaviour

The working example remains user evidence; the new audit additionally verifies OpenAPI 1.1.0
structure without a live lookup. Use its documented shape plus isolated fixtures, not that evidence
as a claim of end-to-end provider availability:

1. URL-encode postcode for GET /v1/addresses. Validate council/required_input and retain permitted
   candidate labels/IDs privately. Render text labels and verify membership on selection. Changing
   postcode invalidates old candidates. Property ID is not a UPRN or a NodeHerder resource ID.
2. Persist canonical body bytes and a new idempotency key before POST /v1/lookups, using postcode
   plus the selected property_id. No string-concatenated JSON. Reuse exact body/key only within a
   verified provider window after an ambiguous response; changed selection is a different operation.
3. Inspect status independently of HTTP 201. done validates/commits exact data; queued/running/partial
   retains lookup identity. M1 Check status uses that lookup after Retry-After; M2 automates the same
   continuation. Never submit again simply because the first lookup is slow.
4. The supplied done result includes refuse, paper/card and mixed-material services, concrete dates,
   confidence/limited horizon, UPRN, council-specific property identity and calendar URL. Normalize
   service meaning without assuming category/colour is unique. Discard direct matched-address fields from resource DTOs, retain the minimal encrypted binding,
   and encrypt snapshots too because preserved provider notes can contain private text.

The prior council request used a p: property ID as a UPRN. It does not establish failure with a real
UPRN. Existing Sheffield connector code uses a valid UPRN; DeepSeek's reported corrected live call
has not been independently reproduced here. Direct friendly council discovery remains extra work.

Documented provider facts: preserve full property_id but remove p: for the schedule URL token;
conditional schedule GET uses ETag and 304, which is not fresh council verification. Poll one lookup
respecting Retry-After; provisional neighbour results are not persistent schedules/reminder sources.
A stored schedule GET can start background refresh, and its ICS feed shares that same source rather
than being an independent fallback. Regular use/free-token and commercial-contact conditions apply.
[Provider guide](https://whenisbins.com/v1/guide)

The HTTP profile explicitly extracts headers/body/status separately and validates token conversion.
No '$.retry_after' body assumption, missing ETag extraction or arbitrary status-expression dialect.
M1 supports the verified property_id search/select journey. Different council input journeys must
be supported by an explicit recipe/configuration and tested before being advertised; do not guess.
M1 is exact-choice only. The YAML discriminator rejects other required_input journeys; provisional/
non-exact input never publishes a household result and no representative opt-in is submitted. A
future consent/display branch must first add a supported recipe, rather than remain an M1 promise.

Prefer the documented JSON schedule resource for refresh; offer the returned ICS subscription URL.
An independently sourced official ICS provider can be evaluated separately. Current free-source
comparisons remain in section 4; selection still depends on intended deployment/access and coverage.

## 9. Manual work, later scheduling, caching and recovery

M1 service methods admit bounded user-triggered refresh/Check status, coalesce duplicate clicks and
respect shared origin/provider quota deadlines. One active attempt per instance, four global external
requests and proposed 50-instance capacity. Each attempt has bounded context/time/decoded size.
A local Save claims its setup once and preserves a minimal completion receipt; repeated local
HTTP submission cannot create another instance/provider key. Pending state survives restart for
manual continuation until expiry; M1 starts no recurring provider
refresh/lookup polling. The independent expired-record janitor is owned/cancelled/joined explicitly.

M2 adds one context-owned due-work coordinator using utils.Clock plus durable records and a simple
scan/timer. It resumes enabled pending/due work with jitter. Poll waits persist retry-at and release
worker capacity, rather than occupying a lane for minutes. Provider-wide rate limits pause affected
instances together. Honor Retry-After seconds/date; otherwise bounded backoff/jitter. Provider calls
are never hidden inside local GET reads. Startup recovery does not blindly repeat a POST.

A declared renewal_request can start one new logical lookup after exact no_schedule or exhausted
service dates, using saved confirmed inputs and a new immutable attempt/key. Unknown 404s, missing
inputs and invalid property identities never create a guessed renewal. Frequency/quota/backoff
limits and total work budgets apply. See API/recovery for the complete generic state machine.

Keep last_checked_at, captured_at, source_retrieved_at, fresh_until and coverage independent. A 304
updates successful-check metadata, not source time/expiry. Validators are request/config/binding/
credential scoped; config repair invalidates them. Without a matching retained payload, allow only
one bounded unconditional retry, never publish an empty or old-mapping payload as newly validated. A provider hint cannot substitute for
per-service useful coverage: in the supplied example the dated horizon ends before refresh_due_at.
Do not manufacture weekly recurrences or interpret missing service dates as cancellation.

capabilities/waste.go owns pure read-time countdown/freshness projections using injected time and
Europe/London. Visible UI local reads every 60 seconds/on focus update across midnight without
provider I/O. M2/M3 coordinator projection/reminder deadlines are zoned, separate from provider work;
no second scheduler or output-field timers. Recompute after restart/DST changes.

| Boundary | Behaviour |
| --- | --- |
| Timeout/uncertain POST | Persist immutable body/key; reuse only with verified idempotency guarantee; otherwise reconciliation required. |
| Pending/partial | Same lookup, qualified data; M1 manual continuation, M2 scheduled. No expiry promotes provisional data. |
| 304 | No body parse; preserve payload/provenance/source age. |
| 429/503 | Respect provider/global delay; no busy-loop or parallel retry storm. |
| 401/403 | Provider-specific auth/denied classification; stop credential retry loop. |
| 404 | Distinguish no saved schedule, removed endpoint and missing local instance; bounded explicit re-lookup where supported. |
| Invalid/missing/oversized response | Fail before publication; keep last validated exact snapshot. |
| Persistence failure | No new revision/cache validator/checkpoint is announced until atomic commit. |
| Disable/delete/reconfigure mid-fetch | Advance generation/cancel; reject obsolete results before commit and notification. |
| Shutdown | Stop intake, cancel/join workers and janitor/coordinator, close DB after writes finish; proposed ten-second test target. |

## 10. Evolution, replacement and consumers

Version independently: profile grammar, immutable profile revision/hash, operation/recipe behaviour,
persisted binding config and capability schema. Pending operations keep the version that created
them. Activation changes future work only; incompatible updates preserve the old active binding and
snapshot. A sample response does not prove all error/optional/partial branches.

M4 explicitly delivers replacement: private candidate binding/snapshot, confirmed property identity,
preview of services/dates/coverage/quality, ambiguity resolution, affected-automation review, atomic
mapping/binding activation, and seven-day previous-version rollback. Local resource IDs persist;
service IDs persist only for confirmed semantic mappings. Unmapped/changed services do not silently
retarget recipes. A second capability/provider alone is not replacement acceptance evidence.
See exact identity/tombstone/coexistence rules in the waste contract.

Provider shutdown is visible deprecation plus last-good age and guided replacement, never arbitrary
fallback. No mix of conflicting calendars or postcode-only property switching. Stable UPRN/council
identifiers can assist confirmation, but p: identifiers remain provider-private.

M3 introduces an explicitly protected read-only assistant/MCP consumer and one agreed automation
path. Adding a tool to the existing public MCP server would expose it; protected REST alone does
not fix that boundary. No generic external action/device-command tool. Automation uses resource/
service identity, revision, exactness/freshness/coverage/subscription eligibility and occurrence
identity; do not replay cache restoration or provisional data as a fresh event. Existing recipes,
state/physical-event distinctions and device command validation remain unchanged.

## 11. Security and operational contracts

The configuration and execution/security contracts fix equal authenticated access, shared revisions, cookie mutation checks,
keyring file custody, AES-GCM record context, resumable rotation, YAML+DB+separate key backups,
setup/body/consent retention and logical deletion limits. These are concrete proposed choices,
not claims that code implements them. One-household scope retains existing authentication and system-wide validation.

Cloud outbound requests require approved HTTPS origins and resolved destination checks at dial time
for IPv4/IPv6 loopback/private/link-local/metadata destinations. Preserve hostname TLS verification,
control proxy use, reject or revalidate redirects, and apply policy to returned URLs too. Secret
origin grants are outside editable profiles; public attacker hosts are not permitted credential
recipients merely because they are not private IPs. Later local-network access needs a separate grant.
[SSRF reference](https://cheatsheetseries.owasp.org/cheatsheets/Server_Side_Request_Forgery_Prevention_Cheat_Sheet.html)

Bounds cover decoded payloads, profile grammar, traversal/input/output counts, deadlines, concurrency,
queue admission and retention. No model output or profile can bypass normal deterministic validation.
Logs/diagnostics show sanitized codes, timings, revision IDs and non-PII counters. Device/MQTT operation
must survive integration provider/key failures. Missing/unreadable encryption keys never cause silent
replacement keys or data overwrites. Rotation/restore is tested with mixed key versions, not only a
new empty database.

## 12. Revised milestones, estimates and decision gates

User decisions are recorded: manual refresh first (D-01), UI-editable host YAML/equal signed-in
access (D-02/D-03), and personal Docker use (D-05). The keyring-file contract (D-04) now has concrete
Docker paths/mount/backup behavior. No delivery/deployment clarification remains pending. Technical
fixture/provider-behavior/deployment verification stays in T002/T003; it is not another permission
gate. Later consumer scenarios are M3 gates, not blockers for Integrations-only M1. The user subsequently authorized M1 implementation. Local technical self-review and checks are complete; maintainer ADR adoption and real deployment are not claimed.

The old 16–28/25–46 day ranges are withdrawn as a planning baseline. They did not separately budget
the repair editor, wire contract and replacement workflow. The latest scope contains the UI entirely within Integrations. It removes feature-role
work and adds YAML source-of-truth/host synchronization; do not infer a net time saving without
estimation. The existing provisional ranges below are coarse effort scenarios, not commitments; re-estimate work packages after gate M0 and
captured fixtures. One engineer familiar with Go/Vue; review waits/deployment approvals excluded.
Do not promise a fixed delivery date from these numbers.

| Milestone | Scope and exit evidence | Provisional effort |
| --- | --- | --- |
| M0 — Contracts/evidence | Delivery/deployment decisions confirmed; verify strict YAML/file contracts, provider fixtures/idempotency and task boundaries; revise estimates. | 4–8 engineer-days; may expand for changed deployment needs |
| M1a — Secure manual backend | One repository, schema/identity, keys/retention, bounded transport, generic recipe/request operations, strict YAML compilation, host-save/reload recovery, manual refresh/check-status and anonymous/origin negative tests. | 10–18 days |
| M1b — Integrations UI + repair | Top-level Integrations/routes, catalog/wizard/data/health, shared sessions, Form/YAML editor, offline/live test, save/reload/restore, Git-friendly files, contextual/full Help, consent/error/expiry/touch/viewport checks; no recurring provider schedule. | 8–15 days |
| M2 — Automatic lifecycle | Durable coordinator, quota backoff, pending/due recovery, coverage exhaustion, projection deadlines and shutdown/race evidence. | 5–10 days |
| M3 — Consumers | Protected assistant/MCP transport plus one agreed automation path/recipe UI and occurrence rules. | 6–12 days |
| M4 — Replacement and reuse | Actual second provider/binding, candidate preview/mapping, previous/current coexistence, atomic identity preservation/rollback and automation review; prove common runtime unchanged. | 7–14 days |
| Later — New-provider packages | Generic package distribution/import, signatures, additional transports/auth/actions/multi-household only after demonstrated demand. | Separate estimate; not included |

Indicative M0+manual M1: 22–41 days; through M4: 40–77 days. These wider ranges acknowledge unresolved
scope and are not verified productivity estimates. UI repair is now required. Re-estimate the equal-access YAML/file and contained UI design after M0; if arbitrary
no-rebuild providers/multi-tenancy are required, redesign and re-estimate
rather than hide them in M1. No broad marketplace or workflow framework is included.

## 13. Traceability and verification

Current acceptance mapping, every NH/BE/FE rule, actual commands/results, review corrections and
unrun operational checks are in [implementation-review.md](implementation-review.md).
[Tasks](tasks.md) records completed M1 work and unchecked later milestones. Tests belong to
production source owners; frontend unit companions mirror source paths. Browser verification
uses owned fixtures and a run nonce, with server reuse disabled. No household devices or live
provider jobs were used. Review is an agent self-review, not maintainer approval or ratification.

## 14. Historical planning handoff and review disposition

This section records the pre-implementation handoff. Statements below about unstarted/unrun work
refer to that earlier planning stage; current status is in the implementation review. M1 was
subsequently authorized, implemented and tested. M2–M4 remain roadmap work.

The DeepSeek review's contract, authorization, identity/replacement scheduling, retention, date-owner,
estimate and test-detail findings are incorporated. tasks.md was provided for the subsequent implementation.
Manual first release/personal Docker use are confirmed; UI repair, equal authenticated access and simple desktop/mobile use
are now direct user requirements. Host YAML/Form editing is the recommended configuration design. Provisional
no-cache/no-reminder policy remains provider-backed and now separately permits transient display with
future scoped consent; M1 rejects representative journeys outright. The PATCH mismatch is removed. The historical DeepSeek report remains untouched.

Do not start production-code edits until dependent scope/security/provider evidence gates are resolved.
This handoff contains planning artifacts only: no production-code changes, application tests,
live provider jobs or household device commands. Sample HTML/screenshots and their checker were
removed at the user's request; the local planning branch/commit is authorized, with no push. Before committing, include all linked planning artifacts and the referenced
DeepSeek report, or replace its link with a durable reference; an untracked file is not repository history.


### Final scope review

The earlier Home/Garden/Panel placements were exploratory samples. The user's latest decision is
Integrations-only UI, reached through its main-navigation item. The updated spec replaces those
positive placement requirements with AC-26/27 preservation/no-promotion checks; AC-24/25, FR-14/15
and T029/T030 are retired and not reused. Removed their proposed models/renderers/cache work and
reduced M1b's coarse range accordingly. Future widgets require separate scope, not an implicit
promise. Existing Home/Devices/Overview/Panel content remains unchanged.

The supplied DeepSeek report is still unmodified. UI editing, equal authenticated access, host YAML,
plain-language Help, typed validation, private storage and protected resource contracts remain the
selected proposal. New auth/transports/semantic capabilities may still need reviewed code.

Current planning checks cover local links, typed examples, provider-schema fields, task coverage
and dependency cycles. [Audit evidence](review-evidence.json) records the results; these are not
application acceptance or security tests. Production implementation remains unstarted. OpenAPI structure/idempotency declaration is now
verified; the two requested answers are recorded as manual-first and personal Docker use. Live provider
behavior and all implementation/security/recovery tests remain unrun.


### End-to-end audit corrections

The [audit](review.md) found concrete defects in the previous design, including missing preflight
headers, unjoined lifecycle ownership, runtime temp-path fallback, absent renewal and credential
routes, lost local Save outcomes, unscoped ETags, lost source qualifications and an unsupported
representative branch. The new runtime/API contracts, typed format, example, output examples and
tasks now agree on those paths. The user answered release sequencing/deployment on 2026-10-25: manual first release, personal
Docker use. Those decisions are now fixed; technical proposals still require execution evidence. M3/M4 are later roadmap decisions, not permission to modify other views now.

Source-owned verification includes main/bootstrap startup/cleanup, API+middleware OPTIONS/conditional
mutations, repository migration/write-failure/receipt fixtures, runtime cache/renewal/generation races,
missing credentials/reauth/review, and frontend typed DTO/mock/draft/expiry behavior. Public device/MCP
negative assertions and unchanged other-page content remain mandatory. No new production code,
package dependency, schema engine, widget framework or role hierarchy was added during the audit.


### Release decisions confirmed, 2026-10-25

M1 is manual setup/refresh/Check status for personal Docker use. No recurring provider jobs,
automatic lookup continuation, new external service, widgets or other-view content changes. The
existing configs/data volumes persist definitions/runtime data; an external read-only secrets
mount supplies encryption keys. The Docker contract makes key generation/permissions/rotation/
backup verification implementation work, not an unspecified user choice. T001 is complete as a
planning decision task; all implementation tasks remain unstarted. M2 and later consumers/replacement
remain roadmap phases, outside this first-release scope.

## Follow-up design: code quality and lifecycle ownership

Reopened M1 review after user feedback; existing passing tests did not establish adequate separation
or documentation. No new capabilities, transport/storage contracts or other-view behavior are intended.

- Extract page orchestration into focused catalog/setup/detail composables. Detail presentation
  separates schedule results and settings. Precompute date labels/upcoming dates outside templates.
- Separate provider draft lifecycle, validated save/repair actions and preview execution. ProviderEditor
  composes these owners and presentation components; no component owns direct HTTP orchestration.
- Split ProviderForm into identity/input, request/capture, recipe wiring, mapping and policy sections.
  Shared controlled-form helper clones owned JSON data and emits updates; no nested prop writes.
  Pure domain helpers update request/capture references consistently and are tested independently.
- Replace the single timer/alive flag with explicitly cancellable latest-observer ownership and
  component/session-scoped requests. Guard all awaits, stop timers/listeners/frames on disposal,
  and coalesce/defer draft snapshots. Bound retained draft count, retaining current dirty work.
- Review all added Go production files. Document exported types/functions and important private
  state/JSON Pointer, cache/renewal, fencing, atomicity and encryption algorithms. Prefer named
  helpers/readable multiline code over compound statements. Fix concrete issues found with regressions.

Checks: full Go tests/build/vet; race affected packages; frontend Jest/lint/build, E2E types and
isolated Integrations browser scenarios. Add source-owned pure editing/observer/controlled-form
regressions using Vue effect scopes/fake timers (no new mounting framework). Review every NH/BE/FE
rule again; record actual results and remaining operational limits in implementation-review.md.

Follow-up completed: see implementation-review.md for source ownership, concrete defects corrected,
63 frontend suites/371 tests, 14 browser checks, Go/race/static results and CI regression details.
No persisted/wire schema or other-view behavior was changed.

## Local first-run correction (AC-31)

Bootstrap currently defaults every installation to a Docker-only /run path, leaving native runs
unavailable. Select the native default with os.UserConfigDir()/nodeherder/integration-keyring.json.
Keep explicit INTEGRATION_KEYRING_FILE read-only/required; Compose declares that explicit Docker path.
Repository-owned first creation uses random 256-bit keys, 0700 directory, 0600 temporary file,
file fsync, atomic no-overwrite hard-link publication and parent fsync. A missing key is created
only when integrations.db is absent (Lstat, including zero-length/symlink files as existing).
Existing invalid/unreadable keys and existing DBs never cause replacement. Startup losers read
only the completely published winner. No JWT reuse, new dependency or provider/device I/O.

Verify with source-owned repository/bootstrap tests: first run/restart, encrypted DB round-trip,
missing-key existing DB, invalid/explicit keys, permissions, concurrent publication and fail-closed
errors. Run full Go tests/build/vet and affected-package races; deployment config regression checks.
Update custody/ADR/setup docs. NH-01/03–06 and BE-01/03–05 apply; device semantics unchanged,
no frontend contract change. Starting the running installation is distinct from isolated testing.

## Collection card presentation correction (AC-32)

The global .nh-card + .nh-card margin shifts later grid siblings, and independently sized headers
misalign date content. Override the sibling margin only in the waste collection grid and use CSS
row subgrid to share header/body tracks (scoped fallback for older browsers). Keep shared UiCard
behavior unchanged. Add one presentational WasteBinIcon.vue with static SVG geometry and a finite
pure colour palette; no provider-supplied style/markup, image dependency or external asset request.
Use its existing tools slot for the illustration. Test palette/unknown fallbacks and an isolated
three-card browser fixture with short/long titles at desktop/mobile widths. Run frontend unit,
lint/build, E2E types and Integrations browser checks. NH/FE boundaries and other views are unchanged.

AC-33 refines AC-32 after the user's visual feedback: replace the tall status paragraph block with
a compact native details/summary row, move recent checks below results, reduce header icon/spacing,
and retain colour names in headers without repeated colour/coverage metadata below the date. Keep
source descriptions/verification/refresh notes under Details instead of deleting information. Use
small flat SVG geometry matching UI icon styling; no image network dependency or provider HTML.
Extend the existing isolated browser layout test for compact mobile bounds/disclosure accessibility,
light/dark rendering, and shared desktop tracks; keep the actual application on its own ports.

## Approved Integrations board implementation (AC-34–37, 2026-10-26)

User approved the recommended grouped Connection board and the shared wheelie-bin SVG treatment.
The [board contract](contracts/connections-board.md) supersedes pending design/resize/permission
questions in the investigation. Freeze shared order+visibility (no size) and separate hide/delete.
Add optional per-instance preferences with independent revisions; atomically expose a protected
board GET/PUT, without changing encrypted schema/AAD. This is smaller than a separate widget DB,
while avoiding private Home/device settings or browser persistence. Follow existing boundaries.

Backend owners: models/integrations/state.go (optional preference), internal/integrations/board.go
(layout domain/transaction), internal/api/integrations.go (transport/auth). Frontend owners:
types/contracts/service/store (accepted board), domain/connection-board.ts (pure layout/summary),
useConnectionBoardLayout (draft), useConnectionsBoard (page lifecycle), useConnectionsCatalog
(Add), ConnectionWidget/ConnectionBoard/ConnectionCatalogDialog (presentation), ConnectionsPage
(coordinator). Reuse WasteBinIcon, UiDialog/UiButton/UiIcon and installed vuedraggable. Preserve
all existing detail/provider/help views. Setup receives only a named return intent, not a URL.
No new dependency, automatic upstream polling, device feature, public route or role tier.

Proceed T041–45 in order. Checks/ownership and NH/BE/FE review are in the contract. New acceptance
requires actual application checks; prior prototype passes do not count as implementation evidence.

## Navigation/header polish — AC-38–39

Investigate with stalled isolated read responses before editing. Source hypotheses: ConnectionFeedback
conditionally inserts a Working paragraph into page flow; useConnectionDetail initially ignores the
validated instance already on the board and sequentially toggles busy for instance/provider reads.
A generic Connection title gains a postcode after the first read, changing the header geometry.
Confirm with title text/geometry probes on initial detail entry and back; do not add a transition
animation to disguise movement.

Use matching private session cache for immediate detail presentation while revalidating; keep
request/session guards and explicit not-found handling. Keep initial instance/provider work in one
owned busy interval. Opt board/detail into non-flow accessible loading feedback, preserving other
callers' current feedback. Keep the page aria-busy truthful and use an actual cold-load message when
no cached data exists. Compare no header image (more compact, weaker connection identity) with the
existing generic schedule/calendar icon (consistent Home-like identity, no duplicate bin); use the
latter without a new asset or dependency.

Owners: ConnectionFeedback.vue, useConnectionDetail.ts, ConnectionsPage.vue, ConnectionDetail.vue,
ConnectionWidget.vue and existing ConnectionsPage.spec.ts. NH-01/04 and FE-01 retain presentation/
composable boundaries; NH-02 has no device changes; NH-03/FE-02/05 retain validated private cache,
auth and cleanup; NH-05/FE-04 require measured delayed-navigation browser regressions; NH-06/FE-03
retain errors, missing/deleted states, no new timers/network retries. No backend/wire change.
Run frontend unit/lint/build/E2E types and isolated Integrations checks; no live household fixtures.

## Chronological collections for consumers — AC-40

Current UI maps provider service order directly into both views. Add a shared nullable ISO-date
comparator and raw next-date presentation key in domain/waste-schedule.ts; sort newly projected
cards and their copied upcoming-date arrays. connection-board.ts sorts flattened summary rows with
the same comparator before taking three rows. Same-date/no-date ties preserve source order. This
does not change resource identity, JSON fields or the user's connection layout. The backend
projection below guarantees chronological ordering to all authenticated resource consumers too.

Extend existing source companions waste-schedule.test.ts and connection-board.test.ts for shuffled
provider dates, cross-month/year ordering, ties, undated/exhausted services, UK rollover, no mutation
and multiple-resource widget truncation; add an existing ConnectionsPage.spec.ts regression for
both rendered views. Run unit/lint/build/E2E types and isolated browser tests. NH-01/04/FE-01 keep
this shared pure logic out of components; NH-02/03/FE-02/05 boundaries are unchanged; NH-05/FE-04
require ordering evidence; NH-06/FE-03 preserve UK dates and missing-versus-empty semantics. No
timer, cache lifetime or other-page changes; only the backend domain projection is extended below.

AC-40 clarification after user review: sort the backend's canonical returned resources too, not
only the UI projection. WasteSchedule.Project (models/integrations/resource.go) is already called
by normalization and authenticated instance/resource reads. Clone/sort Services there by the first
canonical date >= the UK local date, keeping missing-date/tie behavior stable. Preserve revisions,
complete date arrays and snapshot identity. Cache immutable London timezone data with sync.OnceValues
and embed the standard-library tzdata fallback so native/container sorting does not depend on an
external zoneinfo install; no new third-party dependency. Unsupported/unresolvable zones retain
source order rather than guessing a local date; v1 validates Europe/London.

Extend resource_test.go (chronology, ties, no mutation and UK-midnight/DST reordering) and existing
service_test.go (provider normalization plus instance/list consumers). Contract: services array is
ordered as of evaluated_at; dates within services remain complete sorted-unique lists. No JSON
fields/route/auth/storage change. NH-03/BE-05/FE-02 preserve private API fields; BE-01 owns this in
the domain, BE-03 gets clone/race checks, BE-04 defines date/order semantics with old snapshots read
without migration. Run Go suite/build/vet/affected races in addition to the frontend checks.

## Calendar ownership and clean presentation — AC-41–42

Repo audit: Compose sets process TZ; frontend mock server has its own TIMEZONE; existing automation
scheduler hardcodes London; MCP accepts a timezone option/default UTC; checked timestamps currently
use browser timezone. There is no shared runtime General timezone setting. Do not repurpose process
TZ as the date semantics for provider data. General timezone remains a separate feature requiring
settings persistence, client/server consumers, DST rules and existing-schedule migration decisions.

Centralize UK waste contract values in backend/models/integrations/calendar.go and frontend/src/config/
integrations.ts; use the frontend constant in contract/type declarations and schedule.timezone in
calendar formatting. Reuse the backend's cached immutable location for projection and horizon checks;
normalization uses the shared contract constant. No changes to automation/device behavior or v1 wire.

Refactor frontend/domain/waste-schedule.ts into a typed schedule composer and small card/label helpers;
extract pure calendar/date formatting and chronology to waste-calendar.ts. Keep the comparator export
compatible. Refactor touched connection-board.ts row/status projection into named helpers and typed
models. No generic framework, new dependencies or cosmetic wrapper layers. Preserve public output
fields and nullable meanings; do not move transport/persistence into presentation.

Update existing Vue/Go rules with mandatory JSDoc/GoDoc for new/changed public APIs and purposeful
comments for complex flow/ownership/units, plus review checks. Keep guidance concise and AGENTS.md
unchanged. New source companion ownership: waste-calendar.ts→domain/waste-calendar.test.ts;
calendar.go→calendar_test.go. Extend existing schedule/board/contract companions and retain all UI
regressions. Run frontend unit/lint/build/E2E types/browser, Go suite/build/vet/affected races. NH-01/
04 and FE/BE-01 enforce ownership; NH-02/device pipeline unchanged; NH-03/FE-02/05 contracts/private
state unchanged; NH-05/FE-04 have regression evidence; NH-06/BE-03–04/FE-03 preserve copy, date and
cleanup behavior. Local user automation config and already-staged changes remain untouched.
