import { ref } from 'vue';
import type { ConnectionScope } from '@/composables/useConnections';
import type { ConnectionCatalogProvider } from '@/types/integrations';
import { RouteName } from '@/types/router';

/**
 * Local providers need no setup input, so they are added in one action.
 *
 * Kept in step with the backend's `Local` marker: both describe the same distinction, and the
 * catalog payload is the only place the frontend can learn it without guessing from a provider key.
 */
function isLocalProvider(provider: ConnectionCatalogProvider | undefined): boolean {
  return provider?.origin === 'local';
}

/**
 * Owns the separate Add dialog and whitelisted setup navigation, not board presentation.
 *
 * `reload` is supplied by the board because an authored provider returns from its wizard through
 * navigation, which reloads the page. A local provider never leaves the page, so adding one has to
 * refresh the board explicitly or the new instance has nothing to render it.
 */
export function useConnectionsCatalog(scope: ConnectionScope, reload?: () => Promise<void>) {
  const open = ref(false);
  /**
   * Add a provider.
   *
   * A local provider is added directly: its "setup" would be a wizard with nothing to ask, so the
   * operator would click through empty steps. Everything else keeps the existing wizard.
   */
  async function add(key: string) {
    const provider = scope.state.value.catalog?.providers.find((entry) => entry.key === key);
    if (isLocalProvider(provider)) {
      const added = await scope.perform(() => scope.api.addInstance(key));
      if (!added) return;
      open.value = false;
      // Catalog counts and the board both come from this read, so one reload updates both.
      await reload?.();
      return;
    }
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
