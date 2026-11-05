# Investigation: a backend spatial recorder

Status: investigation only; no decision taken and no implementation. Branch:
`plan/backend-spatial-recorder`.

## The request

Move replay/recording from the browser into the backend, on by default, without collecting
high-resolution data:

- ignore micro-movement of a stationary target;
- record only a sensible amount, bounded by area/time;
- do not pollute the metrics database or grow it without limit;
- reuse the existing metrics database/purge if that is the right trade-off;
- come back with findings, options, pros/cons and usability.

## Findings in the current system

### Frames available on the hub

- The device lifetime builds synchronized spatial frames after every accepted update and emits
  them through `internal/services/spatial_emitter.go`. Emission is capped at **100 ms per
  device** (`spatialFrameInterval`, line 19) and deduplicated.
- Frames contain up to 3 firmware target slots with X/Y in millimetres plus an active flag. They
  arrive **change-driven**: a still target produces few or no new frames, which is good for a
  recorder (no need to filter an idle stream) but means "stillness" can also look like silence.
- Frames are already the accepted, consistent view used by automation and the browser; a recorder
  hooked there cannot see torn per-axis updates.

### Presence signals already persisted

- The derived presence device has `occupied` and `target_count` plus per-zone
  `zone_*_occupied`/`zone_*_count` as **measurement** entities
  (`internal/spatial/entities.go:105-114`); `last_zone`, `empty_seconds` and dwell are diagnostic.
- Measurement storage is change-driven: `DeviceLifetimeService.Update` filters through
  `acceptExposeReport` and per-expose thresholds before calling `attemptToStoreMetrics`
  (`internal/services/device_lifetime.go:222-243,303-321`). Counter reports every 10 s therefore
  do **not** write unchanged values; only real changes do.
- So zone-level presence history already exists. A position recorder adds a different capability
  (paths/heatmap), it does not replace zone analytics.

### Metrics database

- bbolt file `<data dir>/metrics.db`, root bucket `metrics`, child bucket per device id
  (`internal/metrics/storage/metrics.go:15,30`; `utils/storage/key-value.go:40`).
- Keys are `UTC timestamp (30 bytes, nanosecond format)_expose`; values are gob-encoded
  (`internal/metrics/domain/key_generator.go:25-35`, `utils.AnyToByteArray`).
- Retention: default **10 days**, cleanup every **12 hours**, both configurable
  (`models/settings/config.go:115-118`, `store/tasks.go:46-74`).
- `Prune` deletes expired keys in one write transaction (`utils/storage/key-value.go:232`);
  **bbolt does not shrink its file** after deletion. The file is a high-water mark and needs
  compaction/rotation if a large retention is ever changed downwards.
- Query: `/metrics/query` (through the history service) requires the expose to exist on the
  device, returns at most **10,001 points**, scans at most 250,000 rows, and allows 366-day
  ranges (`internal/metrics/services/query_service.go:40-43,72-120`). Aggregation/binning is
  per expose, not a spatial query.

### Measured row cost

Measured on this machine with 200,000 rows in a fresh bbolt file, including keys, values and
page overhead:

| Layout | File size | Per row |
| --- | --- | --- |
| Metrics style (timestamp+expose key, gob float value) | 42.8 MB | **214 B** |
| Dedicated binary (timestamp+device+slot key, 16-byte X/Y value) | 36.7 MB | **183.5 B** |

bbolt's page overhead dominates; packing is what makes the difference, not the encoding alone.

## Data-volume arithmetic

One "point" = one slot position. Two metrics rows per point if X/Y are separate exposes.

| Strategy | Points/day/device | Metrics-style (10 d) | Binary row store (7 d) |
| --- | --- | --- | --- |
| Raw 4 Hz, 3 slots, 24 h | 1,036,800 | ~4.4 GB (10 d) | ~1.3 GB (7 d) |
| 1 Hz while present, 3 slots, 8 h | 86,400 | ~370 MB (10 d) | ~111 MB (7 d) |
| 1 Hz, motion-gated + 60 s heartbeat, 3 slots, 24 h presence | ~4,300–5,000 | ~18–21 MB (10 d) | ~6 MB (7 d) |
| Chunked/simplified sessions (recommended shape) | same points | n/a | **tens of KB–~0.2 MB/day** |

The raw 4 Hz row is why a recorder must not persist frames directly. Even 1 Hz continuous is
larger than most households need; motion gating plus a heartbeat is the useful middle ground.

## Options

### A. Reuse per-expose metrics with synthetic coordinates

Write `slot1_x`, `slot1_y`, ... as measurement metrics through the existing metrics repository,
optionally on a dedicated "history" device so the real device entity list stays clean.

Pros:

- Reuses the existing purge, retention setting, backup and query path.
- No new storage engine, migration or cleanup task.
- Device history UI can chart X and Y immediately.

Cons:

- **Two rows per point** (~428 B/point): separate keys, not atomic pairs; replay must re-pair by
  timestamps.
- Needs synthetic exposes to satisfy query admission and to appear in `device.Exposes`; adds a
  fake device/entities to device lists, automations and search unless special-cased.
- No session, bounding box or path query; 10k-point limit returns a very short slice of a day at
  1 Hz.
- Retention is global to metrics (default 10 days); per-device "metrics disabled" or a metric
  override can silently stop the recorder.
- Binning/aggregation per series destroys paths; heatmap becomes a client-side join of two
  independently binned series.
- bbolt file still does not shrink.

Usability: fine for two line charts; poor for replay and heatmaps. **Not recommended for raw
points.**

### B. Dedicated compact recorder store, one row per sample

New repository/service writing `{timestamp, device, slot, x, y}` rows as binary values, with its
own device-scoped bucket, retention and purge. Either a separate `spatial.db` or a dedicated child
bucket in `metrics.db` with timestamp-compatible keys.

Pros:

- Atomic X/Y, one row per point (~183.5 B measured, less with delta encoding).
- Independent retention, per-device caps and disable; no fake exposes or device-list pollution.
- Purpose-built query for a time range/session, not a metrics aggregation.
- Can share the existing bbolt file/purge **only if** keys use the same timestamp prefix and the
  retention is intentionally shared; otherwise a small separate purge worker.

Cons:

- New repository, cleanup task, query endpoint and tests.
- If separate DB: another file to back up/monitor; if shared: retention coupling and write
  contention on bbolt's single writer.
- Still a high-water file; high sample counts cannot be solved by retention alone.

Usability: good replay/heatmap foundation, but row-oriented storage still grows with motion.

### C. Session/chunked polyline recorder (recommended direction)

Auto-create a **session** when targets appear, keep at most one sample per slot per second, ignore
movement below a deadband, write fixed-size chunks (e.g. 1–5 minutes) of delta-encoded points plus
session metadata.

Pros:

- Stores **tens of KB to ~0.2 MB per active day/device**; effectively immune to DB pollution.
- Natural "recordings" list: start/end, duration, slots seen, point count, capped flag.
- Few large bbolt writes instead of thousands of tiny ones; purge per chunk/session is trivial.
- Stationary periods become a heartbeat in the session, not a jitter cloud.
- Privacy: no fine-grained raw frame history by design.

Cons:

- Encoder/decoder (delta varint + optional simplification) and chunk lifecycle are new code.
- A chunk must be rewritten as points arrive (small, bounded); or accept a short write delay and
  keep the live 5-minute tail in the browser until the chunk closes.
- Query returns coarse paths, not every frame; heatmap resolution depends on sampling.
- Sessionization has to define gaps, slot churn and multi-target sessions.

Usability: the best match for "recordings I can list and replay", and the lowest storage risk.

### D. Aggregate-only grid/heatmap

Store counts per coarse cell (e.g. 250 mm) per minute, no paths.

Pros: smallest data, simplest query, privacy-friendly; persists the heatmap.
Cons: no replay/path, cannot inspect a moment, limited tuning value. Complementary to C, not a
replacement for a recorder.

### E. Keep the current session-only browser buffer

Pros: zero storage, already shipped, privacy by construction.
Cons: lost on reload/tab switch, no list, no cross-device history — the reason for this request.

## Recommendation: the minimum-points technique

**Option C — movement-only session chunks.** Keep the smallest representation that still
replays a useful path:

- Hook accepted frames in the controller; one worker per source.
- Record at most **1 point per visible target per second**.
- On that tick, emit a point only if the target moved **≥300 mm** since its last point.
- **No heartbeat points while still**: a stationary target adds nothing, so micro-movement never
  becomes a cloud of dots. Presence and "still here" already come from the derived occupancy
  entities; the replay can hold the last point for the session.
- Close a session when derived occupancy clears, the source goes offline, or a **30-minute**
  no-frame safety limit is reached.
- Quantize to **50 mm** and delta-encode points into **5-minute chunks**.
- Default **7-day** retention, **3,000 points/day/device** cap, global size guard.

Why this is the best resolution-per-byte compromise: at walking speed 1 sample/s is smooth
enough to see the route and heatmap; at rest it records zero points. One person walking for one
hour produces at most ~3,600 points; a typical day is ~1,000–4,000; two people roughly double.
Chunked/delta storage puts that at roughly **30–150 KB/day**, versus ~0.7 MB/day as binary rows
and ~1.5 MB/day as metrics pairs. Raw 4 Hz would be >100× that and is excluded.

### Slots answer

The LD2450 exposes **three physical target slots**, not three people. One visible target
normally occupies one active slot → one point per sample; two visible targets occupy two slots →
two points; a ghost/reflection or duplicate can briefly occupy a third. Slots are reused over
time, so the recorder should use the same **duplicate/merge rules as the display** (≤3 visible
tracks, 350 mm merge) and must not treat slot numbers as identities. The daily cap therefore
applies per device, not “3 × slots”.

### Storage and purge answer

- **Not per-expose metrics rows.** A dedicated `spatial` bucket with timestamp-prefixed keys and
  chunked values; one row per 5-minute chunk, not per point.
- Put it in **`metrics.db` but a separate bucket**, so there is still one file to back up, while
  the spatial data can have its own purge.
- **Use a different, shorter timeout:** positions are bulkier and more sensitive than device
  metrics. Proposed **7 days** versus the metrics default of 10; a daily spatial-only purge that
  removes whole chunks, oldest first. Do **not** hitch it to the global metrics prune, or the
  retention becomes 10 days (and any user change to metrics retention silently changes recorder
  retention).
- If isolation matters more than one file, a separate `spatial.db` with the same purge design is
  the cleaner alternative.
- bbolt does not shrink after deletes; monitor file size (chunked storage makes this small).

### UI answer: replace the browser recorder with a sessions player and a heatmap

The current browser buffer and the in-Presence **Diagnostics** view are **removed**; diagnostics
are explicitly out of scope now. Keeping two recorders or a diagnostics pipeline that depends on
them would duplicate logic and leave history trapped in a tab. The server becomes the only source
of truth.

Recording and playback move out of the Presence tab into a new top-level **Recordings** section
(sidebar entry), with a device picker and a date/time range shared by two tabs:

1. **Player** — the range filters a list of auto-created sessions for the chosen device (start,
   end, duration, visible targets, point count, capped badge). Selecting one loads its points and
   plays them with a scrubber, play/pause and speed over the normal room/zones. This is the
   "data-range player": choose the range, pick the session within it, play the recorded
   coordinates.
2. **Heatmap** — the same date range aggregates recorded points into a coarse grid (e.g. 250 mm
   cells) and shows density with zone outlines. No playback in this view.

Playing every session in a range concatenated would be ambiguous across gaps; the range selects
which sessions are offered, and playback is per session. The heatmap is the across-range view.

Presence keeps Live, Zones, Room, Entry/Exit, Firmware, Calibration and Multi-sensor. Diagnostics,
the session-only heatmap and the browser Replay view are deleted rather than left as duplicates.

The server can keep the **active session** in memory and return it as a live tail (flushed chunks
plus the current minutes), so the player can open the session in progress without a browser-side
buffer.

Data endpoints (authenticated, no public/MCP exposure by default):

- `GET /api/spatial/recordings?device&from&to` → session summaries for the date range.
- `GET /api/spatial/recordings/{session}` → points for the player (downsampled if huge).
- `GET /api/spatial/heatmap?device&from&to&cell=250` → aggregated cells for the heatmap.

Heatmap is derived on demand from the stored chunks; there is no second copy of the data. A
seven-day, 4,000-point/day device is only ~28,000 points to aggregate server-side.

Removal scope: the Diagnostics view, its domain helpers/tests, `useSpatialReplay`, the browser
frame recorder, and the Presence Replay/Heatmap tabs. The calibration and firmware views do not
depend on them. This is a deletion, not a hidden second path: no Diagnostics and no browser
recording remain.

## What gets captured when "Record history" is on

For a mapped ESPHome/EPL source, the intended capture is:

1. **Existing metrics** — measurement-category exposes only, filtered by the same per-expose
   policy as today. The target coordinate entities are diagnostic, so they are not written as
   ordinary metric rows.
2. **Spatial sessions** — movement-only points from accepted frames: ≤1 point/s per visible track,
   only after 300 mm movement, 50 mm quantized, grouped into 5-minute chunks. No points while
   still; no raw frames.
3. **Derived presence history** — `occupied`, `target_count`, zone occupied/count change events.
   The derived device **inherits the source device's history switch** rather than needing its own
   opt-in. `last_zone`, `empty_seconds` and dwell stay diagnostic and unrecorded.
4. **Nothing else** — no video, no identity, no still jitter cloud, no diagnostic entities.

When "Record history" is off, neither metrics nor spatial sessions are captured. Retention remains
separate: spatial chunks 7 days, metrics 10 days.

## Defaults worth agreeing

| Setting | Proposed default | Why |
| --- | --- | --- |
| Enabled | **Follows the device's existing "Record history" toggle** (no separate recorder switch, no default-on) | One simple, consistent switch per device |
| Sampling | ≤1 Hz per visible target | Enough path resolution; far below 10 Hz frames |
| Movement gate | 300 mm | Ignores stationary jitter and sway |
| Still behaviour | No points; replay holds last position | Minimum daily points |
| Session close | occupancy clear, source offline, or 30 min without frames | Avoids frame-silence ambiguity |
| Quantization | 50 mm | Noise reduction and smaller deltas |
| Retention | 7 days, spatial-only purge | Shorter and independent of metrics |
| Daily cap | 3,000 points/device | Bounds worst-case walking |
| Global cap | ~100 MB, oldest chunks first | Protect disk |
| Privacy | authenticated API only; no public/MCP by default | Positions reveal occupancy patterns |

## Usability questions this raises

- Auto-sessions vs explicit start/stop? Auto matches "default-on" and needs no user action; an
  explicit mode could be added later for "record this walk".
- Session list should show time, duration, slots and a "capped/coarse" badge, not imply person
  identity. Zones remain room-level overlays.
- Where do recordings live? Backend, so they survive reloads and tabs; the UI must say retention
  and offer delete/disable.
- Replay at 1 Hz may look steppy; the browser can interpolate visually without storing more.
- Multi-sensor: one recorder per source, no cross-sensor identity, same privacy rules.

## Open decisions before a spec

1. Confirm **Option C (session/chunked)** and the replacement scope: remove the browser
   recorder **and Diagnostics**, move recording into a new top-level Recordings section, add the
   per-session data-range player and the date-range persisted heatmap.
2. Confirm the defaults table (sample rate, 300 mm gate, heartbeat, 7 days, caps).
3. Confirm the source device's existing "Record history" toggle is the only switch (and the
   derived presence device inherits it).
4. Shared `metrics.db` bucket vs separate `spatial.db` (storage/purge ownership and backups).
5. Should the browser keep its 5-minute live tail, or is a server round-trip acceptable for all
   replay?
6. Any assistant/MCP exposure? Recommendation: none by default.

## Next step

Pick the option and defaults, then write the spec/plan/tasks and implement behind a per-device
flag first, measuring the real row/chunk counts and bbolt growth before turning it on by default.
