# Protected API and generic recovery paths

Status: Proposed contract; production handlers remain unimplemented.
[Plan](../plan.md) · [Runtime](runtime-lifecycle.md) · [Definition](definition-format.md)

## Transport and revisions

All /api/integrations routes use the established cookie/bearer authentication. Success returns
`{data: DTO, error: null}`; failure returns `{data: null, error: {code, message, field_errors?,
retry_at?}}` with the actual non-2xx HTTP status. Field errors use path/message and optional YAML
line/column. DTOs expose revisions and sanitized provider problem codes, not raw Go errors, host
paths, credentials or upstream payloads. This envelope is specific to these new endpoints; do not
rewrite existing API responses. Frontend ingress/mocks must match it exactly.

Read-only status/resource/catalog calls do no provider I/O. Return no-store for these private
responses. 202 means accepted durable work, not completed results; include operation_id and a
protected status URL. 201 means a saved setup/instance, not proof the provider finished. Disconnected
clients retrieve status rather than resubmit. Bounded API polls stop on unmount/logout/terminal state;
a hidden page does not start another provider worker. Surface an expired setup as 410, not an empty
candidate list. Default body/route limits still apply; 413 remains a real failure.

Definition content_hash is SHA-256 over exact UTF-8 bytes. Its strong ETag is the quoted string
`"sha256:<hex>"`; PUT/DELETE/Reload/Restore use that value in If-Match. New files require
If-None-Match: *. Instance/setup mutations use a scoped quoted ETag containing entity ID/revision,
returned in both DTO revision metadata and response headers. Missing required preconditions: 428;
wrong revision: 409; invalid configuration: 400. Any supplied body revision must agree with the header.
A saved_not_applied result is a non-success 503 carrying current disk/active hashes and a retry action.

Existing CORS only allows Content-Type/Authorization/Cache-Control/Connection. For integration
paths only, allow X-NodeHerder-Request, If-Match and If-None-Match in approved-origin preflights;
expose ETag and Retry-After if the client reads them. Keep existing origin/credentials policy and
other paths unchanged. Cookie POST/PUT/DELETE still check exact Origin and the custom header on
actual requests; CORS is not authentication or CSRF enforcement. Cookie-plus-bearer follows the
actual middleware's cookie precedence. Treat 401 as expired auth and preserve only nonsecret draft
edits in memory; discard setup PII/credential buffers on logout or session invalidation.

## Setup/save state machine and local retries

POST setup creates a private draft, not provider work. POST step with search input admits address
lookup; GET setup/operation status only reads local state. A valid current candidate can be selected;
Save claims the setup once and atomically creates the instance/binding plus durable first attempt
before any lookup POST. A slow provider yields a saved instance with pending health. Resources is
an empty list until its first validated snapshot; do not fabricate a revision-1 schedule. Manual
Check status continues its saved lookup. Editing an existing instance uses the candidate/commit
rules in the runtime contract rather than overwriting its binding before validation.

A setup has at most one committed instance. Repeating the same Save after a lost local HTTP response
returns its existing instance/operation, never a new provider key. A different submission after claim
is 409. Keep an encrypted completion receipt (setup ID, target instance, submission digest, outcome)
for 24 hours; it has no candidate list, full address, secret or exact POST body. A GET setup can return
this completed receipt. Before claim, stale revisions fail; completion checks do not revive expiry.
Enforce uniqueness for an active definition/property/capability binding and offer the existing
connection when the same property is configured again. No silent merging of distinct properties.

## Missing control paths

In addition to catalog/setup/instance/resource/definition paths in the plan, provide:

| Route | Request and outcome |
| --- | --- |
| PUT /api/integrations/setup/:id/credentials/:slot | Write-only value and setup If-Match; encrypt immediately; return configured/revision only |
| PUT /api/integrations/instances/:id/credentials/:slot | Write-only value and instance If-Match; rotate reference/revision, cancel obsolete work; never echo value |
| DELETE /api/integrations/instances/:id/credentials/:slot | Explicit credential removal with If-Match; required auth becomes reconnect_needed; optional auth changes only after this explicit action |
| DELETE /api/integrations/definitions/:type/:provider | If-Match; 409 definition_in_use while instances or setup/pending refs exist; no cascade deletion |

Slots must be declared in the compiled definition; callers cannot name another instance's secret.
Optional bearer means “no configured token is allowed”, never “discard a token after its 401”.
Provider 401 pauses authenticated work and offers replacement through these same controls. Setup
credentials are scoped to that setup and transferred atomically on save; cancellation/expiry removes
orphaned secret records. No plaintext credential appears in GET/export/history/diagnostics.

Origin/credential review is part of normal shared Save, not a new role or review-ticket service.
Validate reports old/new origin and affected instance/credential revisions. If destination changes,
Save supplies a separate destination_review acknowledgement naming the exact new origin and
explicit credential reassociation entries with their current revisions. Recompute and check them
server-side alongside If-Match; stale reviews are conflicts. The acknowledgement is outside YAML,
so pasted/imported config cannot grant itself secrets. Instances not reassociated remain
reconnect_needed. No extra signed token, signing key, expiring review store or privileged account
is needed. All authenticated configuration authors are trusted to authorize these transmissions;
write-only secrets prevent accidental disclosure, not a stronger privilege boundary against them.

## Search/select recipe and refresh recovery

M1 supports exact property-choice journeys. The search step's optional `expect` discriminator
checks the provider's required-input field before displaying choices. If another council requires
street/locality/free-text/weekday inputs, return unsupported_setup with a clear explanation; do not
show an empty choice list or guess values. Unknown candidate IDs and stale setup revisions fail.
The supplied M1 definition/flow never sets representative opt-in or displays neighbouring-property
data as an exact result.
Receiving provisional/non-exact data cannot publish a household snapshot. Future representative
support requires its own supported recipe/consent and positive acceptance cases, not a hidden branch.

Save the minimal confirmed inputs (postcode/opaque property ID), provider state captures and binding
identity encrypted per instance. These are separate from the exact per-attempt bytes/key, which are
deleted on terminal/expiry. They allow a *new* renewal request without retaining a completed request
history or full address list. The live body is rebuilt using the pinned declarative request and
saved typed inputs, then persisted immutably with its new key before sending.

`refresh.request` reads the stored schedule. Optional `refresh.renewal_request` names the existing
selection request for a new exact-property lookup. No provider/request-name switch is permitted.
Configured HTTP-status/problem rules classify provider errors; their actions are a bounded enum,
not scripts. For example, only HTTP 404 + no_schedule may request renewal; an unknown 404/not_found
is a broken endpoint and must not trigger another POST. lookup_not_found requires recovery/re-setup,
401 reconnects, 429/503 wait, and ambiguous/unsupported property responses require user input.

A renewal action is only valid for the configured refresh request, never a nested renewal or an
unrelated search/lookup error. This phase guard and one-renewal limit prevent rules from creating
cycles. One user-triggered M1 Refresh may read the schedule then start at most one renewal when the recipe
allows it and the saved binding is valid. M2 uses the same path. Renewal is also warranted when a
previously dated applicable service has no future date left; never poll a terminal lookup expecting
it to change. Bound renewal frequency (initial proposal: at most once per instance per 24 hours),
global/provider quota and total attempt budget. If renewal is unavailable, recently attempted or
cannot resolve its state, retain/label old data and show the next action/time instead of looping.
A valid partial schedule may remain visible while renewal runs. Missing/unknown input never becomes
a guessed property, and a changed selected property requires a new confirmed setup binding.

An ambiguous POST reuses identical body/key under the verified provider contract and local attempt
TTL. Retry deadlines and credential/grant identity also apply. Never convert an uncertain retry
into a new key automatically. Provider deletion can invalidate its retained idempotency record:
there is no unconditional exactly-once guarantee across remote data loss. A returned conflicting
lookup/property identity is quarantined as reconciliation_required. Local expiry is 24 hours even
if the provider advertises longer retention.

A 301/other redirect is provider_moved in M1: no automatic credential forwarding; retain old data
and offer reviewed reconfiguration. 304 requires the matching cache entry as defined in the runtime
contract. All recovery attempts are durable/coalesced and generation-fenced. DELETE/Cancel only
cancels local work; no provider cancellation endpoint is assumed.

## Test and repair behavior

Offline Test validates only a supplied redacted sample; it cannot prove a new URL works. Live Test
is an explicit POST action, uses ordinary origin/secret/admission/idempotency rules and returns a
private test-operation ID. It must not publish instance snapshots, alter active config, or trigger
consumers. Cancel/expiry cleans its encrypted temporary state. A newer draft never mutates an
already admitted test. Disable/delete/config/credential races have the same fencing as real work.

Use source-owned API/middleware/service/setup/configuration/repository/frontend companions. Include
cross-origin conditional-save preflights, stale destination acknowledgements, missing preconditions, credential
create/remove/reauth, 404 code distinctions, missing payload on 304, saved-input renewal, expired
setups, disabled reads versus mutations, property edits and failed publication. No household devices
or routine live POSTs. UI faults/statuses remain inside Connections, not other pages.

## Closed error/size boundaries

All mutation failures are typed; callers never infer an action from English text. Error rules are
bounded (at most 32), exact status/code comparisons only, and cannot turn 401/429/503 into a retry
that bypasses common safety. Provider code/detail strings are bounded and rendered as text; logs
retain sanitized codes, not raw notices/candidate lists/definitions. Default outcome is a useful
provider_error without dropping a stored credential or inventing data.

Use the existing 1 MiB request-body ceiling for local API transport, then the smaller 64 KiB YAML/
secret/sample limits before parsing. Declaration parsing bounds depth, count and encoding; HTTP
responses are single bounded JSON documents. A supplied sample is limited independently and cannot
select an origin or secret. Keep the feature origin/network checks on actual dial/redirect targets;
an offline validator may confirm syntax but cannot certify future DNS answers or live availability.

Concrete initial local limits: secret value 8 KiB, offline sample 256 KiB, YAML 64 KiB; transport
ceiling remains 1 MiB. At most 128 definition files, 32 active setup/test sessions, 64 admitted queued
attempts, four executing external requests and one logical operation per instance. Refuse excess
with a bounded busy outcome instead of spawning workers. Provider notices are at most 32 entries /
2 KiB each. The 2 MiB decoded upstream limit dominates the theoretical product of list/date limits.
These are proposed operational bounds to validate with fixtures, not measured production capacity.
