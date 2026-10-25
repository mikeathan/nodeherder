import { computed, onMounted, onBeforeUnmount, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useConnections } from '@/composables/useConnections';
import { confirm } from '@/composables/useConfirm';
import { RouteName } from '@/types/router';
import type { ProviderDefinition } from '@/types/integrations.type';

/** Owns this page's application actions; the route component only presents its state. */
export function useConnectionSetup() {
  const route = useRoute();
  const { api, isActive, state, router, busy, error, perform, commit, observe, signal } = useConnections();
  const setup = computed(() => (state.value.setup?.id === route.params.id ? state.value.setup : null));
  const postcode = ref('');
  const property = ref('');
  const name = ref('Bin collections');
  const token = ref('');
  const definition = ref<ProviderDefinition | null>(null);
  async function reload(requestSignal = signal) {
    const value = await api.setup(String(route.params.id), requestSignal);
    if (!isActive() || requestSignal.aborted) return value;
    commit('setup', value);
    if (value.instance_id && value.status === 'completed')
      await router.replace({ name: RouteName.Connection, params: { id: value.instance_id } });
    return value;
  }
  onMounted(async () => {
    const value = await perform(() => reload());
    if (!value) return;
    postcode.value = value.postcode ?? '';
    const provider = await perform(() => api.definition(value.provider_key, signal));
    if (provider) definition.value = provider.document;
    if (value.operation_id && value.status === 'searching') observe(value.operation_id, reload);
  });
  signal.addEventListener(
    'abort',
    () => {
      token.value = '';
    },
    { once: true }
  );
  onBeforeUnmount(() => {
    token.value = '';
  });
  async function search() {
    const current = setup.value;
    if (!current) return;
    property.value = '';
    const value = await perform(() =>
      api.step(current.id, current.etag, { action: 'search', postcode: postcode.value })
    );
    if (value) {
      commit('setup', value);
      if (value.operation_id) observe(value.operation_id, reload);
    }
  }
  async function save() {
    const current = setup.value;
    if (!current) return;
    const value = await perform(() =>
      api.step(current.id, current.etag, {
        action: 'save',
        property_id: property.value,
        name: name.value,
      })
    );
    if (!value) return;
    commit('setup', value);
    if (value.instance_id) await router.replace({ name: RouteName.Connection, params: { id: value.instance_id } });
    else if (value.operation_id) observe(value.operation_id, reload);
  }
  async function check() {
    const current = setup.value;
    if (!current) return;
    const operationId = current.operation_id;
    if (!operationId) return;
    const value = await perform(() => api.check(operationId, current.etag));
    if (value) observe(value.id, reload);
  }
  async function saveToken() {
    const current = setup.value;
    if (!current) return;
    const slot = definition.value?.auth.secret_slot;
    if (!slot) return;
    const value = await perform(() => api.credential('setup', current.id, slot, current.etag, token.value));
    token.value = '';
    if (value && 'choices' in value) commit('setup', value);
  }
  async function cancel() {
    const current = setup.value;
    if (!current) return;
    if (
      !(await confirm({
        title: 'Discard setup?',
        message:
          'This removes this unfinished setup and its saved inputs. A lookup already submitted to the provider cannot be cancelled.',
        confirmLabel: 'Discard setup',
        danger: true,
      }))
    )
      return;
    const value = await perform(() => api.cancel(current.id, current.etag));
    if (value) await router.replace({ name: RouteName.Connections });
  }
  const reloadPage = () => perform(() => reload());
  return {
    reloadPage,
    setup,
    postcode,
    property,
    name,
    token,
    definition,
    busy,
    error,
    search,
    save,
    check,
    saveToken,
    cancel,
  };
}
