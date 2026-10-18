import { computed, ref } from 'vue';
import type { Expose } from '@/types/device';
import { findDevice, statusOf } from './useHub';
import { controlOf, isWritable } from '@/domain/exposes';

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
    if (['brightness', 'color_temp'].includes(exposeName))
      return ['brightness', 'color_temp', 'state'].includes(e.name);
    if (['energy', 'power'].includes(exposeName)) return ['state', 'child_lock'].includes(e.name);
    return false;
  }
  return { device, expose, adjustment, adjustable, selected, controls, related, status };
}
