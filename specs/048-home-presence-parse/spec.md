# Spec: Decode Home presence cards

ID: 048 | Status: Fixed (unreviewed) | Updated: 2026-11-05
Constitution: 1.0.0 (proposed) | Components: frontend Home contract

## Problem

Adding a Home presence card saved successfully on the server, then the browser rejected the
layout response as invalid. Because the rejected card is persisted, every later Home read failed
the same way: the dashboard showed "Home returned an invalid layout. Existing areas are kept."
permanently and layout editing could not proceed.

Cause: `contracts/home/index.ts` decodes card kinds with an explicit `device` case and treats
every other kind as an integration reference. A `presence` card carries `device_id`, so the
integration branch rejected it as a mixed reference. The server-side layout validation and
admission already accept `presence` (spec 046), and the response parser was not updated.

## Scope and acceptance

- AC-01: A stored `presence` card with `device_id` and no expose/integration fields parses into a
  `presence` card, so existing broken layouts load again and can be edited or cleaned up.
- AC-02: A `presence` card carrying `expose`, `instance_id`, `resource_id` or `service_id` is
  rejected as ambiguous, matching server validation.
- AC-03: Device and integration cards keep their current decode behaviour.
- AC-04: No server, storage or migration change; the fix is a response-decoding correction.
- AC-05: Contract tests cover accepted and rejected presence shapes.

## Requirements and success

- FR-01: Add the missing decode branch; keep bounds/control-character checks.
- SC-01: New contract tests plus the full frontend checks and Home browser spec pass.

## Questions and assumptions

- Q-01: The card reached storage because the server accepted it; recovery needs no data change,
  only the corrected decoder. Verified by reproducing the parser rejection against a
  server-shaped card in tests.
