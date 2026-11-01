# ADR 001: Device ownership, provider routing and private exposure

Status: Proposed, not approved | Date: 2026-10-30

## Context

The existing device store and `DeviceLifetimeService` own confirmed device state and automation/event publication. UI commands and automation actions are currently routed directly to MQTT, while `/api/context/devices` and MCP handlers are registered public and read devices from that same store. ESPHome requires a persistent, bidirectional native TCP session and private key material. A second authoritative device store would create inconsistent device truth; adding new device data to public consumers without a policy would expand household-data exposure.

## Proposed decision

1. Statically register providers and route a command by a stable provider key on the existing device identity. Define small consumer-owned interfaces for lifecycle and command dispatch; expose discovery only on providers that support it. Avoid dynamic plugins and a mandatory universal interface.
2. Reuse the existing device registry and observation/automation pipeline after adding a protocol-neutral registration boundary. New providers must not impersonate Zigbee bridge records or derive identity from mutable friendly names. Stable identity derives from a persisted provider-scoped device ID and entity IDs from the provider's stable keys; reject collisions and duplicate onboarding. Preserve existing Zigbee IDs and persisted automations unchanged.
3. Route browser and automation commands through the same validation/dispatch service. Resolve the current confirmed device, provider and writable capability; validate type, bounds and options; reject unsupported or offline commands. Never optimistically overwrite confirmed observations or automatically repeat an ambiguous command. Keep the existing Zigbee command payload and ordering behavior covered by regression tests.
4. Default all newly onboarded provider devices to **private from existing public context and MCP resources**. Keep authenticated hub/WS consumers, existing public Zigbee behavior and existing routes compatible. Enforce filtering at each public projection and lookup, not only in the browser; no secret, address, entity name, state, historical metric or automation-access path may leak through public consumers. An explicit opt-in or authenticated MCP redesign requires a separately specified review, not an implicit promotion.
5. Store per-device credentials in a dedicated encrypted repository with separate lifecycle and keyring from the optional HTTP integration runtime. The device provider must degrade independently on missing keys; existing Zigbee startup must remain available. Use existing AES-GCM/keyring patterns only after verifying key-file provisioning, access control, identity-bound AAD, rotation, backups, schema compatibility and lost-key recovery. Never serialize credentials in shared `Device` or browser response DTOs.
6. Manual onboarding accepts a numeric private unicast address and explicit TCP port; disallow loopback, link-local, multicast, unspecified and public addresses in production. Use a dedicated injectable dialer and resolved address, never a caller-supplied URL or DNS redirect. Discovery offers candidates but must not connect, trust, or enroll them automatically. Numeric addresses deliberately avoid DNS rebinding; mDNS may produce an address candidate subject to the same policy. Tests inject loopback dialers instead of weakening production policy. Reconsider IPv6/other local networks explicitly, rather than guessing.

## Alternatives and consequences

- Reusing the HTTP integration service directly would make ESPHome depend on its separate startup, definitions and single database lifecycle; reject. Shared cryptographic mechanisms may be extracted when ownership is explicit.
- Routing only UI commands through an adapter leaves automation MQTT-specific; reject. A compatibility shim must cover both producers.
- Exposing new devices through today's public routes silently expands sensitive data; reject. Filtering existing public projections adds code but preserves current consumers.
- Allowing arbitrary DNS names or public TCP destinations broadens an authenticated configuration endpoint into a network probe. Numeric private unicast is a conservative first boundary; the exact IPv4/IPv6 ranges and Docker deployment behavior require tests and review.

## Required evidence and remaining decisions

This ADR is a proposal, not maintainer approval or an exception. Prove device identity stability, no public-data leakage, key loss/rotation/backup, duplicate onboarding, offline commands, provider isolation, shutdown and repeated-event ordering. Define precise network ranges, persistence API/schema, reconnect/queue bounds, source capability normalization and command/automation migration in the contract before implementation. The candidate native API client fails isolated socket-cleanup and race checks, and a safe implementation or replacement must also bound hostile frames and support cancellation.
