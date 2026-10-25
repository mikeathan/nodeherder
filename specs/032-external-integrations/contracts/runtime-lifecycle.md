# Runtime ownership, recovery and publication

Status: M1 implementation contract; later-milestone provisions remain deferred. See [implementation review](../implementation-review.md).
[Plan](../plan.md) · [Security](execution-and-security.md) · [API/recovery](api-and-recovery.md)

## Composition and shutdown

Current `backend/main.go` calls store.Create, hub.Register and h.Listen, with os.Exit on startup
errors. `internal/api/server.go` closes HTTP on context cancellation and Listen can return before
background callbacks finish. AppStore has no Close method. A context.AfterFunc alone is not a join.

Proposed `internal/integrations/bootstrap.go` owns one optional Feature with its service/repository,
child context, bounded worker group and idempotent Close(context.Context). Construct it in main's
composition root, independently of AppStore; inject the service into hub/router registration through
an optional composition option. Preserve existing constructors/callers and device-only test fakes.
Do not add mandatory integration repositories to every AppStore mock or return a fatal store.Create
error when this optional feature is unavailable.

Put the new ownership in a run function in main.go: startup errors return after cleanup; only the
outer main decides the process exit. On signal, HTTP stop, listener failure or hub registration
failure: stop integration admission, cancel its I/O/timers, join workers/janitor/coordinator, then
close its DB exactly once. The close deadline uses a fresh bounded background context, not the
already-cancelled application context. Existing server/device protocols are not redesigned.
Source-owned checks: main_test.go, bootstrap_test.go, hub_test.go and affected service companions;
use fake constructors/listeners and controlled synchronization, not household services.

M1 has bounded user-request workers and a local expiry janitor. M2 adds the one due-work coordinator.
Polling waits are persisted timestamps, not sleeping goroutines or occupied FIFO lanes. Each task
owns its cancellation and wait-group membership. Do not use Clock.Sleep for interruptible waits;
reuse Clock.AfterFunc plus cancellation. Closing twice, cancelling during a write, a stuck fake HTTP
request and a listener that fails immediately must all have bounded outcomes.

## Strict storage and degraded startup

The generic GetDataDir/EnsureDir and KV constructor can fall back to temporary storage. Do not use
that behavior for integration DBs, keys or host definitions. Read the configured data path directly,
create a private directory or return a feature error, and open one bbolt handle with a bounded lock
wait. Record the actual selected path in deployment documentation; do not log private record bodies.
Partial constructor failures close any already-open handle and start no leftover worker.

An unavailable key, unwritable directory, locked/corrupt DB, unsupported schema or invalid definition
produces a degraded Connections service/status, not a fatal hub.Register/store.Create failure.
Its read-only catalog can report feature availability; operations needing private state return 503
with a stable code. Do not pretend the installation has no connections, replace keys/DBs, run
migrations on unknown versions or overwrite unreadable data. Unaffected device management still runs.
No permanent feature warning/banner is added to other views; status lives within Connections.

Store schema version in the owned DB. Validate and back up before an upgrade, migrate in one bbolt
transaction per schema step, and commit the version with its data. Abort/rollback on failure; unknown
newer schemas fail closed. Definition-format versions and capability versions are independent.
Backups require a consistent DB snapshot, matching host YAML and separate key material. Test old,
current, newer and corrupt fixtures; “migrations supported” is not sufficient evidence.

## Identity, publication and edits

One logical operation per instance/setup owns its immutable request bytes/key, pinned definition,
binding revision, generation, credential revision and origin-grant revision. Creating an attempt
is durable before provider I/O; publishing result, identity mappings, validator and run checkpoint
is one repository transaction. Failed writes publish nothing. Expose operation IDs/status rather
than keeping a browser request open through an upstream lookup.

Disable, delete, credential changes and activation advance the relevant generation and cancel old
I/O. Check generation/revisions again inside the publication transaction. M1 enable does not start
provider work automatically. Disabled instances retain readable last-good data with a disabled
status; Refresh/Check status are rejected until enabled. M2 enable can admit due work normally.
Cancel stops local work; it cannot retract a provider lookup already admitted upstream. State that
honestly and never create a replacement lookup merely because a browser disconnected.

Editing a saved property uses an isolated setup candidate while the previous binding/snapshot
remains active. Only a validated completion commits the candidate. Same confirmed provider property
can keep resource identity. A different property allocates a new resource identity and flags old
consumer references for explicit future review; never relabel old dates as the new property's dates.
Failure/cancel leaves the old binding intact. Shared revisions reject conflicting edits.

Pending POST attempts keep their original bytes/key/definition. Config edits do not rewrite them.
For pending status continuation, retain the old compiled revision until completion/expiry; if it no
longer works, explain that and offer explicit restart of setup. Do not migrate an ambiguous POST to
a new URL or credential. Pending operations cannot silently acquire a new credential/grant version.
An origin change may activate a definition while some instances are blocked awaiting reconnection;
return their states/counts honestly rather than claiming every connection is usable.

## Bounded cache and work

Cache validators are scoped to request ID, definition hash, binding/generation, credential/grant
revision and canonical URL/query/representation. Never use one instance-global ETag for lookup and
schedule requests, or reuse it after mapping/origin/credential changes. Keep sensitive key material
out of cache keys/logs. Store a validator only after successful validation/persistence.

Accept 304 only when that exact key has a retained validated payload. Otherwise make at most one
unconditional retry within the work budget; a second invalid/304 response is an error. A 304 updates
last successful check, not source age, freshness expiry, coverage or content revision. Status polls
and schedule reads have independent validators. M1 can simply leave lookup conditional reads off.

For status continuation use max(now + minimum interval, Retry-After, quota deadline/backoff), with a
positive bounded delay for missing/past/invalid headers. The initial minimum is five seconds. Honor
provider-wide limits across instances and retain deadlines over restart. Expiry is checked before
admission, not just by a best-effort janitor. No retry after deletion/disable/expiry.

Definition activation invalidates old validators; host file save/reload follows the separate
configuration journal. Reconcile journal/file hashes after crashes before resuming provider work.
Do not repair an invalid externally edited file by silently overwriting it. Size/count limits also
cover number of definitions, runtime history and queued attempts, not only bytes per HTTP response.

## Composition interfaces (proposed)

The API depends on a small IntegrationService for catalog/setup/instance/definition/operation DTO
methods; it does not own DB handles or start timers. Feature owns that concrete service and Close.
The repository exposes transactions at aggregate/publication boundaries, not buckets to controllers.
The runner consumes an immutable CompiledDefinition, attempt context, scoped credential handle,
policy-bound HTTP client and clock, returning pending/validated/unchanged/typed failure. The waste
normalizer receives both selected-property evidence and lookup status alongside its mapped result;
response_root must not discard the context needed to reject a provisional or mismatched result.
Inject these at constructors. Do not create package-global service state or a broad mutable plugin
service locator. New fields do not extend AppStore or public hub/device serialization.

Once an operation is durably admitted, execute it under the feature-owned context and bounded
attempt context, not the HTTP request context that ends with a 202 response or disconnected client.
A client can cancel through the explicit mutation endpoint; status polling has no worker ownership.
Test disconnect-after-admission and shutdown-during-publication separately. If the shutdown deadline
expires before workers stop, return an error and do not close a DB underneath active writers; do not
claim a successful graceful shutdown. The process-exit/OS recovery limit remains explicit.

Capture state is binding/attempt scoped. A different-property candidate starts with empty provider
state; optional capture preservation applies only within the same binding/continuation. On definition
activation, missing newly required saved inputs block that instance as reconfigure_needed rather
than guessing, borrowing another binding's state or discarding its last-good snapshot.

## M1 continuation clarification

Disable pauses a durable pending attempt and fences its running response. Enable alone performs
no provider I/O; an explicit Check status may resume that same attempt/key before expiry. Delete,
credential replacement and incompatible configuration invalidate obsolete publication.
