# Historical preview verification

These results belong to the earlier synthetic prototypes and research. They do not certify
production code. See [plan.md](plan.md) for current implementation and verification.

## Actual verification and completion

2026-10-26, installed Node v26 and frontend Playwright/Chromium, local `file://` preview only.

- JavaScript compilation via `new Function` passed; final browser execution reported zero page errors.
- `node /private/tmp/nh-home-design-qa.cjs`: **274 checks passed**. Matrix: 216 design/mode/theme/
  data-state/viewport cases; 18 opt-in future-example cases; 27 pin/detail/room/status/exit flows;
  12 simulated phone/tablet-width cases; one dialog keyboard-wrap case. Viewports: 390, 1024 and
  1280px wide, 900px high. Every matrix case checks page/app horizontal overflow.
- Pin checks include save, cancel and Escape, plus Home hidden/panel visible and Home visible/panel
  hidden. Enter opens the synthetic detail page; Refresh sends no request; Back restores the view.
  Room arrows change the sample climate; status/exit remain available. The native dialog gained
  explicit Tab/Shift+Tab wrap after a keyboard check found focus could leave its controls.
- Network observation: zero HTTP(S) requests. Final localStorage and sessionStorage lengths: zero.
  No app server, MQTT broker, integration provider or household device was contacted.
- 19 screenshots generated under `/private/tmp/nh-home-design-qa/`. Inspected all three desktop
  Home designs, all three panel directions, phone Home/panel and light phone styling. Real touch,
  assistive technology, production routing/history and text-zoom verification remain unrun.
- Whitespace checks against each new artifact and all 24 local Markdown links passed.
- Initial Chromium launch was prevented by the execution sandbox; the same isolated file-only QA
  was permitted outside it. An ambiguous harness editor selector was corrected before the final
  pass; the subsequent keyboard-wrap failure was fixed in the prototype and the full matrix rerun.
- Frontend unit/lint/build and backend suite/build were **not run**: no application source changed.

### Earlier B draft-editor revision (superseded interaction)

`node /private/tmp/nh-household-order-qa.cjs`: **258 checks passed** for the revised prototype,
using the same isolated Chromium setup. This includes 216 design/mode/theme/data-state/width
cases and targeted B flows at 390/1024/1280px: all five groups moved using keyboard Enter on
named movement buttons; first/last boundaries; repeated movement; save/cancel/Escape; independent
draft pin cancellation/commit; saved panel ordering; Home/End/wrapped arrows; hidden Household
filtering; detail/back; dialog focus and nested Escape. Native pointer drag moved Kitchen ahead of
Living room, then verified the saved panel sequence and arrow navigation in that order.

No horizontal page/app overflow, browser page errors, HTTP requests or browser persistence.
Inspected the updated normal desktop layout plus desktop and phone arranging screenshots under
`/private/tmp/nh-household-order-qa/`. Application suites/build/lint remain unrun because production
source is unchanged. Future-example scenarios have the earlier revision's evidence only; actual
touch/assistive technology and production history/persistence remain unverified.

### Current screenshot-aligned revision

The supplied screenshot superseded the prior sample's outer Save/Cancel model and separate
control row. Current behaviour matches existing save-as-you-go editing with New area / Export /
Import / Done; Done/Escape retain changes. Pin/name/delete dialogs retain their own cancellation.
AreaCard.vue and icons.ts were inspected directly. Inline drag/move/rename/delete SVG paths now
match the app's mdiDragVertical, mdiArrowUp/Down, mdiPencilOutline and mdiTrashCanOutline exactly.

`node /private/tmp/nh-inline-edit-qa.cjs`: **273 checks passed**, including 216 design/mode/theme/
data-state/viewport cases. At 390/1024/1280px, checked same-row handle/title/four ghost controls
for all five groups, original toolbar text/order, red tile-remove badges, Add tiles, movement and
panel sequence, rename/duplicate-name validation, group-delete cancel/confirm, remove/restore tile,
independent integration pins, New area, Done and nested Escape. Additional desktop cases covered
native pointer drag, untrusted rename text remaining plain text, placement-only Household deletion,
preview export/import roundtrip and invalid-import rollback. The initial harness expected incorrect
leading spaces in toolbar text; corrected the assertion and reran successfully. Reran all 273
checks after substituting exact app icon paths; the current result is that final run.

No page/app horizontal overflow, browser errors, HTTP requests or local/session storage entries.
Inspected normal screenshot-aligned desktop edit mode and the edited phone layout; screenshots
are in `/private/tmp/nh-inline-edit-qa/`. Preview export/import stores only example group references
and labels, with its own preview-only format; production export format is unchanged. Actual touch,
assistive technology and real application persistence/history remain unverified. App unit/lint/build
and backend suites/build remain unrun: no production code changed.

The screenshot-aligned B refinement was verified at that point. The subsequent research request
rejects its mandatory Household container; the HTML now represents the earlier reviewed design.

### Home Assistant research verification

Observed the supplied live demo in Firefox through computer-use accessibility state and a desktop
screenshot. Living room, Kitchen, Energy, Climate, Study, Outdoor and Updates were visible as
sections containing individually presented results. Entered edit mode without altering data;
observed separate Edit card/Add card controls and Create section. Read Header settings without
changing values; no dashboard Save, card action, device command or provider configuration was used.
The demo was not resized and cross-section drag was not executed: those behaviours are documented,
not claimed as manually tested. The in-app browser was unavailable; Firefox was the working fallback.

Read current official concepts, cards, sections, tile, masonry and panel documentation using the
web tool. Recorded source URLs and bounded claims in the investigation; checked local links and
whitespace. No application code/HTML changed in this follow-up; no application tests were needed
or run. Independent-result rendering, breakpoint migration, stable service placement and panel
projection require a subsequent design/prototype update and production acceptance evidence.

### Current card-first preview verification

Applied `.agents/skills/sdd-plan/SKILL.md`, both component constitutions and Go/Vue/integration
engineering rules to the scope/ownership plan. No repository skill requests sub-agents, so none
were used. No production code, dependencies or package locks changed.

`node /private/tmp/nh-card-model-qa.cjs`: **164 checks passed** in installed isolated Chromium.
Four widths (390/1024/1280/1440), Home/panel, light/dark and seven accepted sample data states
include changed dates and distinct pending/empty/unavailable. Geometry checks verify no block
overlap and no page/preview horizontal overflow after layout settles. Group header geometry and
toolbar match the existing UI; move between groups/direct Home and native group/card drag work.
Optional summary and individual service cards read one accepted source catalog. Names remain text.

100 consecutive value updates retained the original card/link nodes and keyboard focus, with
maximum synchronous work **0.10ms** on this machine. 80 Home/panel mode cycles took **46.7ms**
total synchronous work. These are local preview measurements, not production browser/heap guarantees.
During the cycles there was exactly one observer, one listener AbortSignal, no intervals and no
detached observed nodes. The final four observed targets matched the four remaining live blocks.
After non-persisted pagehide: zero targets, pending animation frames, intervals or live export
URLs; every listener signal aborted. Persisted bfcache pagehide retained working controls.

Actual fixes during review: corrected conflicting CSS grid-row spans and seeded measurements
before first paint to remove overlap; visibility-only standalone edits no longer recreate/reorder
the block; summary uses a full group row; selected panel pages scroll into view on phones.
Final complete run passed after these changes. Source-owned application tests remain future work.

Zero browser page errors, HTTP requests, localStorage/sessionStorage entries or provider/device
commands. Exported file contains only synthetic placement references/preferences. Unknown imported
refs preserve accepted layout. Inspected desktop/mobile normal/edit and Garden panel screenshots
under `/private/tmp/nh-card-model-qa/`. Syntax, whitespace and local Markdown links passed.

Application `npm test`/lint/build and Go suite/build/vet/races were not run: application source
is unchanged. Actual touch/assistive-technology, real network cancellation, production migration,
privacy-safe compatibility and production heap/performance benchmarks remain unrun. Required
production checks and contract gates are listed above; no assurance of zero undiscovered bugs
or production memory/performance issues is inferred from this isolated preview evidence.
