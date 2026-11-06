# ADR-001: Separate spatial history store and movement-gated recording

Status: accepted (unreviewed draft) | Date: 2026-11-06 | Spec: 056

## Context

Positions are high-volume and sensitive compared with ordinary device metrics. The existing
metrics store records one row per expose, which cannot represent an atomic X/Y pair, loses path
semantics under aggregation, and shares a 10-day retention with all device history. Storing raw
frames would be both too large and too revealing.

## Decision

- Record into a dedicated `spatial` bbolt database in the existing data directory. It has its own
  bucket, its own 7-day purge and can be inspected separately by the storage card.
- Gate recording on the existing per-device **Record history** switch; no second user concept.
- Record at most one point per visible target per second, only after 300 mm of movement; no points
  while still; 50 mm quantisation; duplicate slot reports within 350 mm merged; 5-minute/300-point
  chunks; 3,000 points/day cap.
- Store only positions, timestamps and slot numbers. No identity, video, raw frames or diagnostic
  entities. The API is authenticated and not exposed publicly or through MCP.
- Retention is a code constant for v1. If configuration is ever required, it belongs on the
  device history setting, not a root `hub.spatial` block.

## Consequences

Positive: bounded storage, simple delete/retention, no fake exposes, path/heatmap semantics,
source-of-truth enablement.

Negative: positions are plaintext in the data volume and must be included in backup/privacy
decisions; bbolt does not shrink after purge; there is a second file; session closure relies on a
firmware clear frame because frame silence is normal for a still target.
