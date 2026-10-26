import { computed } from 'vue';
import { useRoute } from 'vue-router';
import { integrationInstanceId, type IntegrationNavigationContext } from '@/domain/integration-navigation';

/** Expose provider/Help navigation from explicit route context, without loading or guessing
 * an instance. This readonly projection owns no subscriptions, private answers or HTTP requests.
 */
export function useProviderSections() {
  const route = useRoute();
  return computed<IntegrationNavigationContext>(() => ({
    providerKey: `${route.params.type}/${route.params.provider}`,
    instanceId: integrationInstanceId(route.query),
    query: route.query,
  }));
}
