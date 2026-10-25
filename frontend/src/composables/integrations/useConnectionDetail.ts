import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useConnections } from '@/composables/useConnections';
import { confirm } from '@/composables/useConfirm';
import { RouteName } from '@/types/router';
import type { ProviderDefinition } from '@/types/integrations.type';

/** Owns this page's application actions; the route component only presents its state. */
export function useConnectionDetail() {
  const route = useRoute();
  const { api, isActive, state, store, router, busy, error, perform, observe, signal } = useConnections();
  const item = computed(() => (state.value.current?.id === route.params.id ? state.value.current : null));
  const tab = ref('results');
  const name = ref('');
  const token = ref('');
  const definition = ref<ProviderDefinition | null>(null);
  const providerParams = computed(() => {
    const [type, provider] = (item.value?.provider_key ?? '/').split('/');
    return { type, provider };
  });
  const reload = (requestSignal = signal) =>
    store.dispatch('integrations/loadInstance', { id: String(route.params.id), signal: requestSignal });
  const reloadPage = () => perform(() => reload());
  let projection: ReturnType<typeof setInterval> | undefined;
  const onFocus = () => {
    if (isActive() && !busy.value && !document.hidden) void perform(() => reload());
  };
  onMounted(async () => {
    projection = setInterval(onFocus, 60_000);
    window.addEventListener('focus', onFocus);
    await perform(() => reload());
    const current = item.value;
    if (current && isActive()) {
      name.value = current.name;
      const provider = await perform(() => api.definition(current.provider_key, signal));
      if (provider) definition.value = provider.document;
      if (current.operation_id) observe(current.operation_id, reload);
    }
  });
  signal.addEventListener(
    'abort',
    () => {
      token.value = '';
    },
    { once: true }
  );
  onBeforeUnmount(() => {
    if (projection) clearInterval(projection);
    window.removeEventListener('focus', onFocus);
    token.value = '';
  });
  async function refresh() {
    const current = item.value;
    if (!current) return;
    const operation = await perform(() => api.refresh(current.id, current.etag));
    if (operation) observe(operation.id, reload);
  }
  async function check() {
    const current = item.value;
    if (!current) return;
    const operationId = current.operation_id;
    if (!operationId) return;
    const operation = await perform(() => api.check(operationId, current.etag));
    if (operation) observe(operation.id, reload);
  }
  async function update(enabled: boolean) {
    const current = item.value;
    if (!current) return;
    const result = await perform(() => api.update(current.id, current.etag, name.value || current.name, enabled));
    if (result) await perform(() => reload());
  }
  async function editAddress() {
    const current = item.value;
    if (!current) return;
    const p = await perform(() => api.createSetup(current.provider_key, { id: current.id, etag: current.etag }));
    if (p) await router.push({ name: RouteName.ConnectionSetup, params: { id: p.id } });
  }
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
  async function remove() {
    const current = item.value;
    if (!current) return;
    if (
      !(await confirm({
        title: 'Remove connection?',
        message:
          'This deletes the saved address, credentials and collection data from NodeHerder. The shared provider configuration is kept. A lookup already submitted to the provider cannot be recalled.',
        confirmLabel: 'Remove connection',
        danger: true,
      }))
    )
      return;
    const result = await perform(() => api.remove(current.id, current.etag));
    if (result) await router.replace({ name: RouteName.Connections });
  }
  return {
    item,
    tab,
    name,
    token,
    definition,
    providerParams,
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
