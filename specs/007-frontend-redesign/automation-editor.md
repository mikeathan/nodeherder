# Automation editor redesign

Part of [plan 007](plan.md); acceptance AC-10…AC-15. Current behaviour:
[current-state §7](current-state.md#7-automations-most-complex-area).

## Mental model shown to users

```text
Automation "Living room switch dial"   [Enabled ▾]  [Schedule: on 07:00 · off 23:30]
┌ Trigger 1 ──────────────────────────────────────────────────── [Run] [⋯] ┐
│ “When Living room switch dial · action changes, if action = brightness_step_up │
│   and time is 07:00–23:30, then set Living Room Light brightness + 25.”        │
│ WHEN  [source device: fixed = automation device] [expose: action ▾]            │
│ IF    [action] [=] [brightness_step_up ▾]                    (+ condition)     │
│       [time between] [07:00] – [23:30]                                         │
│ THEN  1. [Step ▾] Living Room Light · brightness  + 25        ↑↓ 🗑           │
│       2. [Set  ▾] Attic room Light · state = ON · after 2 min ↑↓ 🗑 (+ action) │
└────────────────────────────────────────────────────────────────────────────────┘
(+ trigger)                                [Discard] [View JSON] [Save]
```

- One page, no panel stack. Each trigger is a card; collapsed state shows the sentence.
- **WHEN**: source is the automation's device (the data model is keyed by device id);
  expose picker lists non-config exposes with live current value next to each.
- **IF**: expose conditions (operator list restricted for binary/enum to `=`; value input
  by capability: toggle/enum select/number with unit) and time windows (24 h pickers,
  overnight ranges allowed and labelled "overnight"). Drag handle to reorder.
- **THEN**: ordered action list (Q-02: backend runs all actions in order). Action types
  with human names: *Set* (`trigger`: device, one or more expose=value rows, optional
  delay, publish mode batch/single shown only when > 1 row), *Step* (`step`: + − ×
  amount on a numeric expose), *Cycle presets* (`preset`).
- **Warnings** (non-blocking): device trigger with no conditions → "Will only run
  manually — the hub blocks device triggers without conditions"; target device offline;
  action targets the source device and expose (possible feedback loop; the backend has
  cooldown/loop prevention, show as info).
- **Errors** (block Save): missing source expose, action without target, set-row
  without expose or value, step without amount, invalid time.
- **Schedules**: separate card with `NhTimeline24`; add enable/disable times (max one of
  each, as today), shows hub timezone.
- **Run**: available when `canTriggerManually`; button locks while the request is in
  flight; result toast.
- **JSON view**: read-only pretty JSON of the exact `saveAutomation` payload (helps
  support/debugging; never editable to avoid bypassing validation).
- **Leave guard**: dirty draft → confirm dialog on route leave / tab close.

## Implementation shape

| Unit | Responsibility | Test |
| --- | --- | --- |
| `selectors/automation.ts` `describeTrigger(trigger, devices)` | Sentence tokens (text only) | Jest, all types |
| `selectors/automation.ts` `validateAutomation(a, devices)` | `{errors: FieldError[], warnings}` with paths like `triggers[0].actions[1].exposes[0].data` | Jest |
| `composables/useAutomationDraft(id)` | deep-cloned draft, dirty flag, `reset`, `toPayload()` (preserves original shape/optional fields), route leave guard | Jest round-trip (AC-14) |
| `components/automations/editor/AutomationEditor.vue` | page layout | manual |
| `…/TriggerCard.vue`, `WhenBlock.vue`, `ConditionRow.vue`, `ActionList.vue`, `SetAction.vue`, `StepAction.vue`, `PresetAction.vue`, `ScheduleCard.vue`, `PayloadPreview.vue` | blocks | manual + keyboard |

Existing `contracts/automations.ts` helpers (`createConditionFromType`,
`createActionFromType`, `isValid`, `canTriggerManually`) are reused; `isValid` is
superseded by `validateAutomation` but kept until callers are migrated.
`formatTriggerConditions` (currently rendered with `v-html`) is replaced by
`describeTrigger` + `NhSentence` (FE-05).

Automations list: rows show name, description, device, trigger count, status pill
(Enabled/Disabled/Scheduled with next change time), quick enable toggle, Run (if a
manual-capable trigger exists), delete with scope text ("Deletes 3 triggers; cannot be
undone").
