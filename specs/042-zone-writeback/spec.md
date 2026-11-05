# Spec: Firmware zone write-back

ID: 042 | Status: Implemented (unreviewed) | Owner: unreviewed draft | Updated: 2026-11-05
Constitution: 1.0.0 (proposed) | Components: backend and frontend

## Problem, scope, exclusions

The spatial tab can apply a **rectangle** to firmware coordinate entities, but a drawn
polygon zone cannot be applied at all, even though Everything Presence Lite v1.5.0 exposes
`Polygon Zone 1..4` / `Polygon Exclusion 1..2` / `Polygon Entry 1..2` text entities that parse
`x:y;x:y;...` (integers, at most 20 points). Users who draw a polygon are told, incorrectly,
that firmware cannot accept it.

In scope: map a polygon zone to one firmware text entity, serialise and validate the points,
send one command, and show requested versus reported values with the firmware-mode caveat. Also
in scope: explicit mappings and writes for installation angle, inverse mounting, maximum
distance, occupancy off delay, polygon mode, and a per-zone off delay.

Exclusions, each a later increment in the same roadmap: calibration (P6); replay (P7);
diagnostics (P8); Home block; fusion.

## Scenarios and acceptance

### US-02 — Write firmware settings with read-back (priority P1)

- AC-09: Given the Firmware settings view, when it renders, then each of the five setting roles
  has an explicit entity mapping (numeric for angle/distance/off delay, binary for
  inverse mounting/polygon mode); nothing is inferred from names.
- AC-10: Given a mapped numeric setting, when a value is entered, then it is validated against
  the entity's advertised bounds and step in its own unit and applied as one command.
- AC-11: Given a mapped switch, when on/off is chosen, then exactly one boolean command is sent.
- AC-12: Given an unmapped role, then it is skipped and no default value is written.
- AC-13: Given a zone with an off-delay entity mapped, then the zone editor writes that number
  only when a value is entered; a polygon or rectangle zone both accept the role.
- AC-14: Given a saved workspace with firmware mappings, when the backend validates it, then
  known roles within the limit are accepted and unknown roles are rejected; cloning keeps the
  mapping independent.
- AC-15: Requested versus reported values are shown for every applied setting; writes remain
  non-atomic and are not presented as persistence proof.

### US-01 — Apply a drawn polygon to firmware (priority P1)

- AC-01: Given a polygon zone, when the physical-zone editor renders, then it offers a single
  polygon text entity mapping (not the four rectangle coordinates); a rectangle zone keeps
  offering only the four coordinate entities.
- AC-02: Given a polygon with 3–20 points inside the firmware envelope, when it is prepared,
  then the serialised value is `x:y;x:y;...` with integer millimetres, the same order as drawn.
- AC-03: Given a polygon with fewer than 3 points, more than 20, non-integer or out-of-envelope
  coordinates, when it is prepared, then no command is produced and the error names the reason.
- AC-04: Given a valid plan and an online device, when applied, then exactly one text command is
  sent; the editor states the write is not atomic, that firmware polygon mode must be enabled,
  and that rectangle bounds do not apply while it is on.
- AC-05: Given the mapped entity is missing, read-only, not a text entity, or rejected by the
  command path, then the plan is blocked with an actionable message and nothing is sent.
- AC-06: Given points are edited or the zone is resized, then the serialised preview updates and
  the previous notice is cleared.
- AC-07: Given a saved workspace with a polygon entity mapping, when the backend validates it,
  then the mapping is accepted; a rectangle zone carrying a polygon role, or a polygon zone
  carrying a coordinate role, is rejected.
- AC-08: Given a rectangle zone, then the existing four-coordinate path, its validation and its
  notices are unchanged.

## Requirements and success

- FR-01: The backend MUST accept one `polygon` entity role for polygon zones only and keep
  rectangle roles rectangle-only. Acceptance: AC-07.
- FR-02: The frontend MUST serialise and validate polygon text exactly; no clamping, rounding
  or silent truncation. Acceptance: AC-02, AC-03.
- FR-03: Apply MUST send one text command through the existing command path and report
  requested-versus-reported state. Acceptance: AC-04, AC-05.
- FR-04: Guidance MUST state the firmware polygon mode requirement and non-atomic write.
  Acceptance: AC-04.
- FR-05: Firmware setting roles MUST be explicitly mapped and validated in native units, with
  requested-versus-reported feedback. Acceptance: AC-09–AC-13, AC-15.
- FR-06: The backend MUST validate and clone the firmware mapping additively. Acceptance: AC-14.
- NFR-01: No new persistence field beyond the existing zone entity map; no inferred firmware
  capability from entity names.
- NFR-02: Polygon editor remains keyboard operable and verified at 375/768/1440 px.
- SC-01: Domain tests cover serialisation, bounds, point-count and plan blocking.
- SC-02: Backend validation tests cover accepted and rejected role/shape combinations.
- SC-03: Existing suites, build, lint and type-checks pass with results recorded.

## Questions and assumptions

- Q-01: Firmware polygon coordinates use millimetres with X -6000..6000 and Y 0..6000
  (the same envelope as rectangle bounds). Assumption recorded; a real-device check will
  confirm the parser before claiming success.
- Q-02: The firmware text entity is `optimistic` and restores its value, so "reported" echoes
  the written text rather than a parsed echo. The UI says so.
- Q-03: Firmware polygon mode switch is not written by this increment; the UI states the
  requirement and P5b adds verified writes for it.
