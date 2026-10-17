# ADR 001: Generated defaults within the frontend build boundary

Status: Proposed implementation decision | 2026-10-16

Go embeds compatibility defaults from its package; the frontend image is built from
frontend/ only. Cross-directory imports cannot satisfy both constraints. Keep the
backend JSON authoritative and check in a byte-identical generated frontend artifact.
Both browser code and the mock server consume that artifact. A root CI check rejects
drift; generation is explicit and documented. No build hook reaches outside frontend/.

Rejected alternatives: widening Docker/Compose/CI contexts creates unrelated deployment
changes and potentially includes more host files; independently maintained copies
permit drift; relying only on server defaults would change offline/legacy fallback
behavior. Cost is one small generated JSON and a deterministic CI guard, with no
runtime network dependency or new packages.
