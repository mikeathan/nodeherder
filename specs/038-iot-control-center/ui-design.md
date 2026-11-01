# Control center: UI design

Status: scoped design direction, not a completed visual or accessibility review. **User constraint:** preserve all existing layouts, colors, typography, navigation and visual designs; add ESPHome support within the current device experience. Do not introduce new general page layouts or redesign Zigbee UI. Reuse `DeviceCard`, `DeviceFilterBar`, `ExposeRow`, `ExposeControl`, `UiCard`, `UiButton`, `UiDialog`, `UiSelect`, `UiRange` and `useDeviceCommand` where their contracts fit. An ESPHome provider label/icon must accurately distinguish it from Zigbee without changing existing device icons.

This document is the remaining UI guidance for the feature. The historical exploratory prototype was removed; it predated the no-redesign direction and was never implementation guidance. Only the provider-specific concepts below remain in scope.

## Devices overview

```text
Devices                                    [ Add device ]
Search devices…     [ Any status ] [ Protocol ] [ Power ]    8 of 12

┌ Living room sensor ───── Online ─┐  ┌ Hall light ───── Offline ───┐
│ Presence       Detected          │  │ Last seen 12 minutes ago     │
│ Temperature    21.4 °C           │  │ Controls unavailable         │
│ [ Open device ]                  │  │ [ Open device ]              │
└ ESPHome · last seen now ────────┘  └ Zigbee ──────────────────────┘
```

Keep the existing card/list layouts, navigation, search and filters exactly as they are. Present an ESPHome device through those components with its correct provider identity and existing status treatment. Distinguish loading, failure and offline using the existing patterns. Online state must derive from confirmed health, never from a stale sensor value. The sketch above describes content only, not a layout change.

## Onboarding and discovery

Add an **Add ESPHome device** action beside the existing Devices header actions, using the current button style. It opens the ESPHome setup directly; there is no generic provider picker for this release. Zigbee devices continue to be paired through the existing **Permit join** flow. For ESPHome: discover candidates where supported, or enter a numeric private IP and port manually; display candidates as unverified and require explicit selection. Ask for the API encryption key in a password field when configured; explain plaintext selection without displaying stored keys. Validate input locally for feedback and on the backend for enforcement. Saving creates a pending enrollment and starts connection attempts asynchronously, so an unavailable device can be configured before it is online. Open Manage ESPHome immediately after saving so the user can see progress and correct or remove the enrollment; detailed key-versus-network diagnostics are not currently distinguished. Editing and deleting explain whether saved metrics remain, with confirmation for deletion.

```text
Add device   ›   ESPHome
Address     [ 192.168.x.x       ]   Port [ 6053 ]
Security    (•) Encrypted   ( ) Unencrypted
API key     [ •••••••••••••••• ]  [ Show ]
                 [ Cancel ]  [ Connect and inspect ]
```

## Device detail and controls

Use the existing `frontend/src/components/device/DevicePage.vue` layout and its current navigation, sections and styles. Present discovered readings and supported controls through existing entity and settings patterns. Add only provider-specific connection information and entity types that do not fit existing components. A provider-specific extension may contribute content through an explicit registry, never by branching on device name. The sketch below illustrates possible data and is not a new tab or page-layout specification.

```text
← Devices   Living room sensor                 Online · seen just now
            ESPHome                            [ Settings ]
[ Overview ] [ Entities ] [ History ] [ Diagnostics ]
Presence       Detected        Temperature       21.4 °C
Light          [ switch ]      Sensitivity       [ 0 ──●── 100 ] [ 45 ]
Radar          [ Open spatial view ]
```

Use the existing metadata-driven renderer and a small explicit renderer registry for new types. Binary sensors are read-only indicators; writable switches, numbers, selects, buttons and supported light features use capability-specific controls. Unknown entities render safe text/metadata, not an invented command. Numeric controls have accessible names, bounds and a keyboard-operable numeric alternative. Pending commands show a pending indicator while the displayed value remains confirmed; failures explain whether a request was sent and do not automatically repeat it.

Provider button actions require confirmation because discovery does not provide a
reliable safe/destructive classification; the first real inventory includes reboot
and factory-reset actions. The confirmation uses the discovered label and existing
dialog pattern, without changing legacy device controls.

## Spatial detail

Only show a spatial section when a verified or explicitly configured mapping supplies valid coordinates and target semantics. A protocol-neutral spatial component consumes normalized millimetre coordinates, validity and timestamps. Default plane: sensor at bottom centre, positive Y forward, X left/right; grid and scale are labelled. Orientation and range are user-configurable. List each target with ID and X/Y/distance in text as an accessible alternative. A target disappears only on evidence of inactivity/unavailability or a documented stale timeout; partial coordinates show unknown, and offline preserves last-seen context without presenting targets as live. Separate rendering cadence from event processing; no history is inferred or stored.

```text
Radar                           [ Scale ▾ ] [ Orientation ▾ ]
     6 m  · · · · · · · · · ·          Target 1   X −0.8 m   Y 2.4 m
          ·     ● 1       ·            Target 2   X  1.1 m   Y 3.0 m
          · · · · · · · · ·
     0 m        Sensor
```

## Responsive and accessible behavior

Preserve existing responsive behavior and use its existing breakpoints and theme tokens. Any new onboarding or radar content must fit at 1440, 768 and 375 px without changing existing page layouts or causing horizontal scrolling. Use semantic headings/labels, visible focus, logical tab order, status text and error announcements; do not rely on color or motion. Minimize animated movement and respect reduced-motion preferences. Verify keyboard and screen-reader behavior for dialogs, toggles, selects, number input and error recovery. Implement unit formatting through `src/utils/units.ts`.

Verification must use isolated mocked providers and fixtures: desktop/mobile screenshots and interactions at 1440, 768 and 375 px, loading/empty/offline/reconnecting/failure cases, pending versus confirmed commands, keyboard/focus, stale targets and large update streams. These checks are planned, not yet performed.
