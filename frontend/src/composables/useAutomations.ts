/*
 * Automations for screens (adapter over the Vuex `automations` module and the manual-run
 * endpoint). Loads automations once the hub link is up, as the previous UI did.
 */
import { computed, watch } from 'vue';
import { store } from '@/store';
import { Automation, Automations } from '@/types/automation.type';
import { triggerAutomation } from '@/services/automation-trigger.service';

export function useAutomationsLoader() {
  const automations = computed(() => store.getters['automations/listAll']() as Automations);
  watch(
    () => store.getters['ws/getConnectionStatus'],
    (status) => {
      if (status === 'connected' && !store.getters['automations/initialized']()) {
        store.dispatch('ws/emit', { event: 'loadAutomations' });
      }
    },
    { immediate: true }
  );
  return automations;
}

export function useAutomations() {
  const automations = useAutomationsLoader();
  const loaded = computed(() => store.getters['automations/initialized']() as boolean);
  return {
    automations,
    loaded,
    find: (id: string) => store.getters['automations/find'](id) as Automation | undefined,
    save: (automation: Automation) => store.dispatch('automations/save', automation),
    remove: (id: string) => store.dispatch('automations/delete', id),
    setEnabled: (automation: Automation, enabled: boolean) => store.dispatch('automations/save', { ...automation, enabled }),
    /** Runs one trigger's actions now (POST automation/trigger). */
    async run(automation: Automation, triggerName: string) {
      const result = await triggerAutomation(automation, triggerName);
      if (result.success) store.dispatch('alerts/showSuccess', `Ran “${automation.friendlyname}”.`);
      else store.dispatch('alerts/showError', `Could not run “${automation.friendlyname}”: ${result.error ?? 'unknown error'}`);
      return result.success;
    },
  };
}
