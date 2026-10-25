<script setup lang="ts">
  import { computed, nextTick, onBeforeUnmount, ref, watch } from 'vue';
  import { onBeforeRouteLeave, useRoute } from 'vue-router';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import ProviderForm from './ProviderForm.vue';
  import ConnectionFeedback from './ConnectionFeedback.vue';
  import { useConnections } from '@/composables/useConnections';
  import { connectionsApi } from '@/services/integrations.service';
  import { providerChanges } from '@/domain/integrations';
  import { confirm } from '@/composables/useConfirm';
  import { RouteName } from '@/types/router';
  import type {
    ConnectionDefinition,
    DefinitionValidation,
    DestinationReview,
    ProviderDefinition,
    ProviderDraft,
    ConnectionSetup,
  } from '@/types/integrations.type';
  import './connections.css';
  const route = useRoute();
  const { state, router, busy, error, perform, commit, signal, observe } = useConnections();
  const key = computed(() => `${route.params.type}/${route.params.provider}`);
  const provider = ref<ConnectionDefinition | null>(null);
  const yaml = ref('');
  const documentValue = ref<ProviderDefinition | null>(null);
  const mode = ref<'form' | 'yaml'>('form');
  const dirty = ref(false);
  const etag = ref('');
  const message = ref('');
  const validation = ref<DefinitionValidation | null>(null);
  const approve = ref(false);
  const reassociate = ref<string[]>([]);
  const ready = ref(false);
  const testInstance = ref('');
  const sample = ref('');
  const preview = ref<ConnectionSetup | null>(null);
  const focusedField = ref('');
  function rememberFocus(event: FocusEvent) {
    const target = event.target;
    if (target instanceof HTMLElement && ['INPUT', 'TEXTAREA', 'SELECT'].includes(target.tagName) && target.id)
      focusedField.value = target.id;
  }
  const newFile = computed(() => route.query.new === '1');
  const storeLoad = () =>
    connectionsApi.instances(signal).then((items) => {
      commit('instances', items);
      return items;
    });
  const helpRoute = computed(() => ({ name: RouteName.ConnectionHelp, params: route.params, query: route.query }));
  const scroller = () => window.document.querySelector<HTMLElement>('.nh-app-main');
  function remember() {
    if (!ready.value) return;
    const value: ProviderDraft = {
      yaml: yaml.value,
      etag: etag.value,
      mode: mode.value,
      dirty: dirty.value,
      scroll: scroller()?.scrollTop ?? 0,
      focus: focusedField.value,
      document: documentValue.value
        ? (JSON.parse(JSON.stringify(documentValue.value)) as ProviderDefinition)
        : undefined,
    };
    commit('draft', { key: key.value, value });
  }
  watch(
    [yaml, documentValue],
    () => {
      if (ready.value) {
        dirty.value = true;
        validation.value = null;
        approve.value = false;
        remember();
      }
    },
    { deep: true, flush: 'sync' }
  );
  onBeforeRouteLeave(async (to) => {
    remember();
    if (
      dirty.value &&
      to.name !== RouteName.ConnectionHelp &&
      !String(to.path).startsWith(`/connections/providers/${route.params.type}/${route.params.provider}`)
    )
      return confirm({
        title: 'Leave provider editor?',
        message: 'Your draft is kept in memory while you stay signed in. It has not been saved to the host.',
        confirmLabel: 'Keep draft and leave',
      });
    return true;
  });
  onBeforeUnmount(remember);
  async function apply(value: ConnectionDefinition) {
    ready.value = false;
    provider.value = value;
    yaml.value = value.yaml;
    etag.value = value.etag;
    documentValue.value = value.status === 'active' ? value.document : null;
    mode.value = documentValue.value ? 'form' : 'yaml';
    dirty.value = false;
    validation.value = null;
    await nextTick();
    ready.value = true;
    remember();
  }
  watch(
    () => route.fullPath,
    async () => {
      ready.value = false;
      validation.value = null;
      message.value = '';
      const saved = state.value.drafts[key.value];
      const value = await perform(() =>
        connectionsApi.definition(newFile.value ? String(route.query.from) : key.value, signal)
      );
      if (value) {
        if (newFile.value) {
          const copy = JSON.parse(JSON.stringify(value.document)) as ProviderDefinition;
          copy.id = String(route.params.provider);
          copy.name += ' copy';
          documentValue.value = copy;
          etag.value = '';
          yaml.value = value.yaml;
          provider.value = value;
          dirty.value = true;
        } else await apply(value);
      }
      if (saved?.dirty) {
        ready.value = false;
        yaml.value = saved.yaml;
        etag.value = saved.etag;
        mode.value = saved.mode;
        documentValue.value = saved.document
          ? (JSON.parse(JSON.stringify(saved.document)) as ProviderDefinition)
          : null;
        dirty.value = true;
        await nextTick();
        ready.value = true;
        requestAnimationFrame(() => {
          const field = saved.focus ? window.document.getElementById(saved.focus) : null;
          let parent = field?.parentElement;
          while (parent) {
            if (parent.tagName === 'DETAILS') (parent as HTMLDetailsElement).open = true;
            parent = parent.parentElement;
          }
          field?.focus({ preventScroll: true });
          scroller()?.scrollTo({ top: saved.scroll });
        });
      }
      ready.value = true;
      await perform(() => storeLoad());
    },
    { immediate: true }
  );
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
    const value = await perform(() => connectionsApi.acknowledgeShipped(key.value, etag.value));
    if (value) provider.value = value;
  }
  async function validate() {
    const result = await perform(() =>
      connectionsApi.validate(
        key.value,
        mode.value === 'form' && documentValue.value ? { document: documentValue.value } : { yaml: yaml.value }
      )
    );
    if (!result) return;
    ready.value = false;
    yaml.value = result.yaml;
    documentValue.value = result.document;
    ready.value = true;
    validation.value = result;
    message.value = 'Configuration is valid. No provider request has been sent.';
    remember();
    return result;
  }
  async function switchMode(next: 'form' | 'yaml') {
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
      connectionsApi.saveDefinition(
        key.value,
        v.yaml,
        etag.value,
        review(),
        newFile.value || provider.value?.status === 'missing'
      )
    );
    if (value) {
      await apply(value);
      message.value = 'Saved on the host and applied. Existing collection dates were kept.';
      if (newFile.value) await router.replace({ name: RouteName.ConnectionProvider, params: route.params });
    }
  }
  async function reload() {
    if (
      dirty.value &&
      !(await confirm({
        title: 'Replace unsaved draft?',
        message:
          'Reload reads the provider file on the host and replaces this editor buffer. Your saved connection data is kept.',
        confirmLabel: 'Reload from host',
      }))
    )
      return;
    const current = await perform(() => connectionsApi.definition(key.value, signal));
    if (!current) return;
    await apply(current);
    const v = await validate();
    if (!v) return;
    if (v.previous_origin && v.origin !== v.previous_origin) {
      dirty.value = true;
      message.value = 'The host file changes the destination. Review it and Save changes to apply it.';
      return;
    }
    const result = await perform(() => connectionsApi.reload(key.value, current.etag));
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
          'This recreates the missing host file from the last active configuration and replaces your unsaved draft. Saved connection data is kept.',
        confirmLabel: 'Restore file',
      }))
    )
      return;
    const value = await perform(() =>
      connectionsApi.saveDefinition(key.value, provider.value!.yaml, etag.value, undefined, true)
    );
    if (value) {
      await apply(value);
      message.value = 'The last active provider file was restored.';
    }
  }
  async function restore() {
    if (
      !(await confirm({
        title: 'Restore previous provider configuration?',
        message: 'This writes the previous good revision to the host as a new revision. Saved connection data is kept.',
        confirmLabel: 'Restore previous',
      }))
    )
      return;
    const value = await perform(() => connectionsApi.restore(key.value, etag.value));
    if (value) {
      await apply(value);
      message.value = 'Previous configuration restored.';
    }
  }
  async function copy() {
    const id = `${String(route.params.provider).slice(0, 44)}_${crypto.randomUUID().slice(0, 8)}`;
    const value = await confirm({
      title: 'Copy provider?',
      message: `Create a separate editable provider named ${id}. It will not share connection addresses or credentials.`,
      confirmLabel: 'Create copy',
    });
    if (value)
      await router.push({
        name: RouteName.ConnectionProvider,
        params: { type: route.params.type, provider: id },
        query: { new: '1', from: key.value },
      });
  }
  async function test(mode: 'offline' | 'live') {
    const v = validation.value ?? (await validate());
    if (!v) return;
    let input: unknown;
    if (mode === 'offline') {
      try {
        input = JSON.parse(sample.value);
      } catch {
        error.value = { code: 'invalid_input', message: 'Enter a valid JSON sample.' };
        return;
      }
    }
    if (
      mode === 'live' &&
      !(await confirm({
        title: 'Send a test request?',
        message:
          'This sends a saved-schedule request using the selected connection’s private inputs and token. The preview does not replace its saved results.',
        confirmLabel: 'Send test request',
      }))
    )
      return;
    const result = await perform(() =>
      connectionsApi.test(key.value, v.yaml, testInstance.value, mode, input, review())
    );
    if (result) {
      preview.value = result;
      if (result.operation_id)
        observe(result.operation_id, async () => {
          const updated = await connectionsApi.setup(result.id, signal);
          preview.value = updated;
          return updated;
        });
    }
  }

  async function remove() {
    if (
      !(await confirm({
        title: 'Remove provider definition?',
        message:
          'This removes the host YAML file. Connections and setups using it must be removed first. No provider account is deleted.',
        confirmLabel: 'Remove definition',
        danger: true,
      }))
    )
      return;
    const value = await perform(() => connectionsApi.deleteDefinition(key.value, etag.value));
    if (value) {
      dirty.value = false;
      await router.replace({ name: RouteName.Connections });
    }
  }
</script>
<template>
  <div class="connections connection-stack" @focusin="rememberFocus">
    <UiPageHeader
      title="Configure operations"
      :subtitle="provider?.document.name"
      :back="{ name: RouteName.Connections }"
      back-label="Connections"
      ><template #actions><UiButton :to="helpRoute" icon="unknown">Help</UiButton></template></UiPageHeader
    >
    <ConnectionFeedback :error="error" :busy="busy" />
    <div v-if="message" class="connection-notice" role="status">{{ message }}</div>
    <div v-if="provider?.status === 'host_changed'" class="connection-notice" role="status">
      The host file differs from the active configuration. Review it and choose Reload from host to apply it.
    </div>
    <div v-if="provider?.disk_error" class="connection-notice" role="status">{{ provider.disk_error }}</div>
    <UiCard v-if="provider" title="Provider configuration"
      ><div class="connection-stack">
        <p class="connection-muted">
          Changes affect every connection using this provider. Saved on the host as configs/integrations/{{ key }}.yaml.
          Address choices, tokens and results stay separate.
        </p>
        <details v-if="provider.shipped_yaml && !newFile">
          <summary>
            {{ provider.update_available ? 'Updated supplied example available' : 'Compare with the supplied example' }}
          </summary>
          <p class="connection-muted">
            App updates keep your host file. Review the differences before using the supplied example; it replaces only
            the draft until you Save changes.
          </p>
          <pre class="connection-comparison">{{ providerChanges(provider.yaml, provider.shipped_yaml) }}</pre>
          <div class="connection-actions">
            <UiButton :disabled="busy" @click="useShipped">Use supplied example in editor</UiButton
            ><UiButton v-if="provider.update_available" :disabled="busy" @click="keepVersion">Keep my version</UiButton>
          </div>
        </details>
        <div class="connection-actions">
          <UiButton :pressed="mode === 'form'" :disabled="busy" @click="switchMode('form')">Form</UiButton
          ><UiButton :pressed="mode === 'yaml'" :disabled="busy" @click="switchMode('yaml')">YAML</UiButton
          ><span class="connection-muted" role="status">{{ dirty ? 'Unsaved changes' : 'Saved' }}</span>
        </div>
        <fieldset :disabled="busy" class="connection-form-fields">
          <legend>Provider fields</legend>
          <ProviderForm v-if="mode === 'form' && documentValue" v-model="documentValue" />
          <label v-else class="connection-field"
            ><span>Provider YAML</span
            ><textarea
              id="provider-yaml"
              v-model="yaml"
              class="connection-code"
              spellcheck="false"
              aria-describedby="yaml-note"
              maxlength="65536" />
          </label>
        </fieldset>
        <p id="yaml-note" class="connection-muted">
          Form edits may normalize YAML formatting. Fixed values and comments are part of the shareable file; review
          them before committing it to Git. Save does not run Git commands.
        </p>
        <fieldset
          v-if="validation?.previous_origin && validation.origin !== validation.previous_origin"
          class="connection-form">
          <legend>Provider destination changed</legend>
          <p class="connection-break">From {{ validation.previous_origin }} to {{ validation.origin }}.</p>
          <label><input v-model="approve" type="checkbox" /> I approve this destination for future requests.</label>
          <p class="connection-muted">
            Choose which existing connections may send their saved inputs and token to it. Others will need
            reconnecting.
          </p>
          <label v-for="item in validation.affected_instances" :key="item.instance_id"
            ><input v-model="reassociate" type="checkbox" :value="item.instance_id" />
            {{ state.instances.find((i) => i.id === item.instance_id)?.name ?? 'Existing connection' }}
            {{ item.credential_revision ? '— includes its saved token' : '— saved inputs only' }}</label
          >
        </fieldset>
        <details v-if="!newFile">
          <summary>Test without changing saved results</summary>
          <div class="connection-form">
            <label class="connection-field"
              ><span>Use inputs from this connection</span
              ><select v-model="testInstance">
                <option value="">Choose a saved connection</option>
                <option
                  v-for="item in state.instances.filter((i) => i.provider_key === key)"
                  :key="item.id"
                  :value="item.id">
                  {{ item.name }}
                </option>
              </select></label
            >
            <p class="connection-muted">
              Offline test checks a supplied schedule JSON sample. Live test sends only the saved-schedule GET request
              with private inputs; it does not start a new lookup or replace saved collection dates.
            </p>
            <label class="connection-field"
              ><span>Schedule JSON sample for an offline test</span
              ><textarea v-model="sample" spellcheck="false" rows="6" maxlength="262144" />
            </label>
            <div class="connection-actions">
              <UiButton :disabled="busy || !testInstance || !sample" @click="test('offline')">Test sample</UiButton
              ><UiButton :disabled="busy || !testInstance" @click="test('live')">Test saved-schedule request</UiButton>
            </div>
            <p v-if="preview" role="status">
              {{
                preview.last_error?.message ??
                (preview.preview ? 'Test complete. Saved results were not changed.' : 'Test in progress…')
              }}
            </p>
            <ul v-if="preview?.preview" class="connection-list">
              <li v-for="service in preview.preview.services" :key="service.service_id">
                {{ service.name }}: {{ service.dates.join(', ') || 'dates unavailable' }}
              </li>
            </ul>
          </div>
        </details>
        <div class="connection-actions">
          <UiButton :disabled="busy" @click="validate">Validate</UiButton
          ><UiButton variant="primary" :disabled="busy" @click="save">Save changes</UiButton
          ><UiButton v-if="!newFile" :disabled="busy" @click="reload">Reload from host</UiButton
          ><UiButton
            v-if="provider.status === 'missing' && provider.active_hash"
            :disabled="busy"
            @click="restoreMissing"
            >Restore last active file</UiButton
          >
          <UiButton v-if="!newFile && provider.status !== 'missing'" :disabled="busy" @click="restore"
            >Restore previous</UiButton
          >
        </div>
        <div v-if="!newFile" class="connection-actions">
          <UiButton :disabled="busy || provider.status !== 'active'" @click="copy">Copy provider</UiButton
          ><UiButton variant="ghost" :disabled="busy" @click="remove">Remove definition</UiButton>
        </div>
      </div></UiCard
    >
  </div>
</template>
