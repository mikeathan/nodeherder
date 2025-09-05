import { computed, watch } from 'vue';
import { store } from '@/store/index';
import { Automations } from '@/types/automation.type';

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
