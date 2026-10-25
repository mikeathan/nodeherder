<script setup lang="ts">
  import { computed, onBeforeUnmount, onMounted, ref } from 'vue';
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
  const { state, store, router, busy, error, perform, observe, signal } = useConnections();
  const item = computed(() => (state.value.current?.id === route.params.id ? state.value.current : null));
  const tab = ref('results');
  const name = ref('');
  const token = ref('');
  const definition = ref<ProviderDefinition | null>(null);
  const providerParams = computed(() => {
    const [type, provider] = (item.value?.provider_key ?? '/').split('/');
    return { type, provider };
  });
  const reload = () => store.dispatch('integrations/loadInstance', { id: String(route.params.id), signal });
  const today = () =>
    new Intl.DateTimeFormat('en-CA', {
      timeZone: 'Europe/London',
      year: 'numeric',
      month: '2-digit',
      day: '2-digit',
    }).format(new Date());
  const upcoming = (dates: string[]) => dates.filter((d) => d >= today());
  const dateLabel = (date: string) =>
    new Intl.DateTimeFormat('en-GB', { weekday: 'long', day: 'numeric', month: 'long', timeZone: 'UTC' }).format(
      new Date(`${date}T12:00:00Z`)
    );
  let projection: ReturnType<typeof setInterval> | undefined;
  const onFocus = () => {
    if (!busy.value) void perform(reload);
  };
  onMounted(async () => {
    projection = setInterval(onFocus, 60_000);
    window.addEventListener('focus', onFocus);
    await perform(reload);
    if (item.value) {
      name.value = item.value.name;
      const provider = await perform(() => connectionsApi.definition(item.value!.provider_key, signal));
      if (provider) definition.value = provider.document;
      if (item.value.operation_id) observe(item.value.operation_id, reload);
    }
  });
  onBeforeUnmount(() => {
    if (projection) clearInterval(projection);
    window.removeEventListener('focus', onFocus);
    token.value = '';
  });
  async function refresh() {
    if (!item.value) return;
    const operation = await perform(() => connectionsApi.refresh(item.value!.id, item.value!.etag));
    if (operation) observe(operation.id, reload);
  }
  async function check() {
    if (!item.value?.operation_id) return;
    const operation = await perform(() => connectionsApi.check(item.value!.operation_id!, item.value!.etag));
    if (operation) observe(operation.id, reload);
  }
  async function update(enabled: boolean) {
    if (!item.value) return;
    const result = await perform(() =>
      connectionsApi.update(item.value!.id, item.value!.etag, name.value || item.value!.name, enabled)
    );
    if (result) await perform(reload);
  }
  async function editAddress() {
    if (!item.value) return;
    const p = await perform(() =>
      connectionsApi.createSetup(item.value!.provider_key, { id: item.value!.id, etag: item.value!.etag })
    );
    if (p) await router.push({ name: RouteName.ConnectionSetup, params: { id: p.id } });
  }
  async function credential(remove = false, savedSlot?: string) {
    if (!item.value || (!savedSlot && !definition.value?.auth.secret_slot)) return;
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
      connectionsApi.credential(
        'instances',
        item.value!.id,
        savedSlot ?? definition.value!.auth.secret_slot!,
        item.value!.etag,
        token.value,
        remove
      )
    );
    token.value = '';
    if (result) await perform(reload);
  }
  async function remove() {
    if (
      !item.value ||
      !(await confirm({
        title: 'Remove connection?',
        message:
          'This deletes the saved address, credentials and collection data from NodeHerder. The shared provider configuration is kept. A lookup already submitted to the provider cannot be recalled.',
        confirmLabel: 'Remove connection',
        danger: true,
      }))
    )
      return;
    const result = await perform(() => connectionsApi.remove(item.value!.id, item.value!.etag));
    if (result) await router.replace({ name: RouteName.Connections });
  }
</script>
<template>
  <div class="connections connection-stack">
    <UiPageHeader
      :title="item?.name ?? 'Connection'"
      :subtitle="item?.postcode"
      :back="{ name: RouteName.Connections }"
      back-label="Connections"
      ><template #actions
        ><UiButton :disabled="busy || !item?.enabled" icon="refresh" @click="refresh">Refresh</UiButton></template
      ></UiPageHeader
    >
    <ConnectionFeedback :error="error ?? item?.last_error" :busy="busy" />
    <UiButton
      v-if="
        item &&
        [
          'no_collection_dates',
          'lookup_failed',
          'reconfigure_needed',
          'operation_expired',
          'restart_setup',
          'reconciliation_required',
        ].includes(error?.code ?? item.last_error?.code ?? '')
      "
      :disabled="busy"
      @click="editAddress"
      >Repeat address setup</UiButton
    >
    <template v-if="item">
      <nav class="connection-tabs" aria-label="Connection sections">
        <a href="#results" :aria-current="tab === 'results' ? 'page' : undefined" @click.prevent="tab = 'results'"
          >Results</a
        ><a href="#setup" :aria-current="tab === 'setup' ? 'page' : undefined" @click.prevent="tab = 'setup'">Setup</a
        ><RouterLink :to="{ name: RouteName.ConnectionProvider, params: providerParams }"
          >Configure operations</RouterLink
        ><RouterLink :to="{ name: RouteName.ConnectionHelp, params: providerParams }">Help</RouterLink>
      </nav>
      <div v-if="!item.enabled" class="connection-notice" role="status">
        This connection is disabled. Saved dates remain visible.
      </div>
      <div v-if="item.operation_id" class="connection-notice" role="status">
        <p>A lookup is in progress. Check status continues that lookup without submitting a new one.</p>
        <UiButton :disabled="busy || !item.enabled" @click="check">Check status</UiButton>
      </div>
      <details v-if="item.runs.length">
        <summary>Recent checks</summary>
        <ol class="connection-list">
          <li v-for="run in [...item.runs].reverse()" :key="run.operation_id">
            {{ new Date(run.at).toLocaleString() }} — {{ run.message }}
          </li>
        </ol>
      </details>
      <template v-if="tab === 'results'">
        <p v-if="!item.resources.length" class="connection-muted">
          No collection dates have been saved yet. Check the pending lookup or use Refresh.
        </p>
        <section v-for="resource in item.resources" :key="resource.resource_id" class="connection-stack">
          <div class="connection-notice">
            <p>
              {{
                resource.freshness === 'fresh'
                  ? 'Saved collection schedule'
                  : resource.freshness === 'stale'
                    ? 'The saved schedule needs refreshing.'
                    : 'The provider has not supplied a freshness time.'
              }}
            </p>
            <p class="connection-muted">
              Last checked: {{ new Date(resource.last_checked_at).toLocaleString() }}. Provider verification:
              {{
                resource.source_retrieved_at ? new Date(resource.source_retrieved_at).toLocaleString() : 'not supplied'
              }}.
            </p>
            <p class="connection-muted">
              Refresh is manual. Dates use the UK calendar; limited coverage does not mean later collections are
              cancelled.
            </p>
            <p v-for="notice in resource.notices" :key="notice">{{ notice }}</p>
          </div>
          <div class="connection-grid">
            <UiCard v-for="service in resource.services" :key="service.service_id" :title="service.name" icon="schedule"
              ><p class="connection-date">
                {{
                  upcoming(service.dates)[0]
                    ? dateLabel(upcoming(service.dates)[0])
                    : service.date_state === 'unavailable'
                      ? 'Dates unavailable'
                      : 'No upcoming date supplied'
                }}
              </p>
              <p v-if="service.bin_colour" class="connection-muted">{{ service.bin_colour }} bin</p>
              <p v-if="service.subscription === 'required'" class="connection-muted">
                This service requires a subscription. Check that your address is subscribed.
              </p>
              <ul v-if="upcoming(service.dates).length > 1" class="connection-list">
                <li v-for="date in upcoming(service.dates).slice(1, 5)" :key="date">{{ dateLabel(date) }}</li>
              </ul>
              <p class="connection-muted">
                {{
                  service.coverage.state === 'limited'
                    ? 'The provider has supplied a limited date range.'
                    : service.coverage.state === 'next_only'
                      ? 'Only the next collection is confirmed.'
                      : 'The provider has not confirmed the full coverage period.'
                }}
              </p></UiCard
            >
          </div>
        </section>
      </template>
      <UiCard v-else title="Connection settings"
        ><div class="connection-form">
          <label class="connection-field"><span>Connection name</span><input v-model="name" maxlength="120" /></label>
          <div class="connection-actions">
            <UiButton :disabled="busy || !name.trim()" @click="update(item.enabled)">Save name</UiButton
            ><UiButton :disabled="busy" @click="update(!item.enabled)">{{
              item.enabled ? 'Disable connection' : 'Enable connection'
            }}</UiButton
            ><UiButton :disabled="busy" @click="editAddress">Change address</UiButton>
          </div>
          <p class="connection-muted">
            Changing the address uses a separate setup. Your saved dates remain available until the new address is
            verified.
          </p>
          <fieldset v-if="definition?.auth.secret_slot || item.credential_slots.length">
            <legend>Provider token</legend>
            <p class="connection-muted">
              {{
                item.credential_configured ? 'A token is saved. Its value cannot be displayed.' : 'No token is saved.'
              }}
            </p>
            <p class="connection-muted">
              Saved tokens for other slot names are not used by the current configuration. Removing a token stops any
              pending attempt.
            </p>
            <label v-if="definition?.auth.mode !== 'none' && definition?.auth.secret_slot" class="connection-field"
              ><span>{{
                item.credential_slots.includes(definition.auth.secret_slot) ? 'Replacement API token' : 'API token'
              }}</span
              ><input v-model="token" type="password" autocomplete="off" maxlength="8192"
            /></label>
            <div class="connection-actions">
              <UiButton
                v-if="definition?.auth.mode !== 'none' && definition?.auth.secret_slot"
                :disabled="busy || !token"
                @click="credential()"
                >Save token</UiButton
              ><UiButton
                v-for="slot in item.credential_slots"
                :key="slot"
                :disabled="busy"
                @click="credential(true, slot)"
                >Remove saved token ({{ slot }})</UiButton
              >
            </div>
          </fieldset>
          <div class="connection-delete">
            <UiButton variant="danger" :disabled="busy" @click="remove">Remove connection</UiButton>
          </div>
        </div></UiCard
      >
    </template>
    <UiButton v-if="error" @click="perform(reload)">Reload connection</UiButton>
  </div>
</template>
