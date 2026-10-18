import { computed, ref, watch } from 'vue';
import type { Device, Expose } from '@/types/device';
import { exposeRange } from '@/domain/exposes';
import { invalidRange, nativeNumber } from '@/domain/controls';
import { useDeviceCommand } from './useDeviceCommand';

/** One draft/confirmation policy shared by device rows and full-screen controls. */
export function useExposeEdit(device: () => Device, expose: () => Expose) {
  const command = useDeviceCommand();
  const draft = ref<number | null>(null);
  const dragging = ref(false);
  const pending = computed(() => command.isPending(device().id, expose().name));
  const disabled = computed(() => !command.canControl(device()) || pending.value || invalidRange(expose()));
  // A native range needs a finite position; its display must still say unknown
  // until a real reading or a user draft exists.
  const value = computed(() => draft.value ?? confirmedNumber(expose()));
  const displayValue = computed(() => draft.value ?? expose().data);
  function cancel() {
    dragging.value = false;
    draft.value = null;
  }
  function preview(value: number) {
    draft.value = value;
  }
  function commit(value: unknown) {
    dragging.value = false;
    if (disabled.value) {
      cancel();
      return;
    }
    if (!command.send(device(), expose().name, value)) cancel();
  }
  function commitNumber(value: number) {
    const native = nativeNumber(expose(), value);
    if (native === null) {
      cancel();
      return;
    }
    commit(native);
  }
  watch(pending, (now, was) => {
    if (was && !now) cancel();
  });
  watch(disabled, (value) => {
    if (value && !pending.value) cancel();
  });
  watch(
    () => expose().data,
    () => {
      if (!dragging.value) draft.value = null;
    }
  );
  watch(() => device().id + '|' + expose().name, cancel);
  return { value, displayValue, pending, disabled, dragging, preview, commit, commitNumber, cancel };
}
function confirmedNumber(expose: Expose): number {
  if (typeof expose.data === 'number' && Number.isFinite(expose.data)) return expose.data;
  return exposeRange(expose)?.min ?? 0;
}
