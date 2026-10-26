# Integrations board contract

Approved implementation scope: AC-34–37, 2026-10-26. Decision: [ADR-002](../adr-002-connection-board-layout.md). One widget per saved instance, no resizing.
All signed-in users share order/visibility. No Home settings/device/WS storage is reused.

## Protected API and private storage

GET /api/integrations/board → 200 {data:{etag:string,entries:[{instance_id:string,hidden:boolean}]},error:null}.
ETag header matches body. Entries include all existing instances; array order is display order.
PUT same path with If-Match and {entries:[...]} → the same shape on successful atomic commit.
Existing auth, cookie Origin/X-NodeHerder-Request checks and private/no-store apply. Require exactly
one entry per live instance, no duplicates/unknown IDs, boolean hidden, max 50 entries, IDs <=256 chars.
Missing fields are invalid; absent If-Match→428, stale membership/preferences→409; invalid input→400;
unavailable/storage failure→503. Layout reads/mutations never perform provider I/O.

Persist optional Instance.board {order:int,hidden:bool,revision:uint64}; absent means visible, ordered
after explicit preferences by immutable CreatedAt then ID. Atomically update only changed preferences.
Do not increment instance lifecycle Revision/Generation/UpdatedAt or alter snapshot/credentials.
Board ETag hashes the sorted live instance IDs plus preference revisions. Membership changes conflict;
provider refresh/name/credential changes do not. Deletion naturally removes preferences. Restore only
changes hidden. Unknown JSON fields are rejected at the API boundary. Existing ciphertext schema/AAD
and bucket set remain unchanged. Old records need no rewrite; additive optional fields load normally.
Downgrading to an older binary can lose preferences when it rewrites an instance; private schedules
remain compatible. Same-key restart persistence and write rollback are required checks.

## Frontend ownership

Vuex holds accepted instances/board/catalog and ephemeral return-focus/scroll data; logout clears all.
useConnectionsBoard owns page reads, bounded single-operation observation, focus/minute local reads,
manual Refresh/Check and request/session cancellation. useConnectionBoardLayout owns drafts, immutable
movement/hide, save/cancel/conflict behavior. Catalog composable owns Add dialog/setup navigation.
Components present typed models and emit user intent. No per-widget timers/requests, new framework,
raw provider HTML, weather feature, resize dependency or browser-persisted resources.

Card links target the existing private Connection route. A stretched link makes the card content
clickable; sibling action buttons are above the hit area, never nested in a link. Edit mode replaces
links with headings and explicit drag/move/hide controls. Movement has tap/keyboard alternatives.
Normal status/error text remains readable; short UK calendar dates reuse the schedule projection.
No upcoming date is not cancellation: distinguish confirmed_empty, unavailable and exhausted horizon.
Unknown capability falls back to the named navigable connection instead of guessing a value.

Add dialog contains configured-hidden restore and available providers/configuration/reload controls.
New setup receives whitelisted return_to=board; existing-address setup retains the default detail
return. Pending saved instances appear on the board with Check status, never fabricated results.
GET/local projection reads never trigger upstream refresh. Each navigation stores only transient
instance-ID/focus/scroll intent. After return, focus the originating visible link (without scroll)
and restore the prior scroll; use Add/title fallback for removed/hidden instances. Browser reload
needs saved layout, not persistent scroll or PII. Reload after conflict is explicit and discards
only the local layout draft. Refresh uses existing admission/backoff/error reporting.

## Rule gate and verification

NH-01: spec/plan/tasks before code. NH-02/BE-02: device pipelines untouched. NH-03/BE-05/FE-05:
protected no-store API, existing CSRF/auth, IDs/preferences only, generation cleanup. NH-04/BE-01/
FE-01–02: service/repository boundaries, typed runtime-validated contract, no new bucket/schema/WS.
NH-05/FE-04: source-owned tests and isolated browser fixtures at specified sizes. NH-06/BE-03–04/
FE-03: atomic layout writes, independent preference revisions, no new background provider work,
retained failed drafts, missing-vs-empty data and confirmation for existing deletion.

Tests: board.go→board_test.go (membership/conflict/validation/isolation/rollback/restart), existing
api/integrations_test.go (auth/CSRF/revisions/no-store), existing contract/service/store companions,
connection-board.ts→domain/connection-board.test.ts, layout composable companion, and existing
ConnectionsPage.spec.ts (full flow and UI boundaries). Run Go suite/build/vet/affected races;
frontend unit/lint/build/E2E types; isolated browser checks only. Report actual evidence separately.
