<script setup lang="ts">
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import WasteScheduleResults from './WasteScheduleResults.vue';
  import ConnectionSettings from './ConnectionSettings.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import ConnectionFeedback from './ConnectionFeedback.vue';
  import { useConnectionDetail } from '@/composables/integrations/useConnectionDetail';
  import { RouteName } from '@/types/router';
  import './connections.css';
  const {
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
  } = useConnectionDetail();
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
      <WasteScheduleResults v-if="tab === 'results'" :resources="item.resources" />
      <ConnectionSettings
        v-else
        v-model:name="name"
        v-model:token="token"
        :item="item"
        :definition="definition"
        :busy="busy"
        @update="update"
        @edit-address="editAddress"
        @credential="credential"
        @remove="remove" />
    </template>
    <UiButton v-if="error" @click="reloadPage">Reload connection</UiButton>
  </div>
</template>
