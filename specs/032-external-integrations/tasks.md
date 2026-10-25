# Tasks: Configurable external integrations

Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md)
Status: M1 implemented and verified locally, 2026-10-25; M2–M4 deferred

M1 is authorized by the user. This is a dependency-ordered backlog (stable IDs are not execution order);
M2 and later roadmap phases remain outside the first release. Stable task IDs refer to the existing
spec's acceptance IDs. Each test file must have a real production owner and use its name;
frontend unit files mirror source paths under src/__tests__/. M1 source owners and actual verification are recorded in [implementation-review.md](implementation-review.md). Later milestone paths remain proposals. Checkmarks apply only to the M1 scope of cross-milestone tasks.

## M0 — Decisions and evidence

- [x] T001 — Record release/deployment and shared-access/configuration decisions in spec/plan;
  IDs: FR-03/05/08/10, AC-04–05/07/11/15–17/20.
  Depends: none. Evidence: user confirmed on 2026-10-25 manual refresh for the first release and
  personal Docker use; prior UI/YAML and equal-access decisions retained. M1 excludes recurring
  provider work; Docker key mount/backup design is concrete. This completes planning only.
- [x] T002 — Verify the confirmed personal-Docker/equal-access contract, D-04 host keyring/mount and
  proposed ADR/contracts; IDs: AC-13–15/18–19, NH-03/04/06. Depends: T001.
  Done: deployment/secret/backup feasibility and actual review recorded; tenant scope explicit.
- [x] T003 — Verify intended provider access/volume, current schemas, required inputs, idempotency
  retention/window and exact-only initial personal-use journey; paths: provider fixtures and this planning set;
  IDs: AC-01/03–04/10/17–18. Depends: none for read-only research, T001–02 for contract freeze.
  Done: parsed OpenAPI 1.1.0 evidence/hash is recorded in review.md; complete personal-use quota/token/error fixtures,
  full/partial/error fixture extraction and allowance discrepancy review, then re-estimate work
  packages. Later consumer scenarios are M3 gates, not a blocker for M1. No routine live POST/device tests.

## M1a — Secure manual backend

- [x] T004 — Implement typed models and waste normalizer/projections from the contract;
  paths: backend/models/integrations/, internal/integrations/capabilities/waste.go and waste_test.go;
  IDs: AC-03/08–09/18. Depends: T002–03. Done: real dates/DST, unknown enum labels, null/empty,
  coverage, notes/subscription, unreturned services, descriptor keys/rename limitations, provisional/
  partial rejection/qualification and duplicate-key fixtures pass against updated wire examples.
- [x] T005 — Implement repository migrations, atomic identity/binding/compiled-revision/snapshot operations;
  paths: backend/repository/integrations.go and integrations_test.go; IDs: AC-07/09/11–12/19.
  Depends: T004. Done: strict data path/no temp fallback, old/newer/corrupt schema migration fixtures, setup completion
  receipts/uniqueness, stable provider_binding_id across audit revisions, temp-DB restart,
  collision/rollback/write-failure/mapping fixtures; failed
  writes do not publish new revision/ETag/checkpoint. M4 replacement-specific paths remain gated.
- [x] T006 — Implement keyring/private-record encryption, rotation and backup/restore contract;
  paths: backend/repository/integrations.go and integrations_test.go; IDs: AC-14/19.
  Depends: T002/05. Done: mixed old/new key ciphertext, interrupted rotation, missing key,
  encrypted snapshots/provider notices, protected record context and isolated restore evidence; unrelated device paths remain available.
- [x] T007 — Implement feature authorization and cookie mutation checks using verified claims;
  paths: internal/api/integrations.go and integrations_test.go, relevant existing middleware
  companions if changed; IDs: AC-09/13/15. Depends: T002.
  Done: two authenticated users share full controls/setup, anonymous denied, no role/allowlist
  config required, stale-revision conflicts, missing/invalid Origin/header and cookie-plus-bearer; direct handler rejection plus integration-only If-Match/If-None-Match/custom-header preflight and
  exposed headers. Write-only setup/instance credential routes, required revisions, expired setups,
  origin/credential destination acknowledgements and definition_in_use deletion return the exact envelope/HTTP codes.
- [x] T008 — Implement outbound policy/client and admission primitives;
  paths: internal/integrations/http.go and http_test.go; IDs: AC-06/13/17.
  Depends: T002–03. Done: httptest/dial fakes cover IPv4/IPv6/private/link-local/metadata, DNS
  rebinding, redirects/proxy, public-host credential exfiltration, timeout/decoded-size/quota bounds.
- [x] T009 — Implement bounded YAML definition grammar/compilation using the exact contracts;
  paths: internal/integrations/configuration.go and configuration_test.go; IDs: AC-11/13/17/20.
  Depends: T001–03/05/08/27. Done: strict YAML/unknown-field/duplicate-key/tag/alias/size checks,
  Form/YAML parity, JSON Pointer/typed references, status/header/304 parsing,
  allowed endpoint/mapping repair, rejected scripts/new transforms/secret refs, origin reassociation,
  phase-availability/cyclic ref checks, generic renewal/problem classification, profile pinning,
  optimistic activation, validator invalidation, live-test isolation, validation failure and rollback fixtures.
- [x] T010 — Implement shared revision-scoped setup, exact attempt body/key and manual pending continuation;
  paths: internal/integrations/setup.go (tested in service_test.go/coordinator_test.go),
  operations/request.go and request_test.go; IDs: AC-01–04/06–07/17–19/22. Depends: T004–09/27.
  Done: candidate tampering/postcode change, response-loss recovery inside verified window,
  expiry/reconciliation, p:-token transform, exact-only required-input discriminator, candidate edit
  isolation, typed partial data and provisional rejection. Repeat local Save returns its claimed
  instance/operation; no new POST key, stale selection or cross-property publication.
- [x] T011 — Implement lifecycle/manual execution, local cleanup and generation fencing;
  paths: internal/integrations/service.go/service_test.go, janitor.go/janitor_test.go,
  coordinator.go/coordinator_test.go, bootstrap.go/bootstrap_test.go and main.go/main_test.go; IDs: AC-04/06–08/13/19. Depends: T005–10.
  Done: optional degraded bootstrap, run/close ownership after startup/listener/signal errors,
  join-before-DB-close and strict paths; bounded refresh/renewal/check-status, saved inputs/new keys,
  minimum intervals/quotas, request-scoped ETags/one missing-payload unconditional retry, manual
  restart continuation, no recurring M1 provider calls, coalescing, cancel/expiry/orphan cleanup,
  disable/delete/config/credential fencing, disconnect-after-admission and deadline-exceeded close
  behavior. Feature context owns workers, not r.Context. Publish nothing on failed persistence.
- [x] T012 — Wire protected catalog/setup/definition/instance/resource APIs;
  paths: internal/hub.go and existing hub_test.go, internal/api/integrations.go and
  integrations_test.go, optional feature bootstrap/main composition; preserve AppStore mocks and fatal/nonfatal boundaries;
  IDs: AC-01–04/06–09/13/15/17–19. Depends: T007/09–11/25.
  Done: real router producer/consumer DTO tests, PUT support, no upstream call on GET, key failures
  isolated, public /api/mcp and /api/context/devices negative assertions, no integration data or resource references in settings WS payloads.

- [x] T025 — Implement host YAML repository/save/journal/reload and deployment persistence;
  paths: backend/repository/integration_definitions.go and integration_definitions_test.go,
  integration configuration/service companions, deployment/ignore docs when implemented;
  IDs: FR-10, AC-15–17/20. Depends: T005/09. Done: fixed-directory IDs/symlink rejection,
  atomic rename/hash checks, read-only/disk-full/activation-failure recovery, startup reconciliation,
  embedded seed from backend/configs/integrations/embed.go and source-owned embed_test.go,
  binary/container default availability, exclusive default-copy/no overwrite, host Reload and local history; no temporary-directory fallback or Git
  actions. Backups restore YAML+DB+separate key, and private runtime files are excluded from Git.

## M1b — Manual UI and repairs

- [x] T013 — Implement frontend types/runtime parsers/service/Vuex state and mutation header;
  paths: src/types/integrations.type.ts, contracts/integrations.ts, services/integrations.service.ts,
  store/modules/integrations/index.ts and mirrored companion tests; IDs: AC-01–04/08–09/15–20.
  Depends: T004/07/12. Done: exact DTO/mock parity, malformed/unknown schemas rejected, no browser
  PII persistence, local 60-second/on-focus projection refresh without upstream calls, no timer leaks.
- [x] T014 — Implement catalog/wizard/detail/health and Form/YAML definition editor;
  paths: src/components/integrations/, existing router/navigation owners and their appropriate tests;
  IDs: AC-01–04/06/08/15–18/20. Depends: T013.
  Done: equal controls for all signed-in users, editable Form/YAML parity, request/extractor edits,
  offline/redacted live previews, Save/reload/restore and destination reconnect, shared revision
  conflict feedback, pending Check status, exact-only unsupported input, 410 expiry/logout cleanup, credential create/
  replace/remove and shared-origin review, setup receipt recovery, same/different-property edit
  feedback, no hidden token downgrade, isolated live/offline test feedback and mobile repair.
- [x] T015 — Complete M1 acceptance with isolated browser fixtures and backend checks;
  paths: frontend/e2e/ using established suite convention and all changed source companions;
  IDs: M1 acceptance set in spec, NH-05, FE-01–05. Depends: T012–14/26–28.
  Done: verify isolated mock identity/ownership before mutations, disable Node/Vite server reuse
  against occupied ports and guard E2E_EXTERNAL_SERVERS. Update source-owned test harness files
  (playwright.config.ts, tools/server/routes.mjs/config.mjs and e2e/support helpers) with exact DTO
  fixture parity; no production test-identity route. Then keyboard/focus, 390px/1280px,
  errors/offline/login denial, retention/key recovery/host
  deployment docs and recorded tests/lint/build/vet/races. No claim of M2 automatic completion.

- [x] T028 — Expose the top-level Connections category and type/provider/instance navigation;
  paths: existing shell/navigation.ts, shell/AppShell.vue, types/router.ts, router/index.ts,
  new integrations catalog/detail owners and corresponding source-named unit companions;
  IDs: FR-13/16/17, AC-16/23/26/27. Depends: T012–14.
  Done: all existing sections retained, Connections after Devices in desktop/mobile, auth/deep-link/
  active-section/not-found handling, type/provider distinctions, device counts unchanged and no
  device prerequisite. Only installed types have setup actions; shared-provider edit scope visible. Connections-only
  UI: preserve Home/Devices/Overview/Panel content/counts/layout and shared auth routes; no pin,
  room placement, cross-promotion links/buttons/notices or global footer copy. Negative browser
  checks cover existing pages and the isolated fixture baseline of 18 Home tiles (the supplied household screenshot showed 15). No dashboard settings/WS changes.

Retired tasks: **T029** (Home resource tiles/settings) and **T030** (Panel connection tiles/details)
are withdrawn by the Connections-only scope decision. Do not implement their proposed schemas,
renderers or dashboard persistence changes. These IDs remain reserved and are not dependencies.

## M2 — Automatic lifecycle

- [ ] T016 — Add one durable due-work coordinator using the same execution path;
  paths: internal/integrations/coordinator.go and coordinator_test.go; IDs: AC-05–08.
  Depends: T015. Done: fake-clock retry-at/global-provider quota/jitter/pending restart tests,
  per-service horizon exhaustion, projection deadlines/DST and shutdown/race evidence.
- [ ] T017 — Add enabled schedule settings/health UI and contract parity;
  paths: service.go and companion, API/frontend affected owner companions; IDs: AC-05–08/16.
  Depends: T016. Done: browser-closed automatic updates, requested versus confirmed enable state,
  stale source distinct from 304/check time, no hidden provider calls in reads.

## M3 — Actual consumers

- [ ] T018 — Specify and implement an explicitly authenticated read-only assistant/MCP consumer;
  paths: existing internal/mcp/server and transport/HTTP owners plus companions, new capability read
  adapter if justified; IDs: AC-09–10/13/15. Depends: T003/15, T017 for automatic freshness scenarios.
  Done: producer/consumer contract and public-route negatives, no secret/PII leak or action tool.
- [ ] T019 — Add one scoped integration automation/recipe path;
  paths: internal/automations/ engine/context/trigger/condition owners and companions, corresponding
  frontend recipe owners; IDs: AC-08/10, NH-02. Depends: T003/17.
  Done: resource/service IDs, fresh exact occurrence eligibility, subscription handling, dedupe,
  no provisional/cache-restore replay, existing recipe/device event regression evidence.

## M4 — Replacement and extension proof

- [ ] T020 — Add a second real provider binding and private replacement preview/reconciliation;
  paths: providers/new-source owning files/companions, replacement.go/replacement_test.go and
  repository/service companions; IDs: AC-11–12. Depends: T015/19; provider schema/access evidence.
  Done: same-property confirmation, ambiguous service mapping rejected, source rename/restart IDs
  stable, candidate/active isolation, generation fencing and atomic failed-write recovery.
- [ ] T021 — Implement replacement review UI, previous-binding retention and explicit rollback;
  paths: frontend replacement service/domain/UI owners and mirrored companions, API/repository
  owner companions; IDs: AC-11–12/16/19. Depends: T020.
  Done: old/new differences and affected recipes visible, unmapped semantics require review,
  IDs retained only for confirmed mappings, seven-day cleanup, no dual refresh or event replay.

## Handoff gates

- [x] T026 — Add the Configure operations Help tab/child page and contextual help links;
  paths: frontend/src/components/integrations/IntegrationsHelp.vue, pure help-metadata helper and
  its mirrored companion, existing router owner and isolated Playwright flows; IDs: FR-11, AC-16/21.
  Depends: T013–14/27. Done: quick start, files/save, inputs/state, request wiring, schemas/mapping,
  URL/field repairs and errors use the shipped grammar; field/error anchors, local search, offline
  content, keyboard/focus and 390px/1280px navigation preserve unsaved editor buffers. Human-readable topics
  explain tasks before syntax; parameter key/source/fixed-value controls round-trip through YAML,
  Help preserves drafts, and automatically saved values are text rather than false editable inputs. Feature-owned draft/
  scroll restoration survives AppShell routing without changing other views; dirty drafts are not
  overwritten by refetch, and logout/401 clears private buffers. Test reserved reference names.
- [x] T027 — Implement the small typed definition model/checks and generic request/recipe references;
  paths: contracts/definition-format.md, backend/models/integrations/definition.go and
  definition_test.go, backend/internal/integrations/operations/
  request.go/request_test.go, setup.go/coordinator.go and common configuration
  companions; IDs: FR-12, AC-17/22. Depends: T002–05/08.
  Done: typed example + cross-reference fixtures, unsupported version/type rejection, request ID
  rename without Go changes, phase/ref availability, immediate/complete capture ordering, named renewal/error rules, one mapping for all bin colours,
  and provider field movement repair. No provider-specific request-name switch or global scripting
  runtime or JSON Schema engine; UI/help use the same documented format/version. Remaining fixtures recorded honestly.

- [x] T022 — Run required suites/builds/lint/vet/races and selected isolated browser checks for
  each completed milestone; paths/commands: plan section 13; IDs: NH-05, relevant BE/FE checks.
  Depends: each milestone's implementation. Done: exact commands/results, actual failures/unrun
  reasons, source ownership, mock parity and acceptance evidence recorded without live devices.
  Do not run npm run test:e2e until owned Node/Go fixtures are proven; never reuse a merely
  responsive default backend port as a mock.
- [x] T023 — Recheck every NH/BE/FE rule and planned ADR against final scope;
  paths: this same spec/plan/tasks/contracts set; IDs: NH-01/04/06 and all applicable rules.
  Depends: T022 for the release milestone. Done: real reviewer decisions, compatibility/restore
  evidence and deployment docs; no invented approvals or constitutional ratification.
- [x] T024 — Before any commit/handoff, verify all local planning links/artifacts are included;
  paths: this directory and referenced docs/external-integrations-investigation.md; gate: NH-04/05.
  Depends: none for planning, T023 for release. Done: no link depends on an omitted untracked file;
  preserve the other report or replace its link with an agreed durable reference.

Audit source: [review.md](review.md), [runtime lifecycle](contracts/runtime-lifecycle.md) and
[API/recovery](contracts/api-and-recovery.md). The documented provider contract was inspected;
M1 checkboxes are complete; later milestones remain open. M3/M4 require their own scope decisions before work.

## Implementation log

2026-10-25: Read AGENTS.md, component rules and SDD workflow; implementation authorized for M1 on
spec/032-connections. No implementation checks are marked passed until their actual evidence exists.

2026-10-25: M1 implemented and locally verified. See implementation-review.md for source-owned
acceptance evidence, concrete corrections, exact check results and operational limitations.
T003 completes documented-contract/fixture verification only: live account quotas, council coverage
and provider availability remain operational checks. T022/T023 apply to M1, not M2–M4.

## Reopened follow-up: maintainability review

- [x] T031 — Refactor controlled provider forms and pure request/capture edits; AC-28; source owners:
  ProviderForm/section components, useEditableModel and domain/provider-definition companions.
- [x] T032 — Separate catalog/setup/detail and provider editor draft/actions/preview ownership;
  AC-28/29; source owners: feature composables, presentational results/settings/editor sections.
- [x] T033 — Fix/test superseded observers, stale responses, timer/listener/frame cleanup and draft
  retention/cloning; AC-29; source owners: operation-observer, feature composables/domain companions.
- [x] T034 — Review/document all newly added Go production APIs/types and difficult algorithms;
  AC-30; source-owned regressions for any logic changes. Preserve persisted/wire semantics.
- [x] T035 — Re-run required checks; review SOLID/DRY, performance/cleanup and NH/BE/FE evidence;
  update implementation review and existing PR. Depends T031–34. No live devices/provider jobs.

2026-10-25 follow-up: T031–T035 completed. Responsibilities, documentation and lifecycle fixes
are recorded in implementation-review.md. Local checks: full Go suite/build/vet, affected-package
races, 371 frontend tests, build/lint/E2E types, 14 isolated browser checks and 13 Compose config
checks passed. Original PR CI exposed a focus timing race and Compose false-serialization test
assumption; both are corrected with local regression evidence. Updated remote CI is a separate result.

- [x] T036 — AC-31 local key selection and repository-owned no-overwrite first creation; companions:
  repository/integration_keyring_test.go and internal/integrations/bootstrap_test.go. Update Docker
  explicit path and custody/setup docs; verify first-run/restart/lost key/invalid key/concurrency,
  full Go/build/vet/races and isolated deployment configuration; record actual rollout status.

AC-31 verified locally: native first run/restart, encrypted data round-trip, lost-key preservation,
malformed/public keys, private directory policy and 12 concurrent creators pass. Full Go suite,
build/vet, repository/integration races and 13 isolated Compose config checks passed. Tests used
temporary HOME/XDG/data/config paths; the user's running backend was not restarted.

- [x] T037 — AC-32 align collection grid/header/date tracks; reusable colour-bin SVG and safe palette;
  source owners WasteScheduleResults.vue, WasteBinIcon.vue, domain/waste-bin.ts and companion;
  isolated ConnectionsPage browser checks cover three unequal names, colours and mobile overflow.
  Run frontend checks and record visual review; no live devices/provider calls.
- [x] T038 — AC-33 compact status disclosure, collection heading/icon/metadata hierarchy and recent
  checks placement. Owners WasteScheduleResults.vue, ConnectionDetail.vue, domain/waste-schedule.ts,
  WasteBinIcon.vue and existing companions/E2E. Record mobile/light/dark visual and frontend checks.

AC-32/33: final medium 48×60 SVG icons, 70–88px header bounds, compact <=48px status row,
colour-name headings without repeated metadata, hidden source/coverage details and later Recent checks verified.
380 unit tests, frontend build/lint/E2E types and 16 isolated browser checks passed. Desktop/mobile
light/dark screenshots reviewed. Tests used a temporary source snapshot on alternate ports because
the user's local app owns the standard ports; no running app/device/provider mutations were made.
