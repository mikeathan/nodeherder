# ADR-001: Observation source capability mapping

Status: accepted (maintainer decision 2026-11-11) | Date: 2026-11-11
Spec/plan: [spec.md](spec.md), [plan.md](plan.md) | Owner/review: maintainer decision recorded |
Rules: NH-04, BE-01, BE-05, FE-01, FE-02

## Context and options

A new observing source kind must appear in the shared viewer without UI changes. The hub
needs to answer "does this enrolled device observe a medium?" without knowing its
provider name.

Existing repository precedent derives a capability from **data shape**
(`frontend/src/domain/spatial/index.ts:188`, `supportsSpatialMapping`), consumed by
`deviceTabs()` (`frontend/src/domain/devices/page.ts:36`). A competing precedent gates
on provider name (`isESPHomeDevice`, `frontend/src/domain/devices/index.ts:78`;
`deviceTabLabel(tab, isESPHome)`, `frontend/src/domain/devices/page.ts:65`). The latter
works only because ESPHome is currently the sole provider.

Options:

1. **Provider-name gate.** Add `provider === 'esphome'` branches. Cheapest now, but every
   future source edits the viewer, navigation, and filters — the exact outcome this work
   is meant to prevent.
2. **Additive capability field on the device.** The provider declares
   `capabilities: ["wireless.observe"]` at registration. Consumers read the field.
   Requires touching `models/devices/Device`, but keeps consumers provider-agnostic.
3. **Provider-side capability registry.** A registry lookup keyed by provider name.
   No device-model change, but it reintroduces name coupling at every call site and puts
   UI concerns in the provider layer.

Backend seam options, independently: widen the existing `Sink` interface
(`session.go:24`) versus an optional `Observer` interface type-asserted at runtime.
Widening forces every sink (including test doubles) to implement an unrelated concern
and violates single responsibility. The optional-interface idiom already exists in
`deviceproviders` (`HealthReporter`, `IdentityBinder`, `InventoryBinder`).

## Decision and consequences

Choose **option 2** plus an **optional `Observer` interface**. The concrete field is:

```go
// Capabilities lists abstract features this device provides, e.g. "wireless.observe".
Capabilities []string `json:"capabilities,omitempty"`
```

Capability detection is structural (`instance.(Observer)`) on the backend and field-driven
(`supportsObservation(device)`) on the frontend. Neither compares provider names.

Rejected alternative considered during review: gate navigation on the live
`observations` snapshot (`sources.length > 0`) and skip the model change entirely. It is
cheaper and auto-covers future sources, but it reports *has observed* rather than *can
observe*, so an enrolled but silent source would show no UI and no empty state. Rejected
for non-determinism.

Consequences:

- One additive field on `devices.Device`; existing payloads are unchanged because the
  field is optional and omitted when empty.
- Adding a source kind is one medium constant plus one decoder — no viewer, route,
  navigation, or filter change.
- The two existing provider-name gates remain. Out of scope; recorded as debt so they
  are not treated as precedent.
- UI composition is bound in one place, following
  `frontend/src/integrations/registry.ts:17`.
- Provider devices are already excluded from the unauthenticated projection
  (`backend/internal/public_devices.go:28`), so capabilities cannot leak to MCP.

Verification: capability predicate unit tests; a fake second source kind proving the
viewer renders it with no component change.

## Supersession

None. Recorded now to pre-empt the provider-coupling alternative.
