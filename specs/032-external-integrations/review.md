# Connections plan audit

Reviewed 2026-10-24 against `986b923c` on main. Planning/design only; no application implementation,
provider lookup job or household command was run. Planning checks do not establish backend correctness. This audit corrects design defects rather than claiming bug-free code.

## Verdict and decisions

The Connections-only boundary is consistent: one menu entry, no changes to existing page content,
dashboard schema, room/Panel tiles or global promotional text. Architecture remains bounded YAML
plus typed Go operations/capabilities, not a workflow/scripting framework. The earlier design was
not implementation-ready: several required execution/transport paths were described but unowned
or absent. The contracts below now give those paths concrete behavior and source owners.

The user answered the two requested decisions on 2026-10-25: **manual refresh for the first
release; personal use on Docker**. No delivery/deployment answer remains pending. The selected
keyring design now uses a separate read-only Docker secrets-directory mount. Later assistant/
automation consumers are separate scope decisions, not a reason to modify other views during M1.

## Findings and corrections

| Priority | Finding and code evidence | Correction / owned verification |
| --- | --- | --- |
| High | Current CORS in backend/internal/api/middleware.go does not allow conditional-save headers. The plan mentioned its custom header but omitted If-Match/If-None-Match, so cross-origin UI saves could fail before reaching handlers. | Integration-path-only preflight/exposure contract and actual-request cookie checks; middleware_test.go and integrations_test.go include real OPTIONS + PUT/DELETE flows. |
| High | main.go exits on startup errors, ApiServer.Listen uses context.AfterFunc without joining feature workers, and AppStore has no Close. An unjoined DB close callback can be lost on process exit. | One optional feature composition owner, run/cleanup boundary, bounded idempotent join-before-close and failure/listener/signal fixtures. No fatal device-store dependency. |
| High | GetDataDir/EnsureDir and the generic KV constructor can silently choose a temporary directory. Protecting YAML alone left runtime state/credentials vulnerable to disappearing between restarts. | Strict integration DB/key/config paths, bounded lock wait, transactional version migrations and degraded feature status for invalid/missing/locked storage; no fresh replacement DB/key. |
| High | refresh.request only read the stored schedule; no declarative renewal path or distinction between no_schedule and a broken endpoint existed. Terminal lookup polls cannot create a new schedule. | Optional renewal_request plus bounded status/problem actions, saved encrypted confirmed inputs, one durable new logical key, cooldown/backoff and explicit re-setup when identity/input is invalid. Unknown 404 never triggers POST. |
| High | Provider metadata mapping omitted subscription qualifiers/notes. Notes may contain private addresses even after matched_address is discarded. | Map subscription_required/notes; missing subscription is unknown. Encrypt all snapshots, preserve bounded caveats as text and keep them out of logs/public APIs/exports. Typed examples now include notices. |
| High | A single reserved ETag did not distinguish request/definition/binding/credential versions. A 304 after config repair or cache loss could incorrectly reuse old mapped data. | Request/revision-scoped validator keys, invalidation on activation/credentials, atomic validator publication, one bounded unconditional recovery if payload is missing. 304 never refreshes source age. |
| High | Credential provisioning/removal, destination reassociation and definition deletion were required but their control paths/DTOs were unspecified. | Write-only setup/instance credential routes, explicit removal, explicit revision-bound Save acknowledgement, definition_in_use rejection and safe delete/disable semantics. No new role. |
| High | External POST idempotency did not protect a repeated local Save after response loss. Setup cleanup could erase the only clue that an instance was already created. | Claim setup once before provider work, preserve a minimal encrypted completion receipt, return existing operation on an identical retry and enforce definition/property/capability uniqueness. |
| Medium/high | M1 promised representative consent/display while its only recipe supported exact property IDs. required_input and result identity weren't explicit compile/ingress guards. | M1 is exact-choice only; declare search discriminator, reject unsupported journeys and provisional/non-exact snapshots. Future representative support needs a real recipe and consent tests. |
| High | The identity index included binding_revision, which could regenerate service UUIDs on ordinary binding/config revisions despite unchanged property/services. | Use stable provider_binding_id in the index; retain it across ordinary revisions and test service-ID continuity. Different bindings require explicit mapping. |
| Medium | Sources without stable service IDs cannot preserve identity through arbitrary display-name changes automatically. The previous adapter-local-key language left implementers to invent incompatible algorithms. | Provider key when present, otherwise persisted descriptor:v1 key over an exact name/type/container tuple; no colour/index matching, no automatic rename claim, preserve unreturned services as unavailable and allow later reviewed mapping. |
| High | Runtime image copies only the binary (backend/Dockerfile), so definitions stored only as source/spec examples would be missing from a binary install. | Embed the one reviewed YAML seed from backend/configs/integrations; exclusive first-copy, never overwrite host files, and source-owned embed_test.go fixture parity. |
| Medium | AppShell resets scroll on fullPath navigation; a sibling Help route can lose scroll or reload over an unsaved provider draft. | Feature-owned memory draft/revision and restoration after route activation; logout clears all private buffers; no global shell behavior rewrite. |
| Medium | New route/auth/expiry, draft transitions, error envelopes and file preconditions lacked exact shared behavior. | New API/recovery contract fixes scoped ETags, 428/409/410/503 outcomes, pending resources, instance candidate edits, secret-buffer cleanup and GET-without-provider-I/O. |
| Medium | Runtime/definition/credential revisions can race, live tests can accidentally publish, and deleting local work cannot cancel an admitted provider job. | Fence every transaction and publication by generation and credential/grant revision; live tests isolated and ephemeral; local cancellation described honestly. |

The [runtime contract](contracts/runtime-lifecycle.md), [API/recovery contract](contracts/api-and-recovery.md),
[typed format](contracts/definition-format.md), [configuration contract](contracts/configuration.md)
and [waste contract](contracts/waste_schedule.v1.md) are the corrected normative design. Spec/plan/tasks
must point to them and the same source-owned checks; no second scheduler, broad role system, new
public endpoint or arbitrary provider-code runtime was introduced.

## Provider evidence

Public OpenAPI downloaded by documentation-only GET and parsed locally: **1.1.0**, SHA-256
`b8ece29265c75e7e5d5ec1ed12c4d5b284117dfc5bd04ed0a83d3b53dc7771e4`.
The specification declares a 30-day-or-removal idempotency lifetime, separate lookup/schedule
responses, optional provisional/subscription fields, and a schedule redirect response. Those details
inform fixtures; they do not prove live availability or exactly-once behavior. [OpenAPI](https://whenisbins.com/v1/openapi.yaml)

The guide describes a minimum status interval, renewal when usable dates are exhausted, and
regular-use free tokens. Its allowance table and OpenAPI overview differ, so do not hard-code the
more generous claim or size production solely from it. Honor runtime Retry-After and use conservative
limits for the confirmed personal deployment while resolving the provider documentation discrepancy. [Guide](https://whenisbins.com/v1/guide)
Commercial access needs operator contact; no email was sent. [Terms](https://whenisbins.com/termsandconditions)

The YAML example is still proposed NodeHerder grammar, not an installed/live-tested integration.
Positive provider-shape inspection does not cover every council journey, semantic qualification,
rate-limit branch or failure. All routine implementation tests use saved/synthetic shapes and fakes.

## Rules and verification limits

- NH-01/04: same spec/plan/tasks set; separate user decisions from technical proposals and retired
  dashboard experiments. No approval, exception or constitutional ratification is invented.
- NH-02/03; BE-02/05; FE-05: no device impersonation or new public MCP/device data; all mutations
  authenticate and enforce cookie origin rules; credentials/provider text stay private.
- NH-05/06; BE-03/04: meaningful failure fixtures, schema/file recovery, bounded work and joined
  lifecycle are now assigned. Planned tests are not passed tests.
- BE-01; FE-01/02/03/04: domain/repository/controller/Vuex boundaries, runtime ingress, exact DTO
  mocks, meaningful pending/stale states and accessible contained UI remain required.

New APIs/crypto/storage/HTTP are unimplemented. Go/Vue application suites, lint/build/race, live
provider jobs and restore exercises were not run; implementation compliance cannot yet be certified.
Static document/provider-fixture/reference and task-graph checks are recorded after the second pass.
No reviewer can prove absence of all future implementation bugs from a plan.

## Additional second-pass findings

The binary/default-asset and draft/scroll issues above were caught in the second pass. It also
identified an execution-context trap: a worker inheriting r.Context would stop when its accepted HTTP
response ends. Admitted work now explicitly uses the feature-owned context; shutdown never closes
a DB under still-running writers after a timeout.

The existing frontend/playwright.config.ts permits reusing the Node mock server on port 4110, which
is also the normal backend port. Running the broad browser suite without fixture identity checks
could attach to a live installation. No such suite was run. T015/T022 must make mock reuse false,
verify fixture identity/ownership before mutations, and validate any E2E_EXTERNAL_SERVERS override.
The existing isolated Go browser hub has its own health path; the Node fixture needs a corresponding
read-only identity marker in its test-only routes. This is test-harness scope, not a production API.

The final simplification removes an unnecessary signed origin-review ticket: existing revision
preconditions plus a separate structured destination acknowledgement provide the required explicit
user action. No extra signing key/service is needed. Generic config cannot prove every manually
authored parameter semantically safe; authors are trusted, while output/exactness/credential-origin
checks remain deterministic. The shipped M1 flow exposes no representative-consent branch.


## Final verification result

[Machine-readable planning evidence](review-evidence.json) records **26 passed document/provider-shape
checks**: local links/fences/whitespace/JSON, all 25 active acceptance IDs covered by tasks, all 28
active tasks with no explicit dependency cycle (T001 decisions complete; implementation unstarted), all NH/BE/FE rule rows, downloaded provider version/
hash, example field pointers/body inputs against the provider schema, request-renaming and bad-ref
cases, and all five output examples' dates/freshness/identity/notices semantics. These are structural
checks, not tests of an implemented Go executor. Visual exploration files were removed after review at the user's request; no browser test or backend
correctness result is claimed by the retained planning artifacts.

No tracked production files changed. Application tests/build/lint/race and restore/real-lookup tests
remain unrun. The review is complete as a planning audit; sequencing and personal-Docker use are now directly confirmed. Remaining fixture, deployment and
recovery checks are implementation work, not unanswered scope questions.
Provider behavior/allowance disagreement is documented for fixtures and conservative runtime policy;
regular-use token handling remains conditional on actual usage; no commercial deployment is
assumed or required for this first release.


Decision update 2026-10-25: T001 is complete (planning only). Personal Docker persistence, key mount,
rotation/restore paths are documented; no live key/container was created and no production code changed.
