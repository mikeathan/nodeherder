import { computed, onMounted, ref } from 'vue';
import { useRoute } from 'vue-router';
import { useConnections } from '@/composables/useConnections';
import { confirm } from '@/composables/useConfirm';
import { RouteName } from '@/types/router';
import type { DestinationReview, ProviderEditorMode } from '@/types/integrations.type';
import { useProviderDraft } from './useProviderDraft';
import { useProviderPreview } from './useProviderPreview';
import { cloneDefinitionValue } from './useEditableModel';

/** Coordinates validated host-definition actions; draft navigation and previews have separate owners. */
export function useProviderEditor() {
  const route = useRoute();
  const scope = useConnections();
  const { api, state, router, busy, error, perform, commit, signal } = scope;
  const key = computed(() => `${route.params.type}/${route.params.provider}`);
  const newFile = computed(() => route.query.new === '1');
  const draft = useProviderDraft(scope, key);
  const { provider, yaml, documentValue, mode, dirty, etag, validation, approve, reassociate, remember, apply } = draft;
  const message = ref('');
  const affectedInstances = computed(
    () =>
      validation.value?.affected_instances.map((item) => ({
        ...item,
        name: state.value.instances.find((instance) => instance.id === item.instance_id)?.name ?? 'Existing integration',
      })) ?? []
  );
  const instances = computed(() => state.value.instances.filter((item) => item.provider_key === key.value));

  const saved = state.value.drafts[key.value];
  const cached = newFile.value ? undefined : state.value.definitions[key.value];
  // Render the already accepted document/draft immediately; revalidation must not flash defaults.
  if (cached) {
    apply(cached);
    if (saved?.dirty) draft.hydrate(saved);
  }

  onMounted(async () => {
    if (cached && saved?.dirty) await draft.restore(saved);
    // One loading interval covers both reads, so controls and feedback never flicker between them.
    await perform(async () => {
      const value = await api.definition(newFile.value ? String(route.query.from) : key.value);
      if (!scope.isActive()) return;
      commit('definition', value);
      if (saved?.dirty && !newFile.value) {
        // Fresh metadata may change warnings, but the draft and its edit token remain pinned.
        provider.value = value;
        if (!cached) draft.hydrate(saved);
      } else {
        apply(value);
      }
      if (newFile.value) {
        const copy = cloneDefinitionValue(value.document);
        copy.id = String(route.params.provider);
        copy.name += ' copy';
        documentValue.value = copy;
        etag.value = '';
        dirty.value = true;
      }
      const items = await api.instances(signal);
      commit('instances', items);
    });
    // Restore keyboard focus after the loading fieldset is enabled. On slower hosts an earlier
    // animation frame would try to focus a disabled input and silently lose the saved target.
    if (saved?.dirty && scope.isActive()) await draft.restore(saved);
  });
  function useShipped() {
    if (!provider.value?.shipped_yaml) return;
    yaml.value = provider.value.shipped_yaml;
    documentValue.value = null;
    mode.value = 'yaml';
    dirty.value = true;
    remember();
  }
  async function keepVersion() {
    if (!provider.value) return;
    const value = await perform(() => api.acknowledgeShipped(key.value, etag.value));
    if (value) provider.value = value;
  }
  async function validate() {
    const result = await perform(() =>
      api.validate(
        key.value,
        mode.value === 'form' && documentValue.value ? { document: documentValue.value } : { yaml: yaml.value }
      )
    );
    if (!result) return;

    draft.acceptValidation(result);
    message.value = 'Configuration is valid. No provider request has been sent.';
    remember();
    return result;
  }
  async function switchMode(next: ProviderEditorMode) {
    if (next === mode.value) return;
    const result = await validate();
    if (result) {
      mode.value = next;
      remember();
    }
  }
  function review(): DestinationReview | undefined {
    const v = validation.value;
    if (!v || !v.previous_origin || v.origin === v.previous_origin) return undefined;
    if (!approve.value) return undefined;
    return {
      origin: v.origin,
      reassociate: v.affected_instances.filter((i) => reassociate.value.includes(i.instance_id)),
    };
  }
  async function save() {
    message.value = '';
    const v = validation.value ?? (await validate());
    if (!v) return;
    if (v.previous_origin && v.origin !== v.previous_origin && !approve.value) {
      message.value = 'Review the destination below, then choose Save changes again.';
      return;
    }
    const value = await perform(() =>
      api.saveDefinition(key.value, v.yaml, etag.value, review(), newFile.value || provider.value?.status === 'missing')
    );
    if (value) {
      await apply(value);
      message.value = 'Saved on the host and applied. Existing collection dates were kept.';
      if (newFile.value) await router.replace({ name: RouteName.IntegrationProvider, params: route.params });
    }
  }
  async function reload() {
    if (
      dirty.value &&
      !(await confirm({
        title: 'Replace unsaved draft?',
        message:
          'Reload reads the provider file on the host and replaces this editor buffer. Your saved integration data is kept.',
        confirmLabel: 'Reload from host',
      }))
    )
      return;
    const current = await perform(() => api.definition(key.value, signal));
    if (!current) return;
    await apply(current);
    const v = await validate();
    if (!v) return;
    if (v.previous_origin && v.origin !== v.previous_origin) {
      dirty.value = true;
      message.value = 'The host file changes the destination. Review it and Save changes to apply it.';
      return;
    }
    const result = await perform(() => api.reload(key.value, current.etag));
    if (result) {
      await apply(result);
      message.value = 'Reloaded from the host.';
    }
  }
  async function restoreMissing() {
    if (
      !provider.value ||
      !(await confirm({
        title: 'Restore last active file?',
        message:
          'This recreates the missing host file from the last active configuration and replaces your unsaved draft. Saved integration data is kept.',
        confirmLabel: 'Restore file',
      }))
    )
      return;
    const value = await perform(() => api.saveDefinition(key.value, provider.value!.yaml, etag.value, undefined, true));
    if (value) {
      await apply(value);
      message.value = 'The last active provider file was restored.';
    }
  }
  async function restore() {
    if (
      !(await confirm({
        title: 'Restore previous provider configuration?',
        message: 'This writes the previous good revision to the host as a new revision. Saved integration data is kept.',
        confirmLabel: 'Restore previous',
      }))
    )
      return;
    const value = await perform(() => api.restore(key.value, etag.value));
    if (value) {
      await apply(value);
      message.value = 'Previous configuration restored.';
    }
  }
  async function copy() {
    const id = `${String(route.params.provider).slice(0, 44)}_${crypto.randomUUID().slice(0, 8)}`;
    const value = await confirm({
      title: 'Copy provider?',
      message: `Create a separate editable provider named ${id}. It will not share integration addresses or credentials.`,
      confirmLabel: 'Create copy',
    });
    if (value && scope.isActive())
      await router.push({
        name: RouteName.IntegrationProvider,
        params: { type: route.params.type, provider: id },
        query: { new: '1', from: key.value },
      });
  }
  async function remove() {
    if (
      !(await confirm({
        title: 'Remove provider definition?',
        message:
          'This removes the shared provider YAML, not a configured integration. Remove integrations and setups using this provider first; saved operations may also need to expire. No provider account is deleted.',
        confirmLabel: 'Remove definition',
        danger: true,
      }))
    )
      return;
    const value = await perform(() => api.deleteDefinition(key.value, etag.value));
    if (value) {
      dirty.value = false;
      commit('forgetDefinition', key.value);
      provider.value = null;
      await router.replace({ name: RouteName.Integrations });
    }
  }

  const preview = useProviderPreview(
    scope,
    key,
    () => (validation.value ? Promise.resolve(validation.value) : validate()),
    review
  );
  return {
    key,
    newFile,
    provider,
    yaml,
    documentValue,
    mode,
    dirty,
    validation,
    approve,
    reassociate,
    message,
    busy,
    error,
    affectedInstances,
    instances,
    rememberFocus: draft.rememberFocus,
    useShipped,
    keepVersion,
    validate,
    switchMode,
    save,
    reload,
    restoreMissing,
    restore,
    copy,
    remove,
    ...preview,
  };
}
