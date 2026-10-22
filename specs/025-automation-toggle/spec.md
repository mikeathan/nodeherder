# Spec: Restore automation toggle commands

ID: 025 | Status: Implemented | Owner: NodeHerder contributors | Updated: 2026-10-22
Constitution: proposed 1.0.0 | Components: both

## Problem, scope, exclusions

The automation editor introduced in PR #54 constructs binary options from only
`values.on` and `values.off`. Zigbee2MQTT's `value_toggle` still reaches the browser
as `values.toggle`, and the Go automation engine still publishes native TOGGLE.
Users cannot select that command for dial/light automations; existing actions show
an empty selection and an Unknown summary.

Restore schema-declared toggle commands in Set values actions, preserve their native
wire value, and verify the complete browser-to-engine-to-MQTT flow using synthetic
devices. Do not change engine semantics, protocols, persistence, device controls,
state comparisons, engine timing, retries, MCP, or household configurations.
The real-engine creation test also exposed that the editor sends a zero delay,
which the backend rejects. Normalize that UI representation of no wait by omitting
the optional delay at the shared automation save boundary.
Race-instrumented integration also reproduced concurrent dirty-callback registration
and device storage during startup. Synchronize that callback registry without
changing notification order, holding locks during callbacks, or adding allocations
to each device update.
Also preserve an empty trigger list when decoding a saved automation: its current
conversion to null crashes the editor when a user saves after deleting every trigger.

## Scenarios and acceptance

### US-01 — Configure a dial to toggle a light (P1)

- AC-01: Given a binary expose with a distinct, nonempty string `values.toggle`,
  Set values offers Toggle alongside On/Off, with an accessible keyboard-operated
  control at desktop 1280×800 and mobile 390×844 widths.
- AC-02: Saving and reloading preserves the exact native command, and the action
  summary says Toggle. Existing on/off, boolean, numeric, and enum values retain
  their types and behavior.
- AC-03: Creating the automation in the UI and delivering two identical dial events
  through the real Go hub publishes exactly two native TOGGLE commands to the
  synthetic light. Separate MQTT confirmations show On then Off in the browser.

### US-02 — Preserve state and capability boundaries (P1)

- AC-04: Missing, empty, malformed, or on/off-equivalent toggle metadata adds no
  command. Devices without toggle retain their existing controls. Conditions offer
  reported On/Off states only; a command is never formatted as confirmed state.
- AC-05: Existing unchanged automation payload round trips, disabled automation,
  ignored dial events, offline warnings, and confirmed-state handling remain covered
  by existing tests. No new command retries or optimistic light reports are added.
- AC-06: No wait (zero delay) saves as an absent optional delay; positive delays
  retain their value/unit. Existing valid payloads round trip unchanged, and the
  editor's local draft is not mutated by serialization.
- AC-07: Concurrent device writes and dirty-callback registration are race-free;
  callbacks may register another callback without deadlock. Newly registered callbacks
  participate in subsequent notifications; existing callbacks retain their order.
- AC-08: Saving an automation with no triggers and reopening/reloading it keeps
  the editor usable with an empty list. Decoding a replacement list replaces the
  previous list; invalid trigger input does not partially change the current list.

## Requirements and success

- FR-01: Restore only device-declared toggle commands in action inputs/summaries;
  acceptance: AC-01, AC-02, AC-04.
- FR-02: Preserve transport values and command/report separation; acceptance:
  AC-02, AC-03, AC-05, AC-06, AC-07, AC-08.
- NFR-01: Option construction remains constant-sized and synchronous; no new
  production subscriptions, timers, requests, dependencies, or stored state.
- NFR-02: All automated device commands target in-memory MQTT. Browser fixture
  servers bind loopback, require established authentication for controls, use bounded
  waits, and shut down their owned resources.
- SC-01: Regression tests fail before the fix; domain suites and browser-created
  automation pass afterward on both viewport projects, alongside component checks.

## Entities and external interfaces

Expose metadata is owned by the backend's existing bridge conversion. Automation
`exposes[].data` remains its native JSON value. WS saveAutomation and MQTT /set
payloads do not change. A toggle is a command, not a measured binary state.
Persistence, metrics, retention, and deployment configuration are unchanged.

## Questions and assumptions

No blocking questions. Zigbee2MQTT currently supplies toggle as a string; reject
other types instead of inventing capabilities. Synthetic fixtures validate transport
and execution without authorizing tests against household devices.
