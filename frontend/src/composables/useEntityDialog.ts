import { computed, ref } from 'vue';
import type { Expose } from '@/types/device';
import { findDevice, statusOf } from './useHub';
import { controlOf, isWritable } from '@/domain/exposes';

/** Related controls shown beside the selected expose; capability comes from expose metadata. */
const RELATED: Record<string, string[]> = {
  brightness: ['brightness', 'color_temp', 'state'],
  color_temp: ['brightness', 'color_temp', 'state'],
  energy: ['state', 'child_lock'],
  power: ['state', 'child_lock'],
  smoke: ['silence', 'test'],
  water_leak: ['silence', 'test'],
  tamper: ['silence', 'test'],
  device_fault: ['silence', 'test'],
  alarm: ['silence', 'test'],
};

/** Selects presentation only. Device truth and commands have separate shared owners. */
export function useEntityDialog(deviceId: string, exposeName: string) {
  const selected = ref(exposeName);
  const device = computed(() => findDevice(deviceId));
  const expose = computed(() => device.value?.exposes[exposeName]);
  const status = computed(() => (device.value ? statusOf(device.value) : 'missing'));
  const related = computed(() => Object.values(device.value?.exposes ?? {}).filter(relevantControl));
  const adjustment = computed(() => device.value?.exposes[selected.value]);
  const adjustable = computed(() => !!adjustment.value && controlOf(adjustment.value) === 'slider');
  const controls = computed(() => related.value.filter((e) => e.name !== selected.value));
  function relevantControl(e: Expose): boolean {
    if (!isWritable(e)) return false;
    if (e.name === exposeName) return true;
    return RELATED[exposeName]?.includes(e.name) ?? false;
  }
  return { device, expose, adjustment, adjustable, selected, controls, related, status };
}
