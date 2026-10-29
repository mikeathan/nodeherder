import { setupPresentation } from '@/integrations/registry';
import { computed, onMounted, onBeforeUnmount, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { useConnections } from '@/composables/useConnections';
import { confirm } from '@/composables/useConfirm';
import { RouteName } from '@/types/router';
import type { ProviderDefinition } from '@/types/integrations.type';

/** Owns this page's application actions; the route component only presents its state. */
export function useConnectionSetup() {
  const route = useRoute();
  const { api, isActive, state, router, busy, error, perform, commit, observe, checkOperation, signal } = useConnections();
  const setup = computed(() => (state.value.setup?.id === route.params.id ? state.value.setup : null));
  const postcode = ref('');
  const property = ref('');
  const initialName = 'Bin collections';
  const name = ref(initialName);
  let nameEdited = false;
  watch(
    name,
    () => { nameEdited = true; },
    { flush: 'sync' }
  );
  const token = ref('');
  const definition = ref<ProviderDefinition | null>(null);
  // Only named return intents are accepted; setup never redirects to a supplied URL.
  async function finishSetup(id: string) {
    if (route.query.return_to === 'board') {
      commit('boardReturn', { id, scrollTop: 0 });
      await router.replace({ name: RouteName.Integrations });
    } else await router.replace({ name: RouteName.Integration, params: { id } });
  }
  async function reload(requestSignal = signal) {
    const value = await api.setup(String(route.params.id), requestSignal);
    if (!isActive() || requestSignal.aborted) return value;
    commit('setup', value);
    if (value.instance_id && value.status === 'completed') await finishSetup(value.instance_id);
    return value;
  }
  onMounted(async () => {
    const value = await perform(() => reload());
    if (!value) return;
    postcode.value = value.postcode ?? '';
    const provider = await perform(() => api.definition(value.provider_key, signal));
    if (provider) {
      definition.value = provider.document;
      if (!nameEdited) name.value = setupPresentation(provider.document).defaultName;
    }
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
  /** Submit the postcode search and observe its admitted operation; old address choices are discarded. */
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
  /** Save the selected provider property ID and follow the named board/detail return flow. */
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
    if (value.instance_id) await finishSetup(value.instance_id);
    else if (value.operation_id) observe(value.operation_id, reload);
  }
  /** Continue an admitted setup operation without starting a duplicate search/lookup. */
  async function check() {
    await checkOperation(setup.value?.operation_id, reload);
  }
  /** Save the setup token in server-owned custody and immediately clear its plaintext field. */
  async function saveToken() {
    const current = setup.value;
    if (!current) return;
    const slot = definition.value?.auth.secret_slot;
    if (!slot) return;
    const value = await perform(() => api.credential('setup', current.id, slot, current.etag, token.value));
    token.value = '';
    if (value && 'choices' in value) commit('setup', value);
  }
  /** Confirm deletion of this unfinished setup; already-submitted remote work cannot be recalled. */
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
    if (value) await router.replace({ name: RouteName.Integrations });
  }
  /** Retry the local setup read, preserving explicit manual provider admission. */
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
