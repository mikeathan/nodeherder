# Spec: HTTP ordering regression at queue capacity

ID: 020 | Status: Complete | Updated: 2026-10-17
Constitutions: parent/backend v1.0.0 proposed.

PR #60 CI run 39030677507 failed TestHTTPIngestionIsOrderedPerDevice: the test sent
1,000 state reports plus one completion probe to a queue with capacity 1,000. The
probe was rejected under a valid scheduler interleaving. Production explicitly uses
bounded, nonblocking queues and logs/counts overload rejection (spec 005).

- AC-01: Keep all 1,000 HTTP reports and exact broadcast count/order/value assertions.
- AC-02: Submit the probe only after the burst has been observed, using a channel
  acknowledgement, not sleeps, retries, larger production limits or longer timeouts.
- AC-03: Pass repeated single-/multi-CPU runs and affected race/full backend checks.
  Keep explicit overload-rejection tests intact. No production behavior changes.
