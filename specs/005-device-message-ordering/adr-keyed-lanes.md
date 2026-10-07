# ADR-005-01: Per-device ordered lanes for message processing

Status: proposed | Date: 2026-10-07
Spec/plan: [spec.md](spec.md), [plan.md](plan.md) | Owner/review: pending (no reviewer assigned) | Rules: NH-02, NH-04, NH-06, BE-02, BE-03

## Context and options

Zigbee2MQTT 2.14.2 publishes a stale-then-new pair per command; the backend handles messages
concurrently and reverses them (F-01–F-03). Dial automations also read live device state during
evaluation, so concurrent messages for one device corrupt each other's inputs (F-04). Requirements:
per-device order, no cross-device blocking, bounded resources, deterministic tests, no behaviour change
inside `Update` or the automations.

| Option | Ordering | Isolation | Resources | Complexity | Notes |
| --- | --- | --- | --- | --- | --- |
| A. Ordered MQTT only, keep pool | broken by pool | n/a | existing | low | does not fix |
| B. One global worker | total | none | 1 goroutine | lowest | one slow store/automation delays all devices |
| C. N hash-striped workers | per device | colliding devices block | N goroutines | low | fixed pool, head-of-line within a stripe |
| D. Lazy per-key lanes | per device | full | goroutine only while busy | low–medium | cap + recover + shutdown to write |
| E. Sequence numbers in `Update` | state only | n/a | none | medium | automations still see mixed values |

## Decision and consequences

Choose **D**, injected from the composition root behind small interfaces, with an ordered `ingress`
lane in front. Costs: one new package (~150 lines) plus tests; per-lane cap and shutdown to maintain.
Verification: AC-01–AC-04, AC-14, NFR-01–NFR-04 in the plan. Reconsider if the burst benchmark
shows lane overhead beyond NFR-01 (fallback: option B with the same interface, no caller changes
because only `lanes.Executor` changes).

## Supersession

None.
