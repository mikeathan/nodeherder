import { ref } from 'vue';
import type { ConnectionScope } from '@/composables/useConnections';
import { RouteName } from '@/types/router';

/** Owns the separate Add dialog and whitelisted setup navigation, not board presentation. */
export function useConnectionsCatalog(scope: ConnectionScope) {
  const open = ref(false);
  /** Create a shared setup session and navigate with a named board-return intent, never a URL redirect. */
  async function add(key: string) {
    const setup = await scope.perform(() => scope.api.createSetup(key));
    if (setup) {
      open.value = false;
      await scope.router.push({
        name: RouteName.IntegrationSetup,
        params: { id: setup.id },
        query: { return_to: 'board' },
      });
    }
  }
  /** Reload host definitions and update discovery; existing saved connections remain separate state. */
  async function reloadProviders() {
    const catalog = await scope.perform(() => scope.api.reloadProviders());
    if (catalog) scope.commit('catalog', catalog);
  }
  /** Resolve a validated provider key to the existing protected editor route. */
  const providerRoute = (key: string) => {
    const [type, provider] = key.split('/');
    return { name: RouteName.IntegrationProvider, params: { type, provider } };
  };
  return { open, add, reloadProviders, providerRoute };
}
