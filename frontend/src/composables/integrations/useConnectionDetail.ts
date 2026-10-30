import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useConnections } from '@/composables/useConnections';
import { confirm } from '@/composables/useConfirm';
import { RouteName } from '@/types/router';
import { ConnectionError } from '@/services/integrations.service';
import { matchesScheduleEditor, sameScheduleConfig, scheduleEditor, toIntervalSeconds, type ScheduleUnit } from '@/domain/schedule';
import type { ConnectionInstance, ProviderDefinition } from '@/types/integrations.type';
import { integrationDetailSection, type IntegrationNavigationContext } from '@/domain/integration-navigation';

/** Owns this page's application actions; the route component only presents its state. */
export function useConnectionDetail() {
  const route = useRoute();
  const { api, isActive, state, store, router, busy, error, perform, observe, checkOperation, signal, commit } = useConnections();
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
  const initialSchedule = item.value?.schedule ?? null;
  const initialEditor = scheduleEditor(initialSchedule);
  const scheduleEnabled = ref(initialEditor.enabled);
  const scheduleValue = ref(initialEditor.value);
  const scheduleUnit = ref<ScheduleUnit>(initialEditor.unit);
  let confirmedSchedule = initialSchedule;
  let scheduleBaseConfig = initialSchedule;
  const scheduleBaseETag = ref(item.value?.etag ?? '');
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
        commit('forgetInstance', instanceId.value);
      }
      throw cause;
    }
  }
  /** Retry the private instance read without submitting a fresh provider lookup. */
  const reloadPage = () => perform(() => reloadInstance(signal, true));
  /** Copy confirmed schedule state into the editor. */
  const applySchedule = (schedule: ConnectionInstance['schedule']): void => {
    const editor = scheduleEditor(schedule);
    scheduleEnabled.value = editor.enabled;
    scheduleValue.value = editor.value;
    scheduleUnit.value = editor.unit;
  };
  /**
   * Apply a fresh confirmed instance without replacing unsaved editor values. The schedule draft
   * keeps the ETag of the confirmed schedule it was based on; a real schedule change in another
   * session therefore produces a revision conflict instead of a silent overwrite. Unrelated
   * instance revisions may advance the base while the confirmed cadence is unchanged.
   */
  const acceptInstance = (current: ConnectionInstance, forceSchedule = false): void => {
    if (name.value === initialName) name.value = current.name;
    const pristine = matchesScheduleEditor(
      confirmedSchedule,
      scheduleEnabled.value,
      scheduleValue.value,
      scheduleUnit.value
    );
    const sameConfig = sameScheduleConfig(scheduleBaseConfig, current.schedule);
    confirmedSchedule = current.schedule;
    if (pristine || forceSchedule) {
      applySchedule(current.schedule);
      scheduleBaseConfig = current.schedule;
      scheduleBaseETag.value = current.etag;
      return;
    }
    if (sameConfig || !scheduleBaseETag.value) {
      scheduleBaseETag.value = current.etag;
    }
  };
  /**
   * Reload the private instance and reconcile the editor after every successful read. A caller
   * reload explicitly discards an unsaved schedule draft; background reads preserve it.
   */
  async function reloadInstance(requestSignal = signal, forceSchedule = false): Promise<ConnectionInstance> {
    const current = await reload(requestSignal);
    if (isActive() && !requestSignal.aborted) acceptInstance(current, forceSchedule);
    return current;
  }
  let projection: ReturnType<typeof setInterval> | undefined;
  const onFocus = () => {
    if (isActive() && !busy.value && !document.hidden) void perform(() => reloadInstance());
  };
  onMounted(async () => {
    projection = setInterval(onFocus, 60_000);
    window.addEventListener('focus', onFocus);
    // One owned loading interval covers both reads; a failed instance read must not be
    // followed by a provider read that clears its error or uses a stale binding.
    await perform(async () => {
      const current = await reloadInstance();
      if (!isActive()) return;
      const provider = await api.definition(current.provider_key, signal);
      if (isActive()) {
        definition.value = provider.document;
        commit('definition', provider);
      }
    });
    if (isActive() && item.value?.operation_id) observe(item.value.operation_id, reloadInstance);
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
    if (operation) observe(operation.id, reloadInstance);
  }
  /** Continue the current admitted operation without creating a second lookup. */
  async function check() {
    await checkOperation(item.value?.operation_id, reloadInstance);
  }
  /** Save name/enabled state with revision fencing, then reload the accepted resource view. */
  async function update(enabled: boolean) {
    const current = item.value;
    if (!current) return;
    const result = await perform(() => api.update(current.id, current.etag, name.value || current.name, enabled));
    if (result) await perform(() => reloadInstance());
  }
  /** Save the automatic-refresh choice; the server validates it and starts no provider request. */
  async function saveSchedule() {
    const current = item.value;
    if (!current) return;
    const seconds = toIntervalSeconds(scheduleValue.value, scheduleUnit.value);
    const result = await perform(() =>
      api.updateSchedule(current.id, scheduleBaseETag.value || current.etag, scheduleEnabled.value, seconds)
    );
    if (result) {
      confirmedSchedule = result.schedule;
      scheduleBaseConfig = result.schedule;
      scheduleBaseETag.value = result.etag;
      await perform(() => reloadInstance());
    }
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
          'Future requests will use anonymous access if the provider allows it. Saved integration results remain available.',
        confirmLabel: 'Remove token',
        danger: true,
      }))
    )
      return;
    const result = await perform(() =>
      api.credential('instances', current.id, slot, current.etag, token.value, remove)
    );
    token.value = '';
    if (result) await perform(() => reloadInstance());
  }
  /** Confirm destructive deletion, clear the current private DTO and return to the board. */
  async function remove() {
    const current = item.value;
    if (!current) return;
    if (
      !(await confirm({
        title: 'Remove integration?',
        message:
          'This removes this integration’s saved inputs, credentials and results from NodeHerder. Its Home cards will show as unavailable until you remove them. The shared provider definition stays available. A request already sent to the provider cannot be recalled.',
        confirmLabel: 'Remove integration',
        danger: true,
      }))
    )
      return;
    const result = await perform(() => api.remove(current.id, current.etag));
    if (result) {
      commit('forgetInstance', current.id);
      await router.replace({ name: RouteName.Integrations });
    }
  }
  return {
    item,
    tab,
    name,
    token,
    definition,
    scheduleEnabled,
    scheduleValue,
    scheduleUnit,
    sectionContext,
    busy,
    error,
    refresh,
    check,
    update,
    saveSchedule,
    editAddress,
    credential,
    remove,
    reloadPage,
  };
}
