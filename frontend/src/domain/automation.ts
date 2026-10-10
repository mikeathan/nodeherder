/*
 * Automation editor logic (spec 007 US-04, AC-10…AC-15). Framework-free.
 * Paths use dot notation into the automation payload, e.g. "triggers.0.actions.1.id",
 * so the editor can attach messages to the exact field.
 */
import { Device, Expose } from '@/types/device';
import {
  Automation,
  AutomationAction,
  AutomationCondition,
  AutomationStepAction,
  AutomationTrigger,
  isExposeCondition,
  isTimeCondition,
} from '@/types/automation.type';
import { DeviceAutomation, isPresetCyclingAction, isStepAction, isTriggerAction } from '@/contracts/automations';
import { exposeLabel, formatExposeValue } from './exposes';

export type DeviceLookup = (id: string) => Device | undefined;
export type Issue = { path: string; message: string };
export type Validation = { errors: Issue[]; warnings: Issue[] };

const HH_MM = /^([01]\d|2[0-3]):[0-5]\d$/;
const blank = (v: unknown) => v === null || v === undefined || v === '';

const valueText = (device: Device | undefined, exposeName: string, value: unknown): string => {
  const e: Expose | undefined = device?.exposes[exposeName];
  if (blank(value)) return '…';
  return e ? formatExposeValue(e, value) : String(value);
};
const lower = (name: string) => (name ? exposeLabel(name).toLowerCase() : '…');

export function describeCondition(automation: Automation, c: AutomationCondition, lookup: DeviceLookup): string {
  if (isTimeCondition(c)) {
    const { startAt, endAt } = c.timeRange;
    const overnight = startAt && endAt && startAt > endAt ? ' (overnight)' : '';
    return `the time is between ${startAt || '…'} and ${endAt || '…'}${overnight}`;
  }
  if (isExposeCondition(c)) {
    return `${lower(c.name)} ${c.equality} ${valueText(lookup(automation.id), c.name, c.value)}`;
  }
  return 'an unknown condition';
}

const OPERATOR_SYMBOL: Record<string, string> = { '+': '+', '-': '−', '*': '×' };

/**
 * Step actions compute property = step0 op0 (step1 op1 (… amount)), evaluated from the last
 * step (backend operations.go stepOperation.CreatePayload).
 */
export function stepFormula(a: AutomationStepAction): string {
  let acc = blank(a.data) ? '…' : String(a.data);
  for (let i = a.steps.length - 1; i >= 0; i--) {
    const step = a.steps[i];
    const inner = i === a.steps.length - 1 ? acc : `(${acc})`;
    acc = `${lower(step.property)} ${OPERATOR_SYMBOL[step.operator] ?? step.operator} ${inner}`;
  }
  return acc;
}

export function describeAction(a: AutomationAction, lookup: DeviceLookup): string {
  const device = a.id ? lookup(a.id) : undefined;
  const target = device?.friendly_name ?? (a.id ? a.id : '…');
  if (isTriggerAction(a)) {
    const values = a.exposes.length ? a.exposes.map((x) => `${lower(x.name)} to ${valueText(device, x.name, x.data)}`).join(', ') : '…';
    const wait = a.delay && a.delay.value > 0 ? ` after ${a.delay.value} ${a.delay.unit}` : '';
    return `set ${target} ${values}${wait}`;
  }
  if (isStepAction(a)) return `set ${target} ${lower(a.property)} to ${stepFormula(a)}`;
  if (isPresetCyclingAction(a)) return `cycle ${target} ${lower(a.property)} presets`;
  return 'do something unknown';
}

/** Sentence summary shown on trigger cards and in the automation list (AC-10). Text only (FE-05). */
export function describeTrigger(automation: Automation, t: AutomationTrigger | undefined, lookup: DeviceLookup): string {
  if (!t) return 'No triggers yet.';
  const source = lookup(automation.id)?.friendly_name ?? automation.friendlyname;
  // The hub only accepts device triggers (backend triggerTypeRegistry); running by hand is a
  // separate request, not a trigger type.
  const when = `When ${source} ${lower(t.name)} changes`;
  const ifs = t.conditions.length ? `, if ${t.conditions.map((c) => describeCondition(automation, c, lookup)).join(' and ')}` : '';
  const thens = t.actions.length ? `, then ${t.actions.map((a) => describeAction(a, lookup)).join(', then ')}` : ', then …';
  return `${when}${ifs}${thens}.`;
}

export function validateAutomation(a: Automation, lookup: DeviceLookup): Validation {
  const errors: Issue[] = [];
  const warnings: Issue[] = [];
  const err = (path: string, message: string) => errors.push({ path, message });
  const warn = (path: string, message: string) => warnings.push({ path, message });

  if (!a.triggers.length) warn('triggers', 'This automation has no triggers, so it never runs.');
  a.triggers.forEach((t, i) => {
    const tp = `triggers.${i}`;
    if (!t.name) err(`${tp}.name`, 'Choose which property starts this trigger.');
    if (t.conditions.length === 0) {
      // backend/internal/automations/trigger.go: device-triggered runs without conditions are blocked
      warn(tp, 'No conditions: the hub only runs this trigger when you run it by hand.');
    }
    t.conditions.forEach((c, j) => {
      const cp = `${tp}.conditions.${j}`;
      if (isTimeCondition(c)) {
        if (!HH_MM.test(c.timeRange.startAt)) err(`${cp}.timeRange.startAt`, 'Enter a start time.');
        if (!HH_MM.test(c.timeRange.endAt)) err(`${cp}.timeRange.endAt`, 'Enter an end time.');
      } else if (isExposeCondition(c)) {
        if (!c.name) err(`${cp}.name`, 'Choose a property to compare.');
        else if (blank(c.value)) err(`${cp}.value`, 'Enter a value to compare with.');
      }
    });
    if (!t.actions.length) err(`${tp}.actions`, 'Add at least one action.');
    t.actions.forEach((x, k) => {
      const ap = `${tp}.actions.${k}`;
      if (!x.id) return err(`${ap}.id`, 'Choose a device to control.');
      const target = lookup(x.id);
      if (!target) err(`${ap}.id`, 'This device no longer exists.');
      else if (target.availability === 'offline') warn(ap, `${target.friendly_name} is offline; the command waits until it reconnects.`);
      if (isTriggerAction(x)) {
        if (!x.exposes.length) err(`${ap}.exposes`, 'Add a value to set.');
        x.exposes.forEach((row, n) => {
          if (!row.name) err(`${ap}.exposes.${n}.name`, 'Choose a property.');
          else if (blank(row.data)) err(`${ap}.exposes.${n}.data`, 'Enter a value.');
          if (x.id === a.id && row.name && row.name === t.name) warn(ap, 'This sets the property that starts the trigger. The hub prevents loops, but check it is intended.');
        });
        if (x.delay && (!Number.isFinite(x.delay.value) || x.delay.value < 0)) err(`${ap}.delay.value`, 'Enter a wait of 0 or more.');
      } else if (isStepAction(x)) {
        if (!x.property) err(`${ap}.property`, 'Choose a number property.');
        // the hub reads the amount as a JSON number (operations.go), never a string
        if (typeof x.data !== 'number' || !Number.isFinite(x.data)) err(`${ap}.data`, 'Enter an amount.');
        if (!x.steps.length) err(`${ap}.steps`, 'Add at least one step.');
        x.steps.forEach((step, n) => {
          if (!step.id || !step.property) err(`${ap}.steps.${n}`, 'Choose the device and value for this step.');
        });
      } else if (isPresetCyclingAction(x)) {
        if (!x.property) err(`${ap}.property`, 'Choose a property with presets.');
      }
    });
  });
  return { errors, warnings };
}

export const issuesUnder = (issues: Issue[], prefix: string): Issue[] =>
  issues.filter((x) => x.path === prefix || x.path.startsWith(`${prefix}.`));

/** A new, empty automation for a device (automations are keyed by the device that triggers them). */
export function newAutomation(device: Pick<Device, 'id' | 'friendly_name'>): Automation {
  const a = new DeviceAutomation();
  a.id = device.id;
  a.friendlyname = device.friendly_name;
  return { ...a, triggers: [newTrigger()] };
}

/** The hub only accepts device triggers (backend triggerTypeRegistry). */
export function newTrigger(): AutomationTrigger {
  return { name: '', type: 'deviceTrigger', conditions: [], actions: [] };
}

/** Deep copy through JSON, which is exactly what is sent to the hub. */
export const cloneAutomation = (a: Automation): Automation => JSON.parse(JSON.stringify(a));
export const sameAutomation = (a: Automation, b: Automation): boolean => JSON.stringify(a) === JSON.stringify(b);
