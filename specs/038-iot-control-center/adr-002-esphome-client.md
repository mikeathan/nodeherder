# ADR 002: Maintained Go ESPHome client patch

Status: Proposed, user-selected direction; implementation and review pending | Date: 2026-10-30
Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Rules: NH-03/04/05/06, BE-03

## Context and decision

The user requires Go and selected patching [mycontroller-org/esphome_api v1.4.0](https://github.com/mycontroller-org/esphome_api/tree/v1.4.0). [Local probes](client-evaluation.md) found a failed-handshake socket leak, a timestamp data race, and process termination on an oversized plaintext frame. The dependency is not safe unchanged. Its Apache-2.0 license and generated protocol material must retain notices and attribution.

Working design: an in-repository, separately versioned Go module at `backend/third_party/esphome_api/`, based on the exact v1.4.0 source and pinned commit; the backend uses a local `replace` for this module. This makes the patch reviewable, deterministic in offline builds and testable without publishing a fork. Record upstream revision, original license/NOTICE, patch rationale and an upgrade/owner checklist alongside the module. Verify Go module discovery, nested-module test invocation, Docker build context, CI, vulnerability scanning and SBOM before merging. Do not silently upgrade the upstream protocol or assume generated messages are supported merely because bindings exist.

The actual v1.4.0 module declares Go 1.26, although the upstream main-branch page previously inspected declared Go 1.21. NodeHerder targets Go 1.24.8. Test the imported library subset under 1.24.8 before lowering the fork's directive; if incompatible, evaluate a documented toolchain migration across Docker and CI rather than claiming compatibility or silently upgrading.

Patch only evidenced transport/lifecycle defects and the gaps required by the client gate: close sockets on every setup/handshake failure; synchronize shared state; bound lengths before allocation, validate varints and encrypted payloads; provide context-cancellable dial/handshake/read and idempotent bounded close with explicit ownership; serialize concurrent writes/response waiters; handle unknown message types safely. Preserve compatibility where feasible and document unavoidable API changes. Avoid exposing secrets, raw payloads or panic messages in diagnostics. A wrapper is insufficient where allocation or socket ownership is hidden inside the library.

## Alternatives and consequences

- Unmodified v1.4.0: reject due to reproduced failures.
- A newer or less established Go client: no candidate has demonstrated the user's maturity requirement; reassess if evidence changes.
- External maintained fork: allows upstream collaboration but introduces an externally hosted dependency; a separately pinned published fork can replace the local module later after review.
- New transport from scratch: control at the cost of owning protocol and security compatibility; reject unless the patch proves impractical.
- Python companion or MQTT: excluded by the user's Go and native-API direction.

NodeHerder owns patch review, release pinning, upstream contribution assessment and compatibility testing until the changes are available in a verified upstream release. A fork is not automatically mature. No upstream issue, pull request, license determination or maintainer approval is claimed. Never publish or send changes externally without explicit authorization.

## Verification and reconsideration

Turn each local probe into a deterministic regression owned by its patched source file. Test plaintext/Noise handshake success and rejection with fake peers, cancellation at each phase, stalled peer, oversized/truncated/unknown frames, concurrent send/receive/close, state subscriptions, request-response ordering and version negotiation. Run `go test -race` and fuzzing with bounded resources in the nested module, then integration tests under the project toolchain. Compare independent ESPHome protocol fixtures and explicitly record separately scoped real-device evidence. Reconsider when a fixed, independently vetted upstream release or a demonstrably more mature Go alternative is available.
