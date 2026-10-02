# Plan: <name>

Spec: [spec.md](spec.md) | Status: Draft | Author: <name> | Updated: <YYYY-MM-DD>
Constitutions: parent <version>; components <versions/N/A>

## Context and design

Affected paths/current behavior; ownership/data flow; smallest suitable design;
dependencies/costs/alternatives; significant decisions linked as ADRs.
Distinguish observed facts from proposals.

## Constitution Check

Add each applicable BE/FE rule. Results: PASS with evidence; N/A with reason;
BLOCKED; approved exception with reference. Planned checks are not passed checks.
Before implementation, PASS may reference supported design evidence; execution
checks remain pending until run. Recheck after design changes; record actual reviews only.

| Rule | Design/verification evidence | Before implementation | Before merge |
| --- | --- | --- | --- |
| NH-01 | <spec/questions> | <result> | <result> |
| NH-02 | <device/event semantics> | <result> | <result> |
| NH-03 | <trust/exposure/data> | <result> | <result> |
| NH-04 | <boundaries/contracts/ADRs> | <result> | <result> |
| NH-05 | <traceability/checks> | <result> | <result> |
| NH-06 | <integrity/lifecycle/recovery> | <result> | <result> |

## Contracts, data, rollout

Changed HTTP/WS/MQTT/MCP schemas/examples/consumers; persistence/configuration;
compatibility/migration/deployment order/rollback/recovery. Add `contracts/` if useful;
justify N/A sections.

## Traceability and verification

| Requirement/acceptance IDs | Paths | Test/manual procedure | Actual result |
| --- | --- | --- | --- |
| <FR/NFR/AC> | <paths> | <names/steps> | <pending/pass/fail/unrun + reason> |

Select component checks; record environment/exact commands/results, mocked dependencies,
and separately authorized real-device targets. Cover failures/regressions/boundaries.

## Debt, risks, exceptions

Affected gaps/consequences/mitigations/follow-ups. Exceptions: rule/scope/rationale/risk/
mitigation/named maintainer approval/expiry or review date/remediation task.
Pending approval is BLOCKED.

## Completion

Acceptance evidence, changed docs, limitations, review links. Amendments: version
rationale and affected component/template/guidance updates. Remove unused prompts.
