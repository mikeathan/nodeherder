# Read-only sensor observations

Observed 2026-10-16, approximately 14:13–14:15 Europe/London, through the user-authorized
Zigbee2MQTT frontend. No device commands, reporting changes, logging changes, interviews,
reconfiguration, or writes were performed. Household names/addresses omitted here.

## Visible evidence

A Tuya ZY-M100-S_1 presence sensor (TS0601 / _TZE204_sxm7l9xa) showed:

| Local time | Field | Before → after |
| --- | --- | --- |
| 14:13:37 | linkquality | 134 → 131 |
| 14:13:38 | illuminance | 163 → 165 lx |
| 14:13:38 | linkquality | 131 → 134 |
| 14:13:40 | linkquality | 134 → 138 |
| 14:13:41 | linkquality | 138 → 134 |
| 14:13:46 | linkquality | 134 → 138 |
| 14:13:47 | linkquality | 138 → 134 → 131 → 134 (three entries) |
| 14:13:48 | illuminance | 165 → 168 lx |
| 14:13:49 | illuminance | 168 → 169 lx |
| 14:13:49 | linkquality | 134 → 131 |
| 14:15:00 | illuminance | 183 → 185 lx |
| 14:15:00 | linkquality | 131 → 134 |
| 14:15:03–08 | linkquality | 134 ↔ 138 (five entries) |
| 14:15:08 | illuminance | 185 → 187 lx |

The initial dashboard showed ten activity entries across twelve seconds, all from this
sensor. At least three distinct activity entries shared one displayed second. This
establishes bursts of changed values; the UI's whole-second timestamps cannot establish
exact packet spacing or total MQTT report rate. About-page diagnostics displayed an
aggregate 0.7618 messages/s, not a rate measured over this observation window.

Other visible changes: air sensor temperature 20.2 → 20.1 °C and CO2 351 → 352; a plug
power 122 → 128 W; later another power transition 121 → 111 W and current 0.64 → 0.39 A.
These demonstrate why one absolute tolerance for all numeric fields is inappropriate.

Presence remained false and target distance was 0 m. No distance-jitter trace was
observed. Both minimum_range and maximum_range displayed 6 m. This is a configuration
observation, not proof of sensor malfunction or authorization to change its range.
Generic per-device debounce/throttle fields appeared unset; no filtered attribute names
were visible in the accessibility tree. This does not prove that no upstream/global
filter exists. Logs were configured at warning; the level was left untouched.

## Design implications

- Suppress duplicates and small oscillations relative to the last published value,
  preserving cumulative drift (163 → 187 lx) and real step changes.
- Keep binary presence events and writable confirmations immediate.
- Use separate observation policies per field/units; do not throttle a whole MQTT message.
- Preserve raw state for automation reads. Distinguish filtered observations explicitly.
- Choose a distance tolerance as an initial configurable hypothesis, not a measured
  characteristic of this sensor: the live sample did not contain changing distances.
