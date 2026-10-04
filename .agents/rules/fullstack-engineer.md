# Cross-stack rules

Scope: cross-component design, implementation, review. Authority:
[project constitution](../../.specify/memory/constitution.md) and both component constitutions.

- Define contracts before implementation: wire fields, enum values, optionality,
  errors, authentication, producers, consumers. Match serialized Go/TypeScript meanings,
  including missing/null/zero distinctions. Code generation is optional when maintenance cost is justified.
- Trace changed WS events through `backend/internal/ws/eventhub.go`, frontend
  `src/store/modules/ws/index.ts`, affected types/contracts, and mocks.
- Preserve HTTP/WS/MQTT/MCP protocols and state ownership. Replacement protocols or
  cache/state libraries need specified value and design review.
- Minimize unnecessary payloads/round-trips. Evaluate caching headers, invalidation,
  and sensitive data together; cache only for demonstrated need.
- Validate frontend input for feedback and backend input at trust boundaries.
  Retain auth helpers; evaluate cookie flags where cookies apply.
- NH-02: pending/optimistic device presentation stays distinct from confirmed state;
  define reconciliation and failure feedback.
- Map backend errors to useful UI outcomes without leaking internals/secrets.
  Add correlation IDs for an observability need.

Review data flow, consumer compatibility, shared-state consistency, failure/recovery,
and independently testable boundaries. Introduce infrastructure only for specified needs.
