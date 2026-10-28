# Home layout v1

Producer: settings AppConfigCache through protected `GET /api/home/layout` and
`PUT /api/home/layout`. Consumers: Vuex Home layout module, Home/panel/picker; test mocks.

Envelope: `{data: HomeLayout|null, error: {code,message}|null}`. `Cache-Control: private, no-store`.
Success body/header includes a strong content-derived ETag. GET reads local state only.
PUT requires If-Match; missing → 428, stale → 409, malformed → 400, unavailable/storage → 503,
unauthenticated → 401, invalid cookie Origin/request marker → 403. Use existing configured origins
and `X-NodeHerder-Request: 1`; permit established bearer auth without cookie-specific CSRF checks.
Allow the relevant headers in CORS for this protected path. Never admit provider work.

HomeLayout fields: `version:1`, `revision: nonnegative integer`, `etag:string`, `blocks:HomeBlock[]`.
PUT sends the same shape; etag/revision are response metadata, server increments its revision.

HomeBlock: `id:string`, `kind:'group'|'card'`, `name:string` (nonempty group label; empty for card),
`cards:HomeCard[]`. Array order defines reading/panel order. A card block has exactly one card.

HomeCard: `id:string`, `kind:'device'|'waste_service'|'waste_summary'`, `panel:boolean`;
device requires `device_id,expose`; service requires `instance_id,resource_id,service_id`;
summary requires `instance_id,resource_id`. Other source fields must be absent/empty. No values,
provider payloads, arbitrary HTML, sizing coordinates or credentials. New cards default panel=true
in the UI; wire false is explicit. Dates remain source-calendar Europe/London and reuse the waste
projection; identity never uses a date/name/index. Unsupported/missing accepted sources render
unavailable and can be removed; new unknown source refs cannot be introduced by a save/import.

Bounds: ≤50 blocks, ≤500 total cards, unique block IDs and card IDs, ID strings 1–256 bytes,
group names 1–60 runes unique case-insensitively, source fields ≤256 bytes. Body bounded by the
existing 1 MiB API limit. Empty layout is valid; cards/blocks must be arrays, not null. No duplicate
card IDs; multiple distinct placements may reference the same source.

Save is atomic; membership/placement writes conflict by the layout ETag, never provider data
revision. Read/mutate through one cache owner; source-validation reads are local and bounded.
Legacy device edits use the same owner once initialized. No Home data is broadcast over WS.
Frontend reads on page entry/focus and publishes the accepted save reply; a conflict retains accepted local state and
offers reload rather than silent overwrite. Logout clears all Home and integration state.

Home export uses an explicit `nodeherder.home/v1` wrapper containing layout references, never
values. Legacy `{dashboardGroups:...}` imports remain supported via a typed conversion against
current accepted layout, preserving unrelated integration placements. Browser validates imported
types/bounds for feedback and server validates all writes. Import replacement is confirmed and
uses If-Match. Source additions are opt-in; setup/refresh remain on existing integration pages.
