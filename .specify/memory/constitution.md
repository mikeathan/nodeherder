# NodeHerder Constitution

Version: 1.0.0 | Status: Proposed for adoption | Created: 2026-10-01
Ratified: Not yet ratified | Last Amended: 2026-10-01

## Purpose and authority

NodeHerder's Go/Vue smart-home system integrates MQTT/Zigbee2MQTT and assistant/MCP
services. These principles protect physical-device behavior and private household data.

MUST/MUST NOT are mandatory; SHOULD departures require written justification and
review; MAY is optional. Rules govern new/changed behavior, not a claim of existing compliance.

Authority: this constitution → [backend](../../backend/CONSTITUTION.md)/
[frontend](../../frontend/CONSTITUTION.md) rules → specs → plans/ADRs → tasks.
Component rules refine, never weaken, the parent; both apply across components.
Agent instructions and setup guides are subordinate. Conflicts MUST be resolved
before implementation; lower-level documents cannot override higher-level rules.

## Core principles

### NH-01 — Specify outcomes first

Behavior changes MUST have a version-controlled spec before production-code edits:
user problem, scope/exclusions, independently verifiable acceptance scenarios, failures,
and measurable success criteria. Specify outcomes before implementation choices.
Questions affecting safety, security, contracts, or acceptance MUST be resolved before
dependent work. Small fixes MAY use short specs but MUST include reproduction and
expected behavior. Documentation, formatting, and mechanical changes MAY use an
issue/PR description recording scope and verification.

### NH-02 — Preserve device semantics and user control

State, physical events, and command acknowledgements MUST remain distinct. Repeated
events MUST NOT be discarded by value equality. Accepted state MUST reach consumers
before sampling/debounce; metrics/UI throttling MUST NOT suppress automation inputs.
Device-affecting changes MUST specify targets/values, offline behavior, failure feedback,
retry/idempotency, and applicable feedback-loop prevention. Submitted commands MUST NOT
be shown as confirmed state without evidence. Timing changes MUST specify automation behavior.

### NH-03 — Protect trust boundaries and household data

New network entry points MUST use established authentication unless public exposure
is explicitly specified and reviewed with data access, deployment assumptions, and
mitigations. Existing public access MUST NOT broaden implicitly; local-network access
alone MUST NOT authorize new functionality.
External MQTT/HTTP/WebSocket/model payloads MUST be validated before changing domain
state. Secrets/tokens MUST NOT enter committed examples, logs, or browser bundles.
Household context/history MUST be limited to operational need. Model output MUST pass
deterministic validation and authorization before device-affecting execution.

### NH-04 — Keep boundaries and contracts explicit

Business rules MUST be testable independently of transports/UI. Changes MUST use established
controller/service/repository/frontend-state boundaries unless the plan justifies
replacement. New dependencies/abstractions MUST meet a specified need and document
operational cost. HTTP/WS/MQTT/MCP/configuration/persistence changes MUST identify
producers, consumers, and compatibility; breaking changes MUST specify coordinated
migration and recovery. Significant boundary/storage/dependency/trust decisions MUST
have an ADR with alternatives.

### NH-05 — Verify with reproducible evidence

Each acceptance requirement MUST map to a test or explicit manual procedure/result.
Changed behavior MUST cover meaningful failures/boundaries; fixes MUST have regression
evidence. Automated tests MUST isolate external dependencies; routine tests MUST NOT
command household devices. Existing backend/frontend suites remain merge checks.
Risk-appropriate build/lint/race/integration checks MUST be selected in the plan.
Reviews MUST distinguish passed, failed, and unrun checks; test success alone does
not establish constitutional compliance.

### NH-06 — Preserve integrity and recoverability

Metrics changes MUST define units, time boundaries/timezones, retention, and missing
data; missing/stale values MUST NOT silently become measured zero. Persistence changes
MUST specify compatibility and data-loss protection. Destructive operations MUST
define scope and appropriate recovery or informed confirmation.
Background work/integrations MUST bound resources, expose failures, and define shutdown.
Optional assistant services MUST NOT become dependencies of currently independent
device-management paths. New performance/reliability targets MUST be measurable in the spec.

## Development gates

Follow the [SDD workflow](../../docs/sdd/README.md). Plans MUST check every NH rule and
applicable component rule before implementation, after design, and before merge,
with evidence or justified N/A. Tasks MUST trace requirements and verification;
scope changes MUST update specs/plans. Debt affecting the change MUST be recorded;
unrelated debt requires no rewrite. Touched violations MUST be corrected or covered
by an approved scoped exception; documenting a requirement does not fix code.

## Governance

A maintainer adopts this proposal through review and records actual ratification.
Authors/agents MUST NOT invent approval. After adoption, reviewed amendments MUST
explain rationale, affected IDs/specs, and transitions, and update affected component
rules/templates/workflow/agent guidance in the same change. Record impact in the PR
or amendment record.

Versioning: MAJOR removes/incompatibly redefines rules or authority; MINOR adds or
materially expands obligations; PATCH clarifies without changing them. Preserve IDs
and original ratification date. Components version independently and declare parent version.

Exceptions MUST record rule, scope, rationale, risk, mitigation, named maintainer
approval, expiry/review date, and remediation task in the plan/ADR. Agents cannot
self-approve. Unapproved MUST violations block merge. Exceptions are temporary;
permanent changes require amendment. Contributors prepare evidence, reviewers assess
compliance, and maintainers approve adoption/amendments/exceptions through repository review.
