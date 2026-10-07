# Spec: Ordered device message processing

ID: 005-device-message-ordering | Status: Draft | Owner: mikeathan (drafted with Claude) | Updated: 2026-10-07
Constitution: parent 1.0.0 (proposed); backend 1.0.0 (proposed) | Components: backend

## Problem, scope, exclusions

**Problem.** Toggling a device from the UI sends the command, the device changes, and the UI
switch snaps back to its previous position. Automation-driven changes look fine. It is
reproducible on every click, per device and direction. The user reports it was fine for over a
year and began after the 1 October work.

**Observed facts (live system, 2026-10-07, read-only capture).**

- F-01: Zigbee2MQTT 2.14.2 (image `koenkk/zigbee2mqtt`, unpinned; release dated 2026-10-01)
  publishes **two** messages on `zigbee2mqtt/<device>` for one `/set`, in the same millisecond:
  (1) the cached *old* state with a fresh `last_seen` (`linkquality` 0 after restart),
  (2) the *new* state. Source: Zigbee2MQTT debug log (`MQTT publish` lines) and its WebSocket feed.
- F-02: NodeHerder broadcast the pair in the opposite order (`ON`, then `OFF` after an OFF→ON
  click), so the UI ended on the stale state while the light physically changed.
- F-03: The backend receives each MQTT message on its own goroutine
  (`SetOrderMatters(false)`, `backend/internal/mqtt/mqtt.go`), then hands it to a 4-worker pool
  (`HubController.processMessage`). Nothing orders messages for one device. A concurrent replay
  of "stale then new" reversed the final state in 1993/2000 trials on the pre-#51 code and
  1992/2000 on current code: the backend defect predates #51; Zigbee2MQTT's new output exposes it.
- F-04: Automations run synchronously inside `DeviceLifetimeService.Update`
  (`HandleDevice` → `Evaluate`). Conditions read the device's **live stored** exposes
  (`ExposeHandler.Evaluate` → `ctx.GetDevicePayload`), not the message delta, and step actions read
  the dial's stored `action_time` and the target light's stored `brightness`
  (`stepOperation.CreatePayload`). Two dial messages processed concurrently can therefore see each
  other's `action`/`action_time` values.
- F-05: `Device.Evaluate` and `EvaluateTrigger` mutate one shared `DeviceContext` per automation
  (`SetDevicePayload`, `SetManualTrigger`). A manual trigger and a device message evaluated
  concurrently can overwrite `isManual`, which gates unconditional triggers
  (`DeviceTrigger.Process`).
- F-06: Not hazards (verified in code): schedules only call `SetEnabled`, which is mutex-guarded;
  `MqttBaseAction.executeBase` holds the action's mutex for the whole `processAction`, so
  rotate/step operation state is already serialised per action; `ctx.currentData` is read only by
  `ExposeCondition.HasValueChanged` and `resolveValue` and does not influence published commands.

**Automations in use** (`backend/configs/automations/`, four files): a dial that toggles, steps
brightness and cycles colour-temp presets on the Living Room Light; a presence sensor that turns
that light on (with an illuminance condition) or off after a 5-minute delayed action; a
door-sensor alarm (batch publish); and a disabled test automation with enable/disable schedules
and an unconditional manual-only trigger.

**Scope.** Backend path from MQTT/HTTP ingestion to device update, automation evaluation,
storage and WebSocket broadcast; manual-trigger entry.

**Exclusions.** Frontend; Zigbee2MQTT version/config; automation JSON schema; step arithmetic;
schedule semantics; metrics storage; reconciling the optimistic `ctx` state with device state.

## Scenarios and acceptance

### US-01 — UI/command toggle shows the true final state (priority P1)

Independent verification: replay harness with a fake Zigbee2MQTT that answers `/set` with the
stale-then-new pair; live manual check on named devices after approval.

- AC-01: Given a light in either state, when a command produces `[old state, new state]` in one
  millisecond, then the stored state and the last `deviceUpdated` broadcast equal the new state in
  every one of 2,000 randomised-scheduling trials (both directions).
- AC-02: Given `[state S, state S]`, then at most one broadcast occurs and the final state is S.
- AC-03: Given a slow handler for device A, then messages for device B are processed without
  waiting for A.

### US-02 — Dial control stays accurate under rotation bursts (priority P1)

Independent verification: recorded dial traces replayed with a fake clock; compared with a
baseline recorded from current code for non-concurrent sequences.

- AC-04: Given N alternating `dial_rotate_*` events within a burst, then each evaluation sees its
  own `action` and `action_time` and publishes the brightness command its own values imply.
- AC-05: Given a non-overlapping event sequence, then published commands equal the recorded
  baseline for slow/fast, left/right, clamped at min and max.
- AC-06: Given identical consecutive `action` events, then every event is processed (NH-02).
- AC-07: Given `button_1_press_release`, then `state: TOGGLE` is published once per press; given
  `button_2_press_release`, then colour-temp presets cycle in order and wrap.

### US-03 — Manual triggers are isolated from device messages (priority P1)

- AC-08: Given an unconditional trigger and device messages arriving concurrently, when a manual
  trigger fires, then its actions run.
- AC-09: Given the same unconditional trigger, when only device messages arrive (even during a
  manual trigger), then the trigger does not run.
- AC-10: Given a disabled automation, an unknown trigger name or a failing action, then the
  caller receives the same error as today, and a lane that cannot accept work within the request
  deadline returns a distinct, bounded error.

### US-04 — Delayed actions and cooldowns keep their semantics (priority P1)

- AC-11: Given `presence=false`, then OFF is published after the configured delay (fake clock).
  Given `presence=true` before expiry, then the pending OFF is cancelled. Given repeated
  `presence=false`, then the delay is not restarted (current behaviour, recorded as baseline).
- AC-12: Given `presence=true` with `illuminance <= 45`, then ON is published; with a higher value,
  nothing is published.

### US-05 — Enable/disable schedules and lifecycle (priority P2)

- AC-13: Given enable/disable schedules, then messages inside the window evaluate and messages
  outside it are skipped, as today.
- AC-14: Given shutdown, then lanes drain or cancel within a bounded time, with no leaked goroutines.
- AC-15: Given HTTP data-collector ingestion (`DataCollectorHandler`), then it uses the same
  ordering path as MQTT.
- AC-16: Given bridge topics (`bridge/devices`, responses, logging), then their relative order is
  preserved and a blocking subscribe in a handler cannot deadlock MQTT delivery.

## Requirements and success

- FR-01: MUST process messages for one device in arrival order, end to end (update, automation,
  store, broadcast); acceptance: AC-01, AC-02, AC-04.
- FR-02: MUST NOT make unrelated devices wait on each other; acceptance: AC-03.
- FR-03: MUST evaluate each automation run against an isolated run context (own origin: device
  message or manual) layered over persistent automation state; acceptance: AC-08, AC-09.
- FR-04: MUST route manual triggers through the same per-device ordering; acceptance: AC-08, AC-10.
- FR-05: MUST preserve current automation semantics (step maths, preset cycling, TOGGLE,
  delay/cancel, conditions, schedules, repeated events); acceptance: AC-05–AC-07, AC-11–AC-13.
- FR-06: MUST funnel MQTT and HTTP ingestion through one entry point; acceptance: AC-15, AC-16.
- FR-07: MUST define ownership, bounded queues, panic isolation and shutdown for new goroutines;
  acceptance: AC-14.
- NFR-01: Under a 1,000-message mixed burst across 20 devices, p95 handling latency and
  allocations per message MUST be no worse than baseline +10% (baseline measured in T-task before
  any change; absolute target set from it).
- NFR-02: Idle goroutine count MUST equal baseline (no goroutine per idle device).
- NFR-03: `go test -race` on affected packages MUST pass.
- NFR-04: Queue depth per device MUST stay below the configured cap in the burst above; overflow
  MUST be logged, counted and rejected, never silently dropped.
- SC-01: AC-01 passes 2,000/2,000 with the replay harness; the same harness fails on current code.
- SC-02: Live: ten UI toggles each way on the named light leave the UI matching the device.

## Entities and external interfaces

- Message: topic + payload from MQTT or HTTP data collector; ordering key is the device topic.
- WebSocket `deviceUpdated`, HTTP API, MQTT topics and automation JSON are unchanged.
- Internal: `AutomationContext` loses `SetManualTrigger`/`SetDevicePayload` (replaced by a per-run
  context); consumers are conditions, actions and tests/mocks only.

## Questions and assumptions

- Q-01 (blocks no design, may add a filter): does Zigbee2MQTT's `last_seen` cache publish ever
  re-emit `action` for the dial? If yes, event exposes need a snapshot filter because events
  bypass dedupe. Evidence 2026-10-07: Zigbee2MQTT's cached state for "Living room switch dial"
  (sent to a new frontend client) contains only `battery`, `last_seen`, `linkquality` and `update`;
  no `action`/`action_time`, so snapshots built from it cannot carry events. Status: likely
  resolved; confirm with one read-only capture while the dial is rotated (the live check in SC-02).
- Q-02: RESOLVED 2026-10-07 (user): keep "step from last confirmed target brightness" exactly as is.
- Q-03: RESOLVED 2026-10-07 (user accepted recommendation): per-device lane cap 1,000 queued
  messages; beyond it reject the new message, log loudly and count it; never drop silently.
- Q-04: RESOLVED 2026-10-07 (user accepted recommendation): live manual checks may use only the
  Living Room Light (dial and presence sensor) and the Attic room Light (UI toggle, ten each way).
  No other devices. Routine automated tests never command household devices (NH-05).
- A-01: Zigbee2MQTT publishes the pair in the order old, new (observed on four toggles).
- A-02: Household rate is below ~100 messages/s; dial bursts are the peak.
