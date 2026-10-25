# Waste schedule v1 wire and identity contract

Status: Proposed typed API contract; [examples](waste_schedule.v1.examples.json) are developer
reference data, not user configuration. No runtime JSON Schema library is required.
Spec: [../spec.md](../spec.md) | Plan: [../plan.md](../plan.md)

## Representation and nullability

The fields described below have explicit Go/TypeScript types and nullability. Consumers require
a supported contract_version. Map and validate the consumed provider fields; ignore unrelated
upstream additions. Unknown upstream fields do not become additional wire properties. Unknown supported source labels
are preserved in the dedicated nullable raw-label fields and map to the enum `unknown`.

| Field | Type / meaning |
| --- | --- |
| contract_version | Literal `waste_schedule/v1` |
| instance_id, resource_id | Lowercase canonical UUID strings assigned by NodeHerder |
| revision | Integer >= 1; version of accepted persisted resource content |
| kind | `exact`, `partial`, `provisional`, `confirmed_empty`, `unavailable` |
| timezone | Literal `Europe/London` for this UK capability version |
| captured_at, last_checked_at | UTC RFC3339 strings ending Z; fetch/validation time and last successful check respectively |
| source_retrieved_at | UTC timestamp or null; never invented from capture/check time |
| fresh_until | UTC timestamp or null; explicit source freshness expiry, not last collection date |
| evaluated_at | Current projection evaluation UTC timestamp; response metadata, not a persisted revision change |
| freshness | `fresh`, `stale`, `unknown`, evaluated by the rules below |
| quality | Required address_match, evidence_granularity, date_confidence, provisional and nullable source_confidence_label |
| provenance | provider_id/version, positive profile_revision/binding_revision, council_gss nullable, source_url nullable |
| representative_consent | `not_requested`, `declined`, `accepted`, `not_required`; safe summary, not actor/PII |
| notices | Array of provider caveat strings, bounded and rendered as text; may contain private information |
| services | Array of unique locally identified services; may be empty only under the rules below |

The v1 wire name profile_revision refers to the active compiled YAML-definition revision; it is
runtime provenance, not a separate editable DB-only configuration source.

Each service has service_id (UUID), source_service_key (opaque string or null), name (text),
waste_type (`refuse`, `recycling`, `paper_card`, `glass`, `garden`, `food`, `mixed`, `other`,
`unknown`), source_waste_type (string or null), bin_colour/container (string or null),
subscription (`required`, `not_required`, `unknown`), date_state (`known`, `confirmed_empty`,
`unavailable`), dates (date-only string array) and coverage (state, from, through).
The vocabulary is NodeHerder-owned and does not pretend that a mixed brown bin contains glass only.
Unknown source categories preserve their original string; new meaning requires a schema revision
or a supported `other`/`unknown` mapping, not a silent new enum value.

Map subscription_required true to required; absent is unknown, never not_required. This flag does
not establish that the household is actually subscribed. Preserve date confidence/coverage and
provider notes rather than dropping qualifications. Exact-service removal is not established by
absence from a response. A previously seen but unreturned service keeps its identity and appears
unavailable with empty dates/unknown coverage in a partial result, even if the provider lookup
is terminal done; retain original evidence where
needed by a later opted-in consumer, not by guessing a recurrence.

Quality enums:

- address_match: `exact`, `street`, `postcode`, `postcode_representative`, `unknown`.
- evidence_granularity: `property`, `street`, `postcode`, `area`, `round`, `unknown`.
- date_confidence: `published_calendar`, `council_projection`, `next_collection_only`, `unknown`.
- coverage.state: `complete`, `limited`, `next_only`, `unknown`, `not_applicable`.

Ordinary typed decoding and focused validation in the Go normalizer enforce the following rules;
frontend ingress validates the API fields it consumes. No generic schema engine is needed. Dates must be real calendar dates, not merely regex matches.

## Dates, coverage and freshness

Collection dates are sorted unique `YYYY-MM-DD` local dates, not UTC instants or guessed collection
times. Operational timestamps are UTC. Coverage bounds are inclusive local dates or null. A
non-null through must be >= from; provided concrete dates lie inside provided coverage bounds.
Do not fill coverage bounds with min/max returned dates: those dates do not prove completeness.
Use null where the provider does not state a coverage boundary. `limited` does not imply a complete
interval; `next_only` describes a concrete next date without a promise of subsequent dates.

`date_state: known` requires at least one date. `confirmed_empty` requires an explicitly evidenced
empty interval with non-null coverage bounds and dates []; an empty upstream list alone is not that
evidence. `unavailable` requires dates [] and unknown coverage. At resource level, confirmed_empty
with services [] means the provider explicitly establishes no applicable services, not just no
upcoming dates. Preserve the supporting explanation in the sanitized UI outcome. An unavailable
response cannot overwrite a last-good exact snapshot.

`kind: partial` means an incomplete exact/property-qualified result; keep incomplete services and
unknown coverage visible. `kind: provisional` requires quality.provisional true, postcode_representative
address match, fresh_until null and freshness unknown. It is transient, including while a prior exact
snapshot exists. Final non-exact representative results are not promoted to this exact-household
resource in M1. Provisional results never enter the persisted resources endpoint, automation or feeds. M1 is exact
property-choice only and displays no neighbour result or representative-consent flow; the provisional
example is a negative-ingress/future-contract fixture, not an M1 UI promise.

Freshness is independent of coverage and last check: unknown if source_retrieved_at or fresh_until
is null; otherwise fresh if evaluated_at < fresh_until, stale if >=. fresh_until must be >= source
retrieved time. A validated provider refresh_due_at can supply expiry; absent such evidence, use
source_retrieved_at + the explicit profile freshness budget (proposed default 24 hours). Never base
expiry on a new HTTP check/304. A future expiry does not invent dates beyond service coverage.
UI must also show check time and source age so a provider cache check is not represented as fresh
council verification. Reject impossible timestamp ordering rather than silently converting it.

## Stable resource/service identity

Allocate one random UUIDv4 resource_id per configured waste resource, independent of postcode,
provider, display name and service count. Never derive IDs from PII. Persist identity allocation
and its mapping with the first accepted snapshot in the same repository transaction.

Allocate UUIDv4 service_id when a service first appears. Unique binding index:
`(resource_id, provider_binding_id, binding_local_key) -> service_id`. provider_binding_id is a stable persisted UUID for the confirmed provider/property association;
binding_revision is only an audit/version counter and is **not** part of the identity index.
Ordinary revision/credential/config refreshes retain that UUID, so they cannot regenerate service
IDs. A different property/provider binding gets a new association; M4 preserves IDs only through
its explicit reviewed mapping. Provider keys are namespaced by that association, never globally stable. When there is a mapped stable provider service key,
use it as binding_local_key. Otherwise M1 uses a persisted descriptor key from SHA-256 of the
canonical JSON tuple [exact provider name, raw waste type or null, container or null], prefixed
with descriptor:v1:. Serialize the fixed ordered nullable-string tuple with Go encoding/json.Marshal
(default escaping), hash its UTF-8 bytes and use lowercase hex; do not trim/case-fold labels.
Do not use array position or colour. source_service_key stays null on the wire
when the provider supplies none. Duplicate descriptors are an ambiguity, not an index-based split.
A descriptor change receives a new UUID; no claim of automatic rename matching is made. Preserve
old identity as unreturned/unavailable rather than declaring cancellation. Later reviewed semantic
mapping can preserve identity across a rename/replacement. Exact unambiguous persisted mappings may be reused. A changed/ambiguous source
shape must produce reconciliation_required instead of assigning an existing service ID arbitrarily.
Service display renames keep identity when a stable provider key or reviewed mapping establishes it;
without either, automatic rename recognition is not promised.

UUID insertion collision: retry random generation at most three times, then fail the transaction;
never overwrite an existing identity. Duplicate source keys in one payload are invalid unless the
adapter has an explicit tested split rule. Detect duplicate service IDs and dates as ingress errors
or normalizer-deduplication cases as specified above; never silently merge distinct services.

Removed services become tombstones in the identity map; omit them from the current data only when
removal is explicitly evidenced. An incomplete/missing upstream service cannot erase them as if
cancelled. A replacement mapping can reactivate a tombstone only after explicit semantic confirmation.

## Provider replacement

Replacement is M4, not implicitly delivered by a second provider. Keep the active binding and
snapshot readable while a candidate binding/snapshot is fetched privately. Candidate data cannot
publish resource revisions or trigger automations. Require confirmed same property, compatible
capability/units, subscription meaning and an explicit one-to-one service mapping.

The preview lists old/new names, dates, coverage, quality, unmapped services and affected recipes.
Ambiguities remain unresolved until user selection. Mapped services retain service_id; truly
new services receive new IDs; retired IDs remain tombstoned. No match solely on colour/category.
Commit the new binding, mapping and snapshot atomically with generation fencing. Keep one previous
encrypted binding/snapshot for seven days for explicit rollback, without concurrent refresh of both.
Recipes with changed/unmapped semantics require review and must not silently target another service.
Rollback is a new reviewed activation and must not replay historical events.

## Examples and verification

The examples file contains complete structurally valid exact, partial, provisional, confirmed-empty
and unavailable representations. They use synthetic UUIDs/provider data, not household addresses.
Provisional/unavailable examples illustrate setup/run responses, not permission to persist them as
last-good resources. Current exact data is retained separately during those outcomes.

Acceptance evidence must include typed decoding/format checks, real-date validation, null versus
empty, enum unknown handling, duplicated source keys, stable mappings across rename/restart,
ambiguous replacement and atomic failure. Production companion files: capabilities/waste_test.go,
repository/integrations_test.go and src/__tests__/contracts/integrations.test.ts. JSON examples are
contract artifacts, not substitute application tests.


## Storage of provider text

All snapshots and pending/private records are encrypted at rest, not just credentials/bindings.
Provider notes can contain an address even when matched_address is removed. Do not promise an
address-free raw-text snapshot. The protected resource DTO omits direct property identifiers and
credentials; any bounded provider notices remain sensitive household context, rendered as plain
text with no-store, excluded from public routes/logs/exports/agent context unless explicitly needed.
Do not derive bin contents from colour. Source/calendar URLs are validated external links, never
automatic fetches or credential recipients. Unsafe schemes/URLs are omitted with a useful warning.
