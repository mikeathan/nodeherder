import { computed, watch } from 'vue';
import { store } from '@/store/index';

export function useAutomationsLoader() {
  const automations = computed(() => store.getters['automations/listAll']());

  watch(
    () => store.getters['ws/getConnectionStatus'],
    (status) => {
      if (status === 'connected' && !store.getters['automations/initialized']()) {
        store.dispatch('ws/emit', { event: 'loadAutomations' });
      }
    },
    { immediate: true }
  );

  return { automations };
}
