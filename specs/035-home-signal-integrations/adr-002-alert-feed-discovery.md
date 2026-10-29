# ADR-002: Bounded one-hop Atom link discovery for alert snapshots

Status: Proposed; blocks UK alert provider until approved | Date: 2026-10-29
Spec: [spec.md](spec.md) | Plan: [plan.md](plan.md) | Research: [provider research](provider-research.md)
Rules: NH-03/04/06, BE-01/03/05

## Context

The Met Office NSWWS warnings snapshot has no stable URL. Every update publishes a new GeoJSON URL
containing the feed's UUID. The documented flow is: fetch the Atom feed, follow the `rel="related"`
link, then fetch the GeoJSON snapshot. The snapshot also excludes expired and cancelled warnings.

The integration engine executes declarative JSON+JSON-Pointer requests. It has no XML parser and no
way to derive one request's URL from another response except typed `captures` stored as `state.*`.
A captured snapshot URL would be reused on the next refresh and return 404, so capture alone cannot
work. Two-origin `combine` is same-origin only and does not support a dependency between requests.

## Options

1. Hard-code a snapshot path or UUID. Rejected: stale by design and returns 404 after the next
   update.
2. Add a general XML/Atom parser to the runtime. Rejected: new parser surface, more failure modes,
   and only one link field is needed.
3. Implement Met Office warnings as a bespoke Go client outside the provider engine. Rejected:
   breaks the declarative provider model and host-editability.
4. Add a bounded discovery step to the request model: fetch a declared feed, extract exactly one
   same-origin link matching a declared `rel`, and use it immediately for the next request in the
   same refresh.
5. Skip UK alerts and ship only NWS. Rejected as the default because the household is UK-based,
   but retained as a fallback if this ADR is not approved.

## Decision

Adopt option 4, limited as follows:

- `Request` gains an optional `discover` block: `{format: atom, rel: related, same_origin: true}`.
- Discovery executes immediately before the request that declares it, within the same operation
  deadline. It follows **at most one** link and never recurses.
- The feed body is bounded (for example 256 KiB). Only `<link>` elements with the exact declared
  `rel` are considered; exactly one must match, otherwise the operation fails.
- The discovered URL must be `https`, port 443, and the **same origin** as the provider
  `base_url`. Redirects are refused. Private, loopback, link-local and metadata addresses remain
  blocked by the existing transport rules.
- The discovered URL replaces the declared request path/query for that one invocation only. It is
  not persisted as `state.*` and cannot be referenced by later refreshes.
- A parse failure, missing link, multiple matching links, foreign origin, oversize body or
  non-200 response fails the refresh, publishes nothing and preserves last-good data.
- The definition compiler validates the block and restricts it to a single request per operation.

## Consequences

- The engine gains a narrow, auditable transport capability rather than a scripting language.
- Attack surface increases by one parser path for a fixed subset of XML (`<link rel href>`), with
  same-origin enforcement on the result.
- Only providers with this declared shape can use it; ordinary JSON providers are unaffected.
- The Met Office provider still needs `apikey_header` (`apikey`) and its DataHub subscription.
- NWS remains an independent provider with a stable endpoint and does not use discovery.

## Validation and reconsideration

Tests: valid feed with one link; zero links; two matching links; different origin; missing `https`;
redirect; oversized body; malformed XML; 403/404/429; discovery timeout; last-good preservation;
existing non-discovery requests unchanged; race checks on the operation deadline. Reconsider if a
provider publishes a stable snapshot URL or if more than one link hop is ever required.
