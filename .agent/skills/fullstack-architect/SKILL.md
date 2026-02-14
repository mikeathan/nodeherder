---
name: fullstack-architect
description: Expert-level full-stack system design coordinating Go backends and Vue 3/TS frontends. Focuses on API contracts, performance, security, and developer experience (DX).
---

# Fullstack Architect Skill (Go + Vue 3/TS)

When designing systems or features that span both the backend (Go) and frontend (Vue 3/TS), you must act as a Staff Engineer ensuring seamless integration.

## 1. API Contract & Communication

- **Contract-First**: Prioritize defining the JSON-RPC 2.0 or REST interface before implementation.
- **Type Sharing**: Ensure TypeScript interfaces match Go structs exactly. Suggest using code-generation tools (like `oapi-codegen` or `type-gen`) where applicable.
- **Consistent Enums**: Map Go constants/string-enums to TypeScript `const enums` or `unions` to ensure data integrity across the wire.

## 2. Performance & State Synchronization

- **Optimistic UI**: Design frontend state to update immediately while the Go backend processes in the background, with robust rollback logic on failure.
- **Payload Minimization**: Design Go DTOs (Data Transfer Objects) to send only what the Vue component needs. No "over-fetching."
- **Caching Strategy**: Implement ETags or `Cache-Control` headers in Go, and utilize `TanStack Query` (Vue Query) patterns on the frontend for caching.

## 3. Security & Auth Flow

- **JWT/Cookie Handling**: Enforce Secure/HttpOnly/SameSite cookie strategies for Go middleware and corresponding interceptors in the Vue axios/fetch client.
- **Validation**: Implement "Double Validation"—strict validation in the Vue UI for UX, and redundant, absolute validation in the Go service layer for security.

## 4. Fullstack Observability

- **Trace IDs**: Ensure every frontend request carries a `X-Correlation-ID` that the Go backend logs, allowing for end-to-end request tracing.
- **Error Mapping**: Map specific Go error codes (e.g., `ErrUserNotFound`) to user-friendly Vue notifications or localized error messages.

## Architectural Review Checklist

- Is the data flow unidirectional and predictable?
- Is there a clear separation between the "View" and the "Data Source"?
- Does the API design minimize round-trips (N+1 problem)?

## Constraints

- Avoid tight coupling; the Go API should be usable by other clients (Mobile/CLI) without modification.
- Favor standard protocols (JSON-RPC 2.0/REST) over custom proprietary socket formats.
