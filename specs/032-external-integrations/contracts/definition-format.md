# Small integration definition format

Status: M1 implementation contract; later-milestone provisions remain deferred. See [implementation review](../implementation-review.md).
Spec: [../spec.md](../spec.md) | [Operations/help](operations-and-help.md)

The first release uses ordinary Go configuration structs and a short validation pass, matching
the project's existing typed backend/frontend approach. There is no JSON Schema execution engine,
schema-generated UI, or requirement for users to write a response schema. schema_version is simply
a format-version marker so unsupported formats produce a useful error.

## Fields users can configure

| Section | Meaning |
| --- | --- |
| schema_version | Literal nodeherder.integration/v1 |
| integration_type, id, name | Folder/type, provider filename ID and display label |
| capability | Supported output type, initially waste_schedule/v1 |
| base_url, auth | Provider URL and optional private credential-slot reference |
| inputs | Input IDs with label, simple type and required flag |
| requests | Named GET/POST requests with path, query/body references and optional captures |
| setup | A supported search_select recipe referencing those requests and choice/status fields |
| refresh | Saved-schedule request and optional new-lookup renewal request |
| policy | Explicit minimum status-poll, renewal spacing and documented idempotency retention bounds |
| errors | Optional provider error-code pointer and bounded HTTP-status/code/action rules |
| mapping | Locations of the consumed metadata and collection fields in the response |

Request names are user-chosen identifiers. The recipe type and basic operation types are defined
once in code. Each reference must point to a declared input/request/state capture or reserved
transport value. Values use {ref: input.postcode} / {ref: state.lookup_id}; no expression language.
Request path segments are escaped individually. Response selectors use JSON Pointers, not scripts.

The [example](whenisbins.definition.example.yaml) shows the intended provider shape for the bins
flow. Normal household setup does not require writing any of it: users choose the supplied provider,
enter their postcode and select an address. Form and YAML views edit the same host document.

## Minimal necessary validation

Backend validation checks:

1. Valid bounded YAML with recognized fields, one supported format version and matching type/file IDs.
2. Supported HTTPS destination, request method, recipe, credential reference and output capability.
3. Required request/input/field references exist; status states do not conflict; incomplete setup
   cannot refresh before its required saved token is available.
4. The provider returns the fields we use in the expected types: choices/collections are lists,
   names/IDs are strings, dates are valid date-only strings, quality flags have usable meanings.
5. Invalid new data/configuration produces a field-specific error and leaves last-good operation
   usable, subject to the explicit saved-not-applied/file recovery contract.

The search step may declare `expect: {pointer: /required_input, value: property_id}`. A mismatch
is unsupported_setup, not a fabricated empty list or consent branch. A refresh can declare
`renewal_request` naming the existing selection request. It reuses encrypted confirmed inputs,
creates a new logical attempt/key and follows the same status/capture rules. Error rules compare
only an integer HTTP status and exact bounded problem code; actions are renewal, restart_setup,
reconnect or unsupported. 429/503 backoff and 401 handling remain common transport behavior.
Rules cannot bypass egress, cadence, origin grants or validation. No arbitrary conditions/scripts.

Policy defaults are a five-second minimum status check and 24-hour minimum new-lookup renewal
spacing; provider-specific longer values may be configured within validated bounds. Idempotency
retention must come from a documented provider contract; absent evidence, ambiguous automatic
POST replay is disabled. Local attempt retention stays at most 24 hours. The policy is a declared
assumption to validate, not an unconditional exactly-once guarantee across provider data loss.

Mapping metadata has optional notes and source URL selectors; collections may map optional
subscription_required and source service key fields. The capability owns their meanings. Absent
subscription_required is unknown, never false/not_required. Provider text may contain PII: persist
it only in encrypted records and render as text through protected APIs. Ordinary ignored upstream
fields are not copied into snapshots or previews. `provisional` absent can mean false only with
consistent exact-property evidence; a contradictory terminal/provisional combination is invalid.

Request captures accept string, JSON number, boolean, date-only or UTC timestamp values. Optional
absence preserves existing state. when: complete delays a capture until terminal success and output
validation; strip_prefix is the one initially supported configurable string transform. No generic
numeric coercion, expression language or arbitrary transformation framework is required. Common
date validation/sorting and waste-category normalization belong to the capability code.

Server validation is authoritative. The UI adds simple field feedback and displays server errors
with field/YAML positions. Shared contract tests keep Go, TypeScript, help and the example aligned;
we do not introduce another schema engine to generate the whole UI.

## Provider schema changes

A moved/renamed field is a mapping edit; a changed URL is a request edit. Unrelated new upstream
fields are ignored. Wrong types or missing required data cause a clear error. New colours/services
need no definition changes because collection fields are mapped once across the returned list.
New auth protocols/transports/semantics can still require code. No complete copy of a provider's
OpenAPI or every response property is stored in YAML.

The [waste contract](waste_schedule.v1.md) and its synthetic examples describe the internal/public
result for developers, including identity/quality/freshness. They are not user configuration and
do not require a schema per bin or a runtime JSON Schema library. Future generated schemas are
optional if real editor/API tooling makes them worthwhile; they are not part of M1.


## Compile and execution order

Check cross-references and definite availability by phase: search sees declared search inputs,
selection sees the validated choice, status sees a captured lookup ID, refresh/renewal see the
confirmed saved binding. A reference merely declared by a later capture is not usable earlier.
Reject forward/uninitialized/cyclic references and reserved state capture collisions. Status is
read from the full body before result mapping; missing optional result on a pending lookup is
normal. Required immediate captures such as lookup ID are persisted on pending responses;
complete-only captures commit only with terminal validated output. Apply field/type checks before
updating state, and perform generation/revision checks again before publication.

Host definitions can contain ordinary constants; schema checks cannot reliably detect every secret
or personal value someone manually pastes into one. The UI must explain that fixed values/comments
are shareable/exported, reject literals in supported credential positions and never log raw YAML.
Keep credentials/selected addresses in dedicated private controls; exports require intentional review.
This corrects an absolute “secret-free arbitrary YAML” guarantee that validation cannot enforce.

Configuration authors are trusted to choose provider behavior: generic validation cannot infer
the semantics of every manually edited provider parameter. The shipped M1 exact flow has no
representative-consent input; returned non-exact/provisional data is still rejected independently
of any request edits. Do not claim arbitrary YAML edits can be proven semantically consent-safe.
