# Spec: <name>

ID: <id> | Status: Draft | Owner: <author> | Updated: <YYYY-MM-DD>
Constitution: <version> | Components: <backend/frontend/both>

## Problem, scope, exclusions

User/problem/outcome; affected users/devices/data; current code/test/issue evidence.
Separate facts from assumptions. Specify outcomes before libraries/internal design.

## Scenarios and acceptance

### US-01 — <user value> (priority <P1/P2/P3>)

Independent verification: <method>.

- AC-01: Given <state>, when <action>, then <observable outcome>.
- AC-02: Given <failure/boundary>, when <action>, then <safe outcome>.

Cover applicable offline/repeated-event/ambiguity/auth/empty-data/recovery scenarios.
Preserve IDs for traceability.

## Requirements and success

- FR-01: MUST <functional outcome>; acceptance: <AC IDs>.
- NFR-01: MUST <measurable quality target and conditions>.
- SC-01: <measurable success outcome, conditions, verification>.

Specify applicable units/timezones/retention/trust/device-command/accessibility/
compatibility concerns; justify omissions. Avoid vague "fast"/"secure" claims.

## Entities and external interfaces

Meaning/ownership/behavior, current consumers, compatibility period; no internal classes.

## Questions and assumptions

Each question: consequence, owner, resolution. Mark assumptions to validate.
Unresolved safety/security/contract/acceptance questions block dependent work.
Replace prompts with content or justified N/A before review.
