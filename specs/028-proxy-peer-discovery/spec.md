# Spec: Reliable proxy peer discovery for local sign-in

ID: 028 | Status: Implemented | Owner: NodeHerder contributors | Updated: 2026-10-22
Constitution: proposed 1.0.0 | Components: deployment tooling

## Problem, scope, exclusions

Spec 026 made `scripts/manage-stack.sh` inject `TRUSTED_PROXY_CIDRS` from the Compose
network gateway. It derived the network name as `<checkout-directory>_nodeherder`
lowercased by hand, but Compose normalizes the project name itself (`my.Checkout` becomes
`mycheckout`). On such a checkout the lookup misses, the script prints a warning and
starts without trust, so local sign-in silently falls back to Google. Operators also
asked where the resolved value lives, because it never appears in the env files.

Discover the peers from Compose's own project metadata and make the value's location
explicit. Do not widen trust, change auth rules, or add host configuration.

## Scenarios and acceptance

- AC-01: Given a checkout whose directory name Compose normalizes (uppercase, dots),
  when the stack starts, then the project's real network gateway is trusted.
- AC-02: Given several networks owned by the project, then every project network
  gateway is trusted once, with no duplicates and no whole-subnet prefixes.
- AC-03: Given the project has no resolvable gateway, then the stack still starts with
  a warning naming `TRUSTED_PROXY_CIDRS`, and no override file is written.
- AC-04: The applied value is discoverable: the start output names the generated file,
  and the docs show the `docker inspect` command for the container environment.

## Requirements and success

- FR-01: Read the project name from `docker compose config` and list networks by
  `com.docker.compose.project` label; acceptance: AC-01, AC-02.
- FR-02: Keep the existing fail-soft warning and single-host prefixes; acceptance:
  AC-03.
- FR-03: Report the generated override path and document where the value lives;
  acceptance: AC-04.
- SC-01: Script tests cover the normalized-name path, duplicate peers, and the
  no-gateway path; a real container run shows the variable in `Config.Env`.

## Entities and external interfaces

`TRUSTED_PROXY_CIDRS` keeps its meaning for `utils.IsLocalRequest`. The generated
Compose override in `.generated/` and the container environment are the only places the
value appears; env files stay untouched.

## Questions and assumptions

No blocking questions. Assumption: deployments use the repository Compose stack through
the stack manager, as documented.
