<script setup lang="ts">
  import IntegrationPageHeader from './IntegrationPageHeader.vue';
  import { resourcePresenter } from '@/integrations/registry';
  import ConnectionSettings from './ConnectionSettings.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import ConnectionFeedback from './ConnectionFeedback.vue';
  import IntegrationSections from './IntegrationSections.vue';
  import { useConnectionDetail } from '@/composables/integrations/useConnectionDetail';
  import { formatLocalDateTime } from '@/utils/date.utils';
  import './connections.css';
  const {
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
  } = useConnectionDetail();
</script>
<template>
  <div class="connections connection-stack connection-detail" :aria-busy="busy">
    <IntegrationPageHeader :title="item?.name ?? 'Integration'" :subtitle="item?.postcode">
      <template #actions>
        <UiButton :disabled="busy || !item?.enabled" icon="refresh" @click="refresh">Refresh</UiButton>
      </template>
    </IntegrationPageHeader>
    <IntegrationSections v-if="item" :active="tab" :context="sectionContext" />
    <ConnectionFeedback :error="error ?? item?.last_error" :busy="busy" compact-busy />
    <p v-if="busy && !item" class="connection-muted">Loading integration…</p>
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
      <div v-if="!item.enabled" class="connection-notice" role="status">
        This integration is disabled. Saved dates remain visible.
      </div>
      <div v-if="item.operation_id" class="connection-notice" role="status">
        <p>A lookup is in progress. Check status continues that lookup without submitting a new one.</p>
        <UiButton :disabled="busy || !item.enabled" @click="check">Check status</UiButton>
      </div>
      <template v-if="tab === 'results'">
        <component
          v-for="resource in item.resources"
          :key="resource.resource_id"
          :is="resourcePresenter(resource)?.results"
          :resource="resource" />
        <p v-if="!item.resources.length" class="connection-muted">
          No information has been saved yet. Use Refresh to retrieve it.
        </p>
      </template>
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
      <details v-if="item.runs.length">
        <summary>Recent checks</summary>
        <ol class="connection-list">
          <li v-for="run in [...item.runs].reverse()" :key="run.operation_id">
            {{ formatLocalDateTime(run.at) }} — {{ run.message }}
          </li>
        </ol>
      </details>
    </template>
    <UiButton v-if="error" @click="reloadPage">Reload integration</UiButton>
  </div>
</template>
