# ADR 003: Optional ESPHome mDNS discovery

Status: Proposed | 2026-10-31

ESPHome advertises `_esphomelib._tcp` over mDNS. Manual numeric-IP enrollment remains
available when multicast cannot cross a Docker bridge or subnet. The UI initiates a
bounded scan only on request; results are untrusted candidates, not enrolled devices.

Use `github.com/miekg/dns` v1.1.55 (BSD-3-Clause) to parse DNS messages, with one explicitly
owned UDP multicast socket and a bounded query. The initially evaluated HashiCorp mDNS
v1.0.6 can leave sockets open when multicast setup partially fails, so it is not used;
v1.0.7 also requires Go 1.25 while this project targets Go 1.24.8. Pin the parser and
re-evaluate vulnerabilities and dependencies with toolchain updates. Alternatives considered:
manual-only onboarding (does not meet discovery requirement), handwritten DNS parsing
(unnecessary security maintenance) and runtime plugins (unjustified complexity).
Do not perform multicast in automated tests; use fake DNS records.

Discovery lasts at most three seconds, caps candidates at 64 and accepts only validated
private numeric IP addresses and ports. It never resolves a returned hostname, connects,
enrolls, logs raw TXT metadata or trusts a candidate as identity. The authenticated POST
endpoint has the same cookie-origin protections as enrollment and limits concurrent scans.
An empty result is not proof that no devices exist. Document Docker multicast limitations.
