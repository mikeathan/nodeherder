<script setup lang="ts">
  import { computed, nextTick, ref } from 'vue';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import ConnectionFeedback from '../shared/ConnectionFeedback.vue';
  import ConnectionBoard from './ConnectionBoard.vue';
  import ConnectionCatalogDialog from './ConnectionCatalogDialog.vue';
  import { useConnectionsBoard } from '@/composables/integrations/useConnectionsBoard';
  import {
    integrationTypeIcon,
    integrationTypeRank,
    integrationTypeTitle,
  } from '@/domain/integrations/catalog';
  import {
    connectionEntryStatus,
    connectionMatchesQuery,
    type ConnectionEntryStatus,
  } from '@/domain/integrations/connection-board';
  import '../connections.css';

  const { state, byId, busy, error, load, refresh, remember, catalog, layout, reloadLayout } = useConnectionsBoard();
  const { editing, visible, hidden } = layout;
  const { open } = catalog;
  const root = ref<HTMLElement | null>(null);
  const search = ref('');
  const typeFilter = ref('');
  const statusFilter = ref<ConnectionEntryStatus | 'all'>('all');
  const visibleInstances = computed(() =>
    visible.value.flatMap((entry) => byId.value.get(entry.instance_id) ?? [])
  );
  const hiddenInstances = computed(() => hidden.value.flatMap((entry) => byId.value.get(entry.instance_id) ?? []));
  const typeOptions = computed(() => {
    const counts = new Map<string, number>();
    for (const item of visibleInstances.value) {
      const typeId = item.provider_key.split('/')[0];
      if (typeId) counts.set(typeId, (counts.get(typeId) ?? 0) + 1);
    }
    return Array.from(counts, ([id, count]) => ({
      id,
      title: integrationTypeTitle(id, state.value.catalog),
      icon: integrationTypeIcon(id),
      count,
    })).sort(
      (left, right) =>
        integrationTypeRank(left.id) - integrationTypeRank(right.id) || left.title.localeCompare(right.title)
    );
  });
  const filteredEntries = computed(() =>
    visible.value.filter((entry) => {
      const item = byId.value.get(entry.instance_id);
      if (!item) return false;
      const typeId = item.provider_key.split('/')[0];
      return (
        (!typeFilter.value || typeId === typeFilter.value) &&
        (statusFilter.value === 'all' || connectionEntryStatus(item) === statusFilter.value) &&
        connectionMatchesQuery(item, search.value, state.value.catalog)
      );
    })
  );
  const filteredCount = computed(() => filteredEntries.value.length);
  const hasFilters = computed(
    () => !!search.value.trim() || !!typeFilter.value || statusFilter.value !== 'all'
  );
  function clearFilters() {
    search.value = '';
    typeFilter.value = '';
    statusFilter.value = 'all';
  }
  async function focusAction(action: string) {
    await nextTick();
    root.value?.querySelector<HTMLElement>(`[data-board-${action}]`)?.focus();
  }
  async function saveLayout() {
    if (await layout.save()) await focusAction('edit');
  }
  function hide(id: string) {
    layout.hide(id);
    void focusAction('save');
  }
  function cancelLayout() {
    layout.cancel();
    void focusAction('edit');
  }
  async function restore(id: string) {
    if (await layout.restore(id)) open.value = false;
  }
</script>
<template>
  <div ref="root" class="connections connections-board-page" :aria-busy="busy">
    <UiPageHeader title="Integrations">
      <template #actions>
        <template v-if="editing">
          <UiButton :disabled="busy" @click="cancelLayout">Cancel</UiButton>
          <UiButton variant="primary" icon="check" :disabled="busy" data-board-save @click="saveLayout"
            >Save layout</UiButton
          >
        </template>
        <template v-else>
          <UiButton
            v-if="state.instances.length"
            icon="edit"
            :disabled="busy || !state.board"
            data-board-edit
            @click="layout.begin"
            >Edit layout</UiButton
          >
          <UiButton
            variant="primary"
            icon="add"
            :disabled="busy || !state.catalog?.available"
            data-board-add
            @click="open = true"
            >Add integration</UiButton
          >
        </template>
      </template>
    </UiPageHeader>
    <ConnectionFeedback :error="error" :busy="busy" compact-busy />
    <p v-if="busy && !state.catalog" class="connection-muted">Loading integrations…</p>
    <UiButton v-if="error" :disabled="busy" @click="reloadLayout">Reload integrations</UiButton>
    <div v-if="state.catalog?.availability_error" class="connection-notice" role="status">
      <p>{{ state.catalog.availability_error.message }}</p>
      <UiButton :disabled="busy" @click="load">Check again</UiButton>
    </div>
    <section v-if="visibleInstances.length" class="connection-toolbar" aria-label="Find integrations">
      <label class="connection-search">
        <span>Search integrations</span>
        <input
          v-model="search"
          type="search"
          placeholder="Name, type, provider or postcode"
          autocomplete="off" />
      </label>
      <div class="connection-filter-row">
        <div class="connection-filter-types" role="group" aria-label="Filter by type">
          <button
            type="button"
            class="connection-filter-chip"
            :class="{ 'is-active': !typeFilter }"
            :aria-pressed="!typeFilter"
            @click="typeFilter = ''">
            All <span>{{ visibleInstances.length }}</span>
          </button>
          <button
            v-for="option in typeOptions"
            :key="option.id"
            type="button"
            class="connection-filter-chip"
            :class="{ 'is-active': typeFilter === option.id }"
            :aria-pressed="typeFilter === option.id"
            @click="typeFilter = option.id">
            <UiIcon :name="option.icon" />
            {{ option.title }} <span>{{ option.count }}</span>
          </button>
        </div>
        <label class="connection-filter-status"
          ><span>Status</span>
          <select v-model="statusFilter">
            <option value="all">All statuses</option>
            <option value="ready">Ready</option>
            <option value="attention">Needs attention</option>
            <option value="disabled">Disabled</option>
          </select>
        </label>
      </div>
      <div class="connection-filter-summary">
        <span>{{ filteredCount }} of {{ visibleInstances.length }} integrations</span>
        <UiButton v-if="hasFilters" variant="ghost" size="sm" @click="clearFilters">Clear filters</UiButton>
      </div>
    </section>
    <p v-if="editing" class="connection-muted">
      Use the arrows to change the order. Removing an entry keeps its integration configured.
    </p>
    <ConnectionBoard
      v-if="visible.length"
      :entries="filteredEntries"
      :instances="byId"
      :catalog="state.catalog"
      :editing="editing"
      :busy="busy"
      :filters-active="hasFilters"
      @open="remember"
      @refresh="refresh"
      @move="layout.move"
      @hide="hide" />
    <div v-else-if="state.catalog?.available && !busy" class="connection-board-empty">
      <h2>{{ hidden.length ? 'No integrations on this page' : 'No integrations yet' }}</h2>
      <p>
        {{
          hidden.length
            ? 'Use Add integration to show a hidden integration or set up another.'
            : 'Add an integration to see its information here.'
        }}
      </p>
    </div>
    <ConnectionCatalogDialog
      :open="open"
      :catalog="state.catalog"
      :hidden="hiddenInstances"
      :busy="busy"
      :provider-route="catalog.providerRoute"
      @close="open = false"
      @add="catalog.add"
      @restore="restore"
      @reload="catalog.reloadProviders" />
  </div>
</template>
<style scoped>
  .connections-board-page > .nh-btn {
    margin-bottom: 1rem;
  }
  .connection-board-empty {
    padding: 2.5rem 1rem;
    text-align: center;
    color: var(--nh-text-muted);
  }
  .connection-board-empty h2 {
    color: var(--nh-text);
    font-size: 1.15rem;
  }
  .connection-toolbar {
    display: grid;
    gap: 0.7rem;
    margin-bottom: 1rem;
  }
  .connection-search {
    display: grid;
    gap: 0.4rem;
    max-width: none;
    margin-bottom: 0;
    font-weight: 600;
  }
  .connection-search input {
    box-sizing: border-box;
    width: 100%;
    min-height: 44px;
    padding: 0.6rem 0.75rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
    color: var(--nh-text);
    font: inherit;
  }
  .connection-filter-row {
    display: grid;
    grid-template-columns: minmax(0, 1fr) auto;
    align-items: end;
    gap: 0.75rem;
  }
  .connection-filter-types {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
  }
  .connection-filter-chip {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    min-height: 36px;
    padding: 0.3rem 0.65rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: 999px;
    background: var(--nh-surface);
    color: var(--nh-text-muted);
    font: 500 0.8rem/1 var(--nh-font-ui);
    cursor: pointer;
  }
  .connection-filter-chip:hover {
    border-color: var(--nh-border-strong);
    color: var(--nh-text);
  }
  .connection-filter-chip:focus-visible {
    outline: 3px solid var(--nh-focus);
    outline-offset: 2px;
  }
  .connection-filter-chip.is-active {
    border-color: var(--nh-accent);
    background: var(--nh-accent-soft);
    color: var(--nh-text);
  }
  .connection-filter-chip :deep(.nh-ic) {
    width: 1rem;
    height: 1rem;
  }
  .connection-filter-chip span {
    display: grid;
    place-items: center;
    min-width: 1.2rem;
    height: 1.2rem;
    padding: 0 0.3rem;
    border-radius: 999px;
    background: var(--nh-surface-2);
    color: var(--nh-text-muted);
    font-size: 0.68rem;
  }
  .connection-filter-status {
    display: grid;
    gap: 0.3rem;
    color: var(--nh-text-muted);
    font-size: 0.78rem;
    font-weight: 600;
  }
  .connection-filter-status select {
    min-height: 36px;
    padding: 0.3rem 0.6rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
    color: var(--nh-text);
    font: inherit;
  }
  .connection-filter-summary {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    color: var(--nh-text-muted);
    font-size: 0.8rem;
  }
  @media (max-width: 720px) {
    .connection-filter-row {
      grid-template-columns: 1fr;
    }
    .connection-filter-status {
      max-width: 12rem;
    }
  }
</style>
