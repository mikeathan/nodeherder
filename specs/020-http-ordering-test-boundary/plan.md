# Plan: HTTP ordering test boundary

Spec: [spec.md](spec.md) | Status: Complete | Updated: 2026-10-17
Constitutions: parent/backend v1.0.0 proposed; frontend N/A.

## Design and compatibility

In the existing hub-controller_test.go companion, count observed state broadcasts on
its single device lane and close a channel when all 1,000 arrive. Wait for that channel
before enqueueing the trailing probe. The existing buffered capture channel, burst,
awaitProbe and exact sequence assertions remain. No production/API/storage changes.
No new worker, timer ownership, retry or enlarged queue; use the existing two-second probe wait bound for the acknowledgement.

## Constitution Check

| Rule | Evidence/design | Before implementation | Final |
| --- | --- | --- | --- |
| NH-01 | CI reproduction and AC-01–03 recorded first | PASS | PASS |
| NH-02 | Device/runtime semantics unchanged | PASS | PASS |
| NH-03 | Only isolated mock harness | PASS | PASS |
| NH-04 | Existing controller companion owns regression | PASS | PASS |
| NH-05 | Preserve all assertions; CPU/race/full checks | PASS | PASS |
| NH-06 | No persistence/resource changes | PASS | PASS |
| BE-01 | No production boundaries changed | PASS | PASS |
| BE-02 | Every report/order/value retained | PASS | PASS |
| BE-03 | Channel synchronization; race detector | PASS | PASS |
| BE-04 | No storage changes | PASS | PASS |
| BE-05 | No routes/assistant changes | PASS | PASS |

## Verification and completion

Run single-/multi-CPU repeated HTTP ordering tests, controller/lane races, full Go
suite/vet/build. Preserve lane rejection regression coverage. Update PR evidence and
push the isolated test fix; no household commands. Results (2026-10-17):

- Before fix: `go test -count=1 -cpu=1 -run
  '^TestHTTPIngestionIsOrderedPerDevice$' ./internal/controllers` reproduced exactly:
  probe rejected at 1,000 queued tasks; all 1,000 state reports observed.
- After fix: `go test -count=100 -cpu=1,2,4 -run
  '^TestHTTPIngestionIsOrderedPerDevice$' ./internal/controllers`: PASS, 300 runs.
- `go test -race -count=1 -timeout=3m ./internal/controllers ./internal/lanes`: PASS,
  including unchanged queue overflow/rejection tests.
- `go test -count=1 -timeout=5m ./...`: PASS all packages.
- `go vet ./...`, `go build ./...`, `git diff --check`: PASS.
- Reviewed diff: exactly the same burst, capture buffer, final probe, count/order/value
  assertions and existing awaitProbe timeout. Added a bounded channel acknowledgement
  before the extra probe; no production edits or increased limits/timeouts.

No frontend source changed, so frontend suites were not repeated. The logged fixture
errors exercise intentional failures; the real CI assertion was the missing completion
probe. PR #60 receives this isolated follow-up commit and updated verification notes.
