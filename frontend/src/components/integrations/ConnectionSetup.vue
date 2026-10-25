<script setup lang="ts">
  import { computed, onMounted, onBeforeUnmount, ref } from 'vue';
  import { useRoute } from 'vue-router';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import ConnectionFeedback from './ConnectionFeedback.vue';
  import { useConnections } from '@/composables/useConnections';
  import { connectionsApi } from '@/services/integrations.service';
  import { confirm } from '@/composables/useConfirm';
  import { RouteName } from '@/types/router';
  import type { ProviderDefinition } from '@/types/integrations.type';
  import './connections.css';
  const route = useRoute();
  const { state, router, busy, error, perform, commit, observe, signal } = useConnections();
  const setup = computed(() => (state.value.setup?.id === route.params.id ? state.value.setup : null));
  const postcode = ref('');
  const property = ref('');
  const name = ref('Bin collections');
  const token = ref('');
  const definition = ref<ProviderDefinition | null>(null);
  async function reload() {
    const value = await connectionsApi.setup(String(route.params.id), signal);
    commit('setup', value);
    if (value.instance_id && value.status === 'completed')
      await router.replace({ name: RouteName.Connection, params: { id: value.instance_id } });
    return value;
  }
  onMounted(async () => {
    const value = await perform(reload);
    if (!value) return;
    postcode.value = value.postcode ?? '';
    const provider = await perform(() => connectionsApi.definition(value.provider_key, signal));
    if (provider) definition.value = provider.document;
    if (value.operation_id && value.status === 'searching') observe(value.operation_id, reload);
  });
  onBeforeUnmount(() => {
    token.value = '';
  });
  async function search() {
    if (!setup.value) return;
    property.value = '';
    const value = await perform(() =>
      connectionsApi.step(setup.value!.id, setup.value!.etag, { action: 'search', postcode: postcode.value })
    );
    if (value) {
      commit('setup', value);
      if (value.operation_id) observe(value.operation_id, reload);
    }
  }
  async function save() {
    if (!setup.value) return;
    const value = await perform(() =>
      connectionsApi.step(setup.value!.id, setup.value!.etag, {
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
    if (!setup.value?.operation_id) return;
    const value = await perform(() => connectionsApi.check(setup.value!.operation_id!, setup.value!.etag));
    if (value) observe(value.id, reload);
  }
  async function saveToken() {
    if (!setup.value || !definition.value?.auth.secret_slot) return;
    const value = await perform(() =>
      connectionsApi.credential(
        'setup',
        setup.value!.id,
        definition.value!.auth.secret_slot!,
        setup.value!.etag,
        token.value
      )
    );
    token.value = '';
    if (value && 'choices' in value) commit('setup', value);
  }
  async function cancel() {
    if (
      !setup.value ||
      !(await confirm({
        title: 'Discard setup?',
        message:
          'This removes this unfinished setup and its saved inputs. A lookup already submitted to the provider cannot be cancelled.',
        confirmLabel: 'Discard setup',
        danger: true,
      }))
    )
      return;
    const value = await perform(() => connectionsApi.cancel(setup.value!.id, setup.value!.etag));
    if (value) await router.replace({ name: RouteName.Connections });
  }
</script>
<template>
  <div class="connections connection-stack">
    <UiPageHeader title="Set up bin collections" :back="{ name: RouteName.Connections }" back-label="Connections" />
    <ConnectionFeedback :error="error ?? setup?.last_error" :busy="busy" />
    <UiCard v-if="setup" title="Your address"
      ><div class="connection-form">
        <p class="connection-muted">
          The provider receives your postcode and selected address to retrieve collection dates. Refreshes are manual in
          this release.
        </p>
        <details v-if="definition?.auth.mode !== 'none' && definition?.auth.secret_slot">
          <summary>
            Provider credential
            {{
              setup.credential_configured ? '(saved)' : definition?.auth.mode === 'bearer' ? '(required)' : '(optional)'
            }}
          </summary>
          <p class="connection-muted">
            Use a token issued by the provider if your access requires one. It is stored privately and never included in
            the provider YAML.
          </p>
          <label class="connection-field"
            ><span>API token</span><input v-model="token" type="password" autocomplete="off" maxlength="8192" /></label
          ><UiButton :disabled="!token || busy" @click="saveToken">Save token</UiButton>
        </details>
        <form class="connection-form" @submit.prevent="search">
          <label class="connection-field"
            ><span>Postcode</span><input v-model="postcode" autocomplete="postal-code" maxlength="10" required
          /></label>
          <div><UiButton type="submit" :disabled="busy || !postcode.trim()">Find address</UiButton></div>
        </form>
        <form v-if="setup.status === 'choose'" class="connection-form" @submit.prevent="save">
          <label class="connection-field"
            ><span>Choose your address</span
            ><select v-model="property" required>
              <option disabled value="">Select an address</option>
              <option v-for="choice in setup.choices" :key="choice.id" :value="choice.id">{{ choice.label }}</option>
            </select></label
          >
          <p v-if="!setup.choices.length" class="connection-muted">
            No addresses were found. Check the postcode or try another provider.
          </p>
          <label class="connection-field"
            ><span>Connection name</span><input v-model="name" maxlength="120" required
          /></label>
          <div>
            <UiButton type="submit" variant="primary" :disabled="busy || !property || !name.trim()"
              >Save connection</UiButton
            >
          </div>
        </form>
        <div v-if="['searching', 'saving', 'pending', 'error'].includes(setup.status) && setup.operation_id">
          <p class="connection-muted">
            {{
              setup.status === 'searching'
                ? 'Finding addresses…'
                : 'The provider has not finished. Use Check status to continue the same lookup.'
            }}
          </p>
          <UiButton :disabled="busy" @click="check">Check status</UiButton>
        </div>
        <UiButton v-if="setup.instance_id" :to="{ name: RouteName.Connection, params: { id: setup.instance_id } }"
          >Open saved connection</UiButton
        >
        <div><UiButton variant="ghost" :disabled="busy" @click="cancel">Discard setup</UiButton></div>
      </div></UiCard
    >
    <UiButton v-if="error" @click="perform(reload)">Reload setup</UiButton>
  </div>
</template>
