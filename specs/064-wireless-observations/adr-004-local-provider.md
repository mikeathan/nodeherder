# ADR-004: Wireless observation is an ordinary integration, backed by a local provider

Status: proposed | Date: 2026-11-13
Spec/plan: [spec.md](spec.md), [plan.md](plan.md) | Owner/review: none recorded |
Rules: BE-01, BE-04, FE-01, FE-05. Supersedes the device-backed catalog note in
[ADR-003](adr-003-observation-viewer.md).

## Context and options

Wireless observation shipped as a **device-backed** integration type: it appeared in the
catalog whenever a capable device was enrolled, was always present on the board, and could only
be hidden with a browser-local preference. Two consequences made that untenable:

1. **It could not be removed.** Every other integration is an instance the operator adds and
   removes. This one was a permanent feature of the page, with no way back.
2. **It was always on.** The device subscribed to BLE advertisements whenever it could, whether
   or not anyone wanted the data. On the EPL, BLE shares RAM and radio time with the mmWave
   sensor, so "always on" has a cost paid by another sensor.

A YAML provider definition was investigated first and is **impossible**:

| Requirement | Where | Effect |
| --- | --- | --- |
| `base_url` must be `https` on port 443 | `destination/url.go:26` | a local address cannot pass |
| 1–8 HTTP requests; zero is an error | `definitions/compiler.go:257` | a fetchless provider is rejected |
| 2–16 inputs, a setup and a refresh recipe | `compiler.go:164` | nothing to ask, nothing to fetch |

Relaxing those rules would make a validator that describes the format into one with
exceptions, and would ship a provider that can never produce a resource.

Options:

1. **Keep it device-backed; add a hub-stored enable flag.** Smallest change. But there is still
   no add or remove, so it does not behave like an integration.
2. **A local provider: code-defined, no request (chosen).** It is a real provider with a real
   instance, so add, remove, enable and the board all work through the existing paths.
3. **Allow a definition with zero requests.** Rejected: it changes what the grammar means for
   every provider, to serve one.

## Decision and consequences

Choose **option 2**. The distinction is one field, `Definition.Local`, defaulted false:

- **Built in Go, never parsed from YAML.** `CompileLocal` takes the capability rules that
  actually guarantee a resource; `CompileYAML` refuses a document claiming `local`, so an
  operator cannot author a provider with no source.
- **No operation.** `publishLocal` reads a value the hub already produced and normalises it
  through the ordinary capability contract, so nothing downstream knows the difference.
- **No schedule.** Its value is the hub's current state, so a timer would republish the same
  census. `UpdateSchedule` refuses one rather than accepting a control that never fires; the
  scheduler also skips local instances, which covers a hand-edited record.
- **Disabling stops the pipeline.** `observations.Deps.Active` is asked per advertisement batch;
  a disabled integration drops them before they reach the queue or the evidence recorder. The
  device keeps its session and its entity state — this is a data decision, not a device-lifecycle
  one. Disabling clears the stored snapshot, because it describes a radio no longer being read.
- **Not added by default.** The provider appears in the catalog; no instance exists until the
  operator adds one. An installation with no instance observes nothing.

Consequences:

- **The frontend loses code.** `device-backed.ts`, `observation-visibility.ts`, the special
  `ObservationCard` and the board's bespoke group are all deleted; the type is now an ordinary
  catalog entry and an ordinary board row.
- **`Definition` gains a concept.** It is one bool and one code path, but it is a genuine
  extension of a model built around remote providers. That is why this ADR exists.
- **A consumer reading the raw backend provider list now sees the type**, which the
  device-backed approach could not offer. Catalog availability means "can be added", not "has a
  live source", so it does not blink when a sensor is briefly unreachable.
- **The disable path has a silent failure mode.** A bug there is a gap in observation that looks
  like a quiet room. It is covered by a test that asserts the queue stays empty, rather than
  trusting the flag.
