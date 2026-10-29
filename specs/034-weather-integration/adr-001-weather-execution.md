# ADR 001: Weather through shared provider definitions and operations

Status: Accepted design for implementation | Date: 2026-10-28

Weather uses the existing Connections catalogue, encrypted repository, capability registry, setup ownership and revision-fenced publication. The user requires the bin integration's host-editable YAML pattern and manual refresh.

Add `location_select` for synchronous coordinate choice and forecast retrieval. The named search GET may declare a separate public base URL. Cross-origin search receives no credentials; the UI discloses the recipient before submission. Choices contain validated coordinates and visible names/coordinates; the server only admits a current returned choice. A same-origin search can use the selected provider key. This permits Open-Meteo/GeoNames search for Open-Meteo and UK Met Office, and native WeatherAPI search. Met Office is initially restricted to GB locations with Europe/London timezone; expanding coverage requires timezone data in the selection contract.

Add fixed `apikey_header` and `apikey_query` modes. Credentials stay in existing private slots and are bound to the forecast origin. The existing HTTP transport refuses redirects, rejects private DNS answers, bounds request/response size and sanitizes transport failures.

A forecast GET may declare `combine` with one other same-origin, non-conditional GET. No child combination or search combination is permitted. Calls are sequential inside the existing 30-second operation context; child failures propagate to the existing failure/cooldown lifecycle. A validated combined result is published in one existing transaction after owner/revision checks. No partial forecast replaces last-good data.

Capability-owned declarative selectors cover parallel columns, row arrays and nested hourly days, with fixed unit scales and explicit condition maps. Normalization performs no network or storage calls. Unknown fields remain unavailable. Freshness is projected after one hour; no scheduler is introduced.

Alternatives rejected: separate weather runtime/storage (duplicates ownership), hidden HTTP calls inside the adapter (breaks capability contract), arbitrary executable mappings (unnecessary trust and maintenance cost), unrestricted provider URLs/key headers (unneeded scope). Existing waste definitions and records remain unchanged. Older binaries reject unknown weather Home placements; back up the Home layout before rollback and restore its pre-weather version rather than discarding references silently.

Home source discovery uses the optional `SourceProvider` capability interface and opaque DTO references. Home does not decode integration resources. Frontend capability presenters own setup/source/widget/icon/result presentation behind a composition registry. Existing persisted reference identities are retained. Transaction setup code is split by responsibility without moving writes or revision checks outside the original transaction.
