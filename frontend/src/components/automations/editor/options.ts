/*
 * Option lists for the automation editor. Filters come from configs/automation/device.config
 * (the same rules the previous editor used), labels from domain/exposes.
 */
import { Device, Expose } from '@/types/device';
import { ExposeAccessModes, ExposeCategories, ExposeTypes } from '@/types/device.type';
import { exposeLabel } from '@/domain/exposes';
import { isPresetExpose } from '@/configs/automation/device.config';
import { EqualityOperators, NumericOperators } from '@/contracts/automations';
import type { SelectOption } from '@/components/ui/UiSelect.vue';

export type ExposeFilter = (e: Expose) => boolean;

export const writable: ExposeFilter = (e) => e.access_mode !== ExposeAccessModes.Read;
export const notConfig: ExposeFilter = (e) => e.category !== ExposeCategories.Config;
export const numericWritable: ExposeFilter = (e) => e.type === ExposeTypes.Numeric && writable(e) && e.category === ExposeCategories.Measurement;
export const numericMeasurement: ExposeFilter = (e) => e.type === ExposeTypes.Numeric && e.category === ExposeCategories.Measurement;
export const hasPresets: ExposeFilter = (e) => isPresetExpose(e);

export function exposeOptions(device: Device | undefined, filter: ExposeFilter, keep?: string): SelectOption[] {
  const list = Object.values(device?.exposes ?? {})
    .filter(filter)
    .map((e) => ({ value: e.name, label: exposeLabel(e.name) }))
    .sort((a, b) => a.label.localeCompare(b.label));
  // keep a stored value visible even if the device no longer offers it
  if (keep && !list.some((o) => o.value === keep)) list.unshift({ value: keep, label: `${exposeLabel(keep)} (not reported)` });
  return list;
}

export function deviceOptions(devices: Device[], filter: ExposeFilter, keep?: string): SelectOption[] {
  const list: SelectOption[] = devices.filter((d) => Object.values(d.exposes).some(filter)).map((d) => ({ value: d.id, label: d.friendly_name }));
  if (keep && !list.some((o) => o.value === keep)) list.unshift({ value: keep, label: `${keep} (not found)` });
  return list;
}

/** Binary and enum values only compare for equality (as in the previous editor). */
export function comparisonOptions(expose: Expose | undefined): SelectOption[] {
  const ops = !expose || expose.type === ExposeTypes.Numeric ? EqualityOperators : ['='];
  const words: Record<string, string> = { '=': 'is', '<': 'below', '<=': 'at most', '>': 'above', '>=': 'at least' };
  return ops.map((op) => ({ value: op, label: `${op}  ${words[op] ?? ''}`.trim() }));
}

export const STEP_OPERATORS: SelectOption[] = NumericOperators.map((op) => ({ value: op, label: op === '+' ? '+ add' : op === '-' ? '− subtract' : '× multiply' }));
