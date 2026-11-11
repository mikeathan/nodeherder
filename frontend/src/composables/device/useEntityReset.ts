/*
 * Card-level reset orchestration for provider entities with catalogued defaults.
 * Commands go through useDeviceCommand so pending, confirmation, offline and no-retry
 * behaviour stay identical to per-row edits (NH-02, FE-03).
 */
import { computed } from 'vue';
import type { Device, Expose } from '@/types/device';
import { store } from '@/store';
import { confirm } from '@/composables/useConfirm';
import { useDeviceCommand } from '@/composables/useDeviceCommand';
import { entityDefaultTargets, resetDefaultsMessage } from '@/domain/entityDefaults';

/** Resolves reset targets, their availability and the confirmed bulk action. */
export function useEntityReset(device: () => Device, entities: () => readonly Expose[]) {
  const command = useDeviceCommand();
  const targets = computed(() => entityDefaultTargets(entities()));
  const pending = computed(() =>
    targets.value.some((target) => command.isPending(device().id, target.expose.name))
  );
  const canReset = computed(
    () => targets.value.length > 0 && command.canControl(device()) && !pending.value
  );

  /** Confirms the named targets, sends each existing command and reports requests, not confirmations. */
  async function resetDefaults() {
    if (!canReset.value) return;
    const accepted = await confirm({
      title: 'Reset configuration defaults',
      message: resetDefaultsMessage(targets.value),
      confirmLabel: 'Reset',
      danger: true,
    });
    if (!accepted) return;
    const sent = targets.value.filter((target) => command.send(device(), target.expose.name, target.value)).length;
    if (sent) store.dispatch('alerts/showSuccess', `Reset requested for ${sent} setting${sent === 1 ? '' : 's'}.`);
    if (sent < targets.value.length) {
      store.dispatch('alerts/showError', 'Some settings could not be reset. Check the device connection and try again.');
    }
  }

  return { targets, canReset, resetDefaults };
}
