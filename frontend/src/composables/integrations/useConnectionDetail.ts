import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useConnections } from '@/composables/useConnections';
import { confirm } from '@/composables/useConfirm';
import { RouteName } from '@/types/router';
import { ConnectionError } from '@/services/integrations.service';
import type { ConnectionInstance, ProviderDefinition } from '@/types/integrations.type';
import { integrationDetailSection, type IntegrationNavigationContext } from '@/domain/integration-navigation';

/** Owns this page's application actions; the route component only presents its state. */
export function useConnectionDetail() {
  const route = useRoute();
  const { api, isActive, state, store, router, busy, error, perform, observe, signal, commit } = useConnections();
  const missing = ref(false);
  const instanceId = computed(() => String(route.params.id));
  // The board already holds validated private DTOs. Keep the selected identity/snapshot
  // visible during revalidation, rather than flashing a generic title and empty page.
  const item = computed(() => {
    if (missing.value) return null;
    return state.value.current?.id === instanceId.value
      ? state.value.current
      : (state.value.instances.find((entry) => entry.id === instanceId.value) ?? null);
  });
  const tab = computed(() => integrationDetailSection(route.hash));
  const initialName = item.value?.name ?? '';
  const name = ref(initialName);
  const token = ref('');
  const definition = ref<ProviderDefinition | null>(null);
  const sectionContext = computed<IntegrationNavigationContext>(() => ({
    providerKey: item.value?.provider_key ?? '',
    instanceId: item.value?.id,
    query: route.query,
  }));
  async function reload(requestSignal = signal): Promise<ConnectionInstance> {
    try {
      const value: ConnectionInstance = await store.dispatch('integrations/loadInstance', {
        id: instanceId.value,
        signal: requestSignal,
      });
      if (isActive() && !requestSignal.aborted) missing.value = false;
      return value;
    } catch (cause) {
      // Cached presentation must not resurrect a connection deleted in another session.
      if (isActive() && !requestSignal.aborted && cause instanceof ConnectionError && cause.status === 404) {
        missing.value = true;
        if (state.value.current?.id === instanceId.value) commit('current', null);
        commit(
          'instances',
          state.value.instances.filter((entry) => entry.id !== instanceId.value)
        );
      }
      throw cause;
    }
  }
  /** Retry the private instance read without submitting a fresh provider lookup. */
  const reloadPage = () => perform(() => reload());
  let projection: ReturnType<typeof setInterval> | undefined;
  const onFocus = () => {
    if (isActive() && !busy.value && !document.hidden) void perform(() => reload());
  };
  onMounted(async () => {
    projection = setInterval(onFocus, 60_000);
    window.addEventListener('focus', onFocus);
    // One owned loading interval covers both reads; a failed instance read must not be
    // followed by a provider read that clears its error or uses a stale binding.
    await perform(async () => {
      const current = await reload();
      if (!isActive()) return;
      // Setup is usable from cache; revalidation must not replace a name the user has typed.
      if (name.value === initialName) name.value = current.name;
      const provider = await api.definition(current.provider_key, signal);
      if (isActive()) {
        definition.value = provider.document;
        commit('definition', provider);
      }
    });
    if (isActive() && item.value?.operation_id) observe(item.value.operation_id, reload);
  });
  function stopProjection() {
    if (projection) clearInterval(projection);
    projection = undefined;
    window.removeEventListener('focus', onFocus);
    token.value = '';
  }
  signal.addEventListener('abort', stopProjection, { once: true });
  onBeforeUnmount(stopProjection);
  /** Admit a manual provider refresh using the accepted instance revision and observe its progress. */
  async function refresh() {
    const current = item.value;
    if (!current) return;
    const operation = await perform(() => api.refresh(current.id, current.etag));
    if (operation) observe(operation.id, reload);
  }
  /** Continue the current admitted operation without creating a second lookup. */
  async function check() {
    const current = item.value;
    if (!current) return;
    const operationId = current.operation_id;
    if (!operationId) return;
    const operation = await perform(() => api.check(operationId, current.etag));
    if (operation) observe(operation.id, reload);
  }
  /** Save name/enabled state with revision fencing, then reload the accepted resource view. */
  async function update(enabled: boolean) {
    const current = item.value;
    if (!current) return;
    const result = await perform(() => api.update(current.id, current.etag, name.value || current.name, enabled));
    if (result) await perform(() => reload());
  }
  /** Start isolated address setup; existing dates remain available until replacement is verified. */
  async function editAddress() {
    const current = item.value;
    if (!current) return;
    const p = await perform(() => api.createSetup(current.provider_key, { id: current.id, etag: current.etag }));
    if (p) await router.push({ name: RouteName.IntegrationSetup, params: { id: p.id } });
  }
  /** Save/remove a named token slot, clearing plaintext after the request and reloading redacted state. */
  async function credential(remove = false, savedSlot?: string) {
    const current = item.value;
    if (!current) return;
    const slot = savedSlot ?? definition.value?.auth.secret_slot;
    if (!slot) return;
    if (
      remove &&
      !(await confirm({
        title: 'Remove provider token?',
        message:
          'Future requests will use anonymous access if the provider allows it. Collection dates already saved remain available.',
        confirmLabel: 'Remove token',
        danger: true,
      }))
    )
      return;
    const result = await perform(() =>
      api.credential('instances', current.id, slot, current.etag, token.value, remove)
    );
    token.value = '';
    if (result) await perform(() => reload());
  }
  /** Confirm destructive deletion, clear the current private DTO and return to the board. */
  async function remove() {
    const current = item.value;
    if (!current) return;
    if (
      !(await confirm({
        title: 'Remove integration?',
        message:
          'This deletes the saved address, credentials and collection data from NodeHerder. The shared provider configuration is kept. A lookup already submitted to the provider cannot be recalled.',
        confirmLabel: 'Remove integration',
        danger: true,
      }))
    )
      return;
    const result = await perform(() => api.remove(current.id, current.etag));
    if (result) {
      commit('current', null);
      await router.replace({ name: RouteName.Integrations });
    }
  }
  return {
    item,
    tab,
    name,
    token,
    definition,
    sectionContext,
    busy,
    error,
    refresh,
    check,
    update,
    editAddress,
    credential,
    remove,
    reloadPage,
  };
}
