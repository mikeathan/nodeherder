# ADR-003: Capability-owned semantics, generic integration runtime

Status: selected for implementation under the user's 2026-10-26 instruction.

The current coordinator, snapshots, previews and compiler directly depend on waste schedules.
Moving files alone would not allow another capability to execute without editing core logic.

Use an immutable constructor-injected capability registry. Capability adapters validate their
mapping, normalize provider data, project saved results, describe identities and decide domain
renewal. They do not own persistence, goroutines or HTTP. Runtime owns admission, transaction
publication, identity reservation, credential boundaries and cancellation. Search/select is a
separate reusable flow; the current release does not promise OAuth or arbitrary protocol support.

Opaque versioned resource JSON preserves existing flat waste wire/storage shapes. Concrete
codecs validate domain payloads; heterogeneous maps are limited to external/config boundaries.
Strict common-header decoding and capability errors prevent silent empty/default resources.
No database/AAD migration, new framework, recurring work or new production capability is added.

Alternatives: folder-only moves leave coupling; an untyped map runtime weakens contracts;
a universal plugin/workflow framework adds unnecessary security/operational complexity.

Keep the existing restricted HTTP transport separately packaged; OAuth and assistant clients
have different destinations/policies. Root composition wires waste; runtime never imports it.
The visible product name is Integrations, with old Connections URLs retained as aliases.
Verify synthetic non-waste execution, legacy JSON round trips, generation fences, async replay,
normalization failures, client lifecycle and authenticated desktop/mobile navigation.
