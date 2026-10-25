# Connections M1: implementation and verification

Date: 2026-10-25 | Scope: manual-first, personal Docker, Connections-only UI
Status: implemented and verified locally. This is an agent self-review, not maintainer approval,
constitutional ratification, a live-provider certification or a deployment record.

## Delivered behavior

Connections appears after Devices in the existing desktop/mobile menu. Users can add bin
collections, enter a postcode, select an explicit address, save, inspect results, Refresh and
Check status. Pending work survives restart for manual continuation. Every signed-in user has
the same controls; no admin/household/owner hierarchy was introduced.

Configure operations provides Form, editable YAML and human-readable Help. Request names and
links are configured in YAML. One collection mapping handles every bin colour. Supported endpoint,
request-field and JSON Pointer repairs can be saved on the host and applied without a release.
Validation, destination review, stale-edit conflicts, reload, previous-revision restore, missing-file
recreation and supplied-example comparison are included. Unsupported transports, authentication
schemes, operation primitives and capabilities still need application support.

The supplied source is [whenisbins.yaml](../../backend/configs/integrations/waste_collection/whenisbins.yaml).
Runtime files use `configs/integrations/waste_collection/whenisbins.yaml` relative to the backend.
Definitions contain shareable request/mapping configuration; selected addresses and credentials
are encrypted private runtime state. UI saves do not commit or push Git.

Home, Devices, Overview and Panel receive no new content, tiles, counts, links or room placements.
Only shared navigation and Connections-specific routing lifecycle changed. No new public MCP,
device-context or settings-WebSocket data is exposed.

## Source ownership and acceptance evidence

Paths below are relative to the repository; Go companions are next to production files and
frontend unit companions mirror production paths under `frontend/src/__tests__/`.

| Acceptance | Implemented owners and evidence |
| --- | --- |
| AC-01–04, AC-22 | `backend/internal/integrations/setup.go`, `coordinator.go`, `service_test.go`, `coordinator_test.go`: current-candidate membership, duplicate Save receipt, exact request body/idempotency key, pending/ambiguous response continuation, unsupported-input rejection; `frontend/e2e/ConnectionsPage.spec.ts` exercises setup/results. |
| AC-06–08 (M1) | `coordinator.go`, `capabilities/waste.go` and companions: last-good preservation, 304/cache scope, backoff, manual renewal, expiry, restart continuation, local-date/DST projection and late-response fencing. AC-05 automatic work remains M2. |
| AC-03/09/18 | `models/integrations/resource.go`, `capabilities/waste.go`, `identity.go` and companions: typed exact/partial results, unknown fields/categories, missing/empty distinctions, per-service coverage, date-only semantics, stable local IDs and encrypted source notes. Provisional/neighbour results are rejected in M1. |
| AC-11/17/20/22 (supported M1 format) | `configuration.go`, `definitions.go`, `operations/request.go`, model definitions and companions; UI `ProviderForm.vue`, `ProviderEditor.vue`, `RequestValues.vue`: arbitrary request IDs, explicit recipe links, phase/capture validation, shared Form/YAML model, endpoint/field repair, atomic activation, test isolation and recovery. Second-provider proof/replacement is M4. |
| AC-07/12/19 (M1 integrity) | `repository/integrations.go`, `integration_definitions.go` and companions, `definitions_test.go`: encrypted atomic aggregates, schema/AAD checks, rollback, file hash CAS, journal reconciliation, missing file recreation and stable IDs after compatible repair. Full replacement preview/mapping is deferred. |
| AC-13/15 | `internal/api/integrations.go`, existing auth/API middleware companions and `http.go`: authenticated equal access, direct-handler rejection, cookie Origin/custom header, cookie precedence, revision preconditions, origin/slot-pinned credential reassociation, private-IP/DNS/redirect/proxy/response limits. |
| AC-09/13 public boundary | Existing `internal/hub_test.go` installs a sentinel connection and asserts actual public device-context/MCP routes and hub snapshots contain neither its data nor resource references; anonymous protected catalog is denied. |
| AC-14/19 | `repository/integrations.go` companion: AES-GCM, separate versioned keyring, mixed-key rotation, backup/isolated restore, missing/corrupt/newer-schema failure; `bootstrap.go`, `janitor.go`, `main.go` companions: optional degraded startup, retention, cancellation and join-before-close. |
| AC-16/21/23 | Frontend contracts/service/store/domain companions and selected Playwright spec: DTO validation, error/loading states, Form/YAML/Help draft preservation, keyboard focus, desktop/mobile, unknown routes, signed-out deep links and no browser PII persistence. |
| AC-26/27 | Navigation unit companion plus selected Playwright spec checks unchanged Home/Devices/Overview/Panel and no feature promotion. Isolated Home fixture has 18 tiles; the user's separate screenshot showed 15. Neither count is rewritten in production. |

AC-10 consumers, full AC-12 replacement, AC-05 automatic refresh and later-milestone portions of
shared criteria remain outside M1. Retired AC-24/25 and T029/T030 were not implemented.

## Final review corrections

The implementation review found and corrected these paths before handoff:

- Configuration captures must be available on both immediate-completion and pending-completion
  paths before Refresh can use them. Optional/unreachable captures cannot satisfy that check.
- Strict YAML/JSON parity rejects coerced/null booleans and unsafe numeric configuration literals;
  opaque provider numeric captures retain exact `json.Number` values in encrypted persistence.
- Credential destination review pins the active slot plus instance/credential revisions. Changing
  slots cannot transfer an unrelated saved token to a new origin; old unused slots remain removable.
- An unchanged response is valid only with a matching retained snapshot and sent validator. Invalid
  304 state gets one unconditional request; source age/content revision do not advance on 304.
- Ambiguous POST retry retains identical saved bytes/key. Disable pauses pending work; Enable does
  no provider I/O and explicit Check can continue the saved attempt before expiry.
- Failed writes cannot publish uncommitted snapshots. Saved-but-not-applied host files report that
  state and reconcile through their journal. Missing files can be recreated from the last active
  source while preserving instances. Unreadable files are reported separately.
- Shutdown joins owned workers before storage closes, including listener/startup failures. A
  deadline leaves storage open beneath a still-running writer rather than closing it unsafely.
- Help/editor restoration clones saved drafts, failed controls preserve useful errors, route keys
  apply only within Connections, logout/late replies cannot restore old private state, and fixture
  identity/run-nonce checks precede browser mutations.

## Actual checks

Toolchains: Go 1.24.8 (`GOTOOLCHAIN`); Node 24 via `/opt/homebrew/opt/node@24/bin`.
All provider/device-facing tests used synthetic transports, fake MQTT and temporary storage.

| Command / scope | Result |
| --- | --- |
| Backend `go test -count=1 -timeout=5m ./...` | Passed across the repository. |
| Backend `go test -race -count=1 ./internal/integrations/... ./repository ./internal/api ./internal/auth ./internal .` | Passed; final missing-file regression also passed in a subsequent `-race` run of `./internal/integrations/...`. |
| Backend `go build ./...`, `go vet ./...` | Passed. |
| Frontend `npm test -- --runInBand` | 59 suites, 360 tests passed. |
| Frontend `npm run build` | Passed, including the final provider-editor recovery controls. |
| Frontend `npm run lint` | Zero errors; seven existing script warnings remain. |
| Frontend `npm run type-check:e2e` | Passed. |
| Frontend `npm run test:e2e -- ConnectionsPage.spec.ts` | Eight passed across 1280×800 desktop and 390×844 mobile. Includes keyboard, errors, Help/drafts, login denial, privacy and unchanged-page assertions. |
| `node --test scripts/manage-stack.test.cjs scripts/create-integration-keyring.test.mjs` | 66 passed. |
| Isolated Compose configuration tests (`NODEHERDER_COMPOSE_TESTS=1`, standalone Compose binary) | 13 passed. Config-only; Docker daemon deliberately unreachable. |
| `bash -n scripts/manage-stack.sh`; ShellCheck 0.11.0 | Passed. |
| Local artifact/link check and `git diff --check` | Passed before the implementation commit. |

Initial Compose plugin discovery failed in its isolated configuration; rerunning with the installed
standalone Compose binary passed. Temporary desktop/mobile Results and editor screenshots were
visually reviewed; no mock HTML/PNG assets are retained in the repository.

Full historical browser-suite coverage was not run; eight targeted scenarios and the repository's
full unit suites were run. No live provider POST, real account/quota test, household device command,
Docker deployment, user-stack restore, sustained-load benchmark or production security audit was
performed. Synthetic tests establish implementation behavior, not external provider availability.

## Rule review

| Rule | Final M1 self-review |
| --- | --- |
| NH-01 | Existing spec → plan → tasks preceded implementation; confirmed manual/personal-Docker scope retained. |
| NH-02 | No device impersonation, commands or device event/state changes. |
| NH-03 | Protected new routes, deterministic ingress/egress checks, origin-bound credentials, encrypted PII and public-boundary negative assertions. |
| NH-04 | Repository/service/API/Vuex boundaries, typed producer/consumer DTOs, bounded YAML operations and ADR alternatives retained. yaml.v3 is an existing dependency made direct; no new framework. |
| NH-05 | Source-owned tests and acceptance mapping above; actual/unrun evidence distinguished. No live-device testing. |
| NH-06 | Explicit limits/retention, atomic publication/recovery, separate key backups, optional degraded feature and worker shutdown tested. |
| BE-01 | Business logic independent of HTTP handlers and exercised with injected transports/storage/time. |
| BE-02 | N/A: DeviceLifetimeService/event/debounce/automation semantics unchanged. |
| BE-03 | Owned bounded workers/queue/janitor, cancellation, controlled-channel race cases and affected-package race runs. No recurring M1 provider timer. |
| BE-04 | Initial schema fails closed on corrupt/newer/unknown storage; encrypted backup/rotation/restore and file recovery tested. Dates/coverage never become guessed zero data. |
| BE-05 | Auth/error/DTO/precondition contracts and actual public MCP/context exclusion tests. No new assistant tools. |
| FE-01 | Existing Vuex/services/composables/shared controls retained; components present feature data. |
| FE-02 | Runtime parsers and typed service/store ingress; isolated real Go feature API and mock/degraded responses agree. |
| FE-03 | Admission, pending, completion, failure, stale/unknown coverage, disabled and offline states remain distinct. |
| FE-04 | Existing theme/shared controls, labels/focus/touch targets, confirmations and recorded desktop/mobile browser evidence. |
| FE-05 | Existing authenticated API helpers, memory-only private state, logout cleanup, write-only tokens and escaped text rendering. |

No scoped exception was requested or self-approved. Maintainer review remains the normal merge
step; the ADR and constitutions have not been declared ratified by this work.

## Operational handoff and remaining roadmap

[docs/connections.md](../../docs/connections.md) describes explicit external keyring provisioning,
optional `--connections`/Compose overlay, host paths, backups, rotation, recovery and manual use.
Without a readable matching keyring, Connections degrades inside its page while device features
remain available. Installation/page reads never create provider lookup jobs.

Provider availability, current access conditions/quotas and host mount permissions must be checked
on the actual installation. Keys are not auto-generated/replaced at runtime. Logical deletion does
not securely erase bbolt free pages or backups; preserve/protect old keys for retained backups.

Next milestones require separate work: M2 automatic scheduling/pending continuation; M3 protected
assistant/automation consumers; M4 second-provider/replacement mapping and identity review. Widgets,
Home/Panel placements, marketplace, scripts and additional household roles remain excluded.
