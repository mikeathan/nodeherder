<script setup lang="ts">
  import { computed, nextTick, ref } from 'vue';
  import UiPageHeader from '@/components/ui/UiPageHeader.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import ConnectionFeedback from './ConnectionFeedback.vue';
  import ConnectionBoard from './ConnectionBoard.vue';
  import ConnectionCatalogDialog from './ConnectionCatalogDialog.vue';
  import { useConnectionsBoard } from '@/composables/integrations/useConnectionsBoard';
  import './connections.css';

  const { state, byId, busy, error, load, refresh, remember, catalog, layout, reloadLayout } = useConnectionsBoard();
  const { editing, visible, hidden } = layout;
  const { open } = catalog;
  const root = ref<HTMLElement | null>(null);
  const hiddenInstances = computed(() => hidden.value.flatMap((entry) => byId.value.get(entry.instance_id) ?? []));
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
    <p v-if="editing" class="connection-muted">
      Drag a card or use the arrows to change the order. Removing a card keeps its integration configured.
    </p>
    <ConnectionBoard
      v-if="visible.length"
      :entries="visible"
      :instances="byId"
      :editing="editing"
      :busy="busy"
      @open="remember"
      @refresh="refresh"
      @move="layout.move"
      @hide="hide"
      @reorder="layout.reorder" />
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
</style>
