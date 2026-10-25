import { onMounted } from 'vue';
import { useConnections } from '@/composables/useConnections';
import { RouteName } from '@/types/router';

/** Owns catalog loading and navigation after an admitted setup. */
export function useConnectionsCatalog() {
  const { api, state, store, router, busy, error, perform, signal } = useConnections();
  const load = () => perform(() => store.dispatch('integrations/load', signal));
  onMounted(load);
  async function add(key: string) {
    const setup = await perform(() => api.createSetup(key));
    if (setup) await router.push({ name: RouteName.ConnectionSetup, params: { id: setup.id } });
  }
  async function reloadProviders() {
    const value = await perform(() => api.reloadProviders());
    if (value) await load();
  }
  const providerRoute = (key: string) => {
    const [type, provider] = key.split('/');
    return { name: RouteName.ConnectionProvider, params: { type, provider } };
  };
  return { state, busy, error, load, add, reloadProviders, providerRoute };
}
