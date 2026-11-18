<script setup lang="ts">
  // Grouped entry rows; layout actions stay in the page and each entry remains presentation-only.
  import { computed, reactive, watch } from 'vue';
  import type {
    ConnectionBoardEntry,
    ConnectionCatalog,
    ConnectionInstance,
  } from '@/types/integrations';
  import { integrationTypeIcon, integrationTypeRank, integrationTypeTitle } from '@/domain/integrations/catalog';
  import IntegrationEntry from '../entries/IntegrationEntry.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const props = defineProps<{
    entries: ConnectionBoardEntry[];
    instances: Map<string, ConnectionInstance>;
    catalog: ConnectionCatalog | null;
    editing: boolean;
    busy: boolean;
    filtersActive: boolean;
  }>();
  const emit = defineEmits<{
    move: [id: string, offset: number];
    hide: [id: string];
    refresh: [item: ConnectionInstance];
    open: [id: string, event: MouseEvent];
  }>();

  interface EntryGroup {
    id: string;
    title: string;
    icon: ReturnType<typeof integrationTypeIcon>;
    items: ConnectionInstance[];
  }

  const collapsed = reactive(new Set<string>());
  watch(
    () => props.entries,
    () => collapsed.clear()
  );
  const groups = computed<EntryGroup[]>(() => {
    const grouped = new Map<string, EntryGroup>();
    for (const entry of props.entries) {
      const item = props.instances.get(entry.instance_id);
      if (!item) continue;
      const typeId = item.provider_key.split('/')[0];
      const group = grouped.get(typeId) ?? {
        id: typeId,
        title: integrationTypeTitle(typeId, props.catalog),
        icon: integrationTypeIcon(typeId),
        items: [],
      };
      group.items.push(item);
      grouped.set(typeId, group);
    }
    return Array.from(grouped.values()).sort(
      (left, right) => integrationTypeRank(left.id) - integrationTypeRank(right.id) || left.title.localeCompare(right.title)
    );
  });
  function providerName(item: ConnectionInstance): string {
    return props.catalog?.providers.find((provider) => provider.key === item.provider_key)?.name ?? '';
  }
  function toggleGroup(id: string) {
    if (collapsed.has(id)) collapsed.delete(id);
    else collapsed.add(id);
  }
</script>

<template>
  <div class="connection-entry-groups">
    <section v-for="group in groups" :key="group.id" class="connection-entry-group" :aria-label="group.title">
      <header>
        <h2>
          <button
            type="button"
            class="connection-entry-group-toggle"
            :aria-expanded="!collapsed.has(group.id)"
            @click="toggleGroup(group.id)">
            <span class="connection-entry-group-icon"><UiIcon :name="group.icon" /></span>
            <span class="connection-entry-group-title">{{ group.title }}</span>
            <span class="connection-entry-group-count">{{ group.items.length }}</span>
            <UiIcon
              class="connection-entry-group-chevron"
              :class="{ 'is-collapsed': collapsed.has(group.id) }"
              name="chevronDown" />
          </button>
        </h2>
      </header>
      <div v-show="!collapsed.has(group.id)" class="connection-entry-group-list">
        <IntegrationEntry
          v-for="(item, index) in group.items"
          :key="item.id"
          :item="item"
          :type-title="group.title"
          :provider-name="providerName(item)"
          :editing="editing"
          :busy="busy"
          :first="index === 0"
          :last="index === group.items.length - 1"
          @open="emit('open', item.id, $event)"
          @refresh="emit('refresh', item)"
          @hide="emit('hide', item.id)"
          @move="emit('move', item.id, $event)" />
      </div>
    </section>
    <p v-if="filtersActive && !groups.length" class="connection-entry-empty">
      No integrations match these filters.
    </p>
  </div>
</template>

<style scoped>
  .connection-entry-groups {
    display: grid;
    gap: 1.5rem;
  }
  .connection-entry-group {
    display: grid;
    gap: 0.55rem;
    min-width: 0;
  }
  .connection-entry-group-toggle {
    display: flex;
    align-items: center;
    gap: 0.55rem;
    width: 100%;
    padding: 0;
    border: 0;
    background: transparent;
    color: inherit;
    font: inherit;
    text-align: start;
    cursor: pointer;
  }
  .connection-entry-group-toggle:hover .connection-entry-group-title {
    text-decoration: underline;
  }
  .connection-entry-group-toggle:focus-visible {
    outline: 3px solid var(--nh-focus);
    outline-offset: 0.2rem;
    border-radius: var(--nh-radius-s);
  }
  .connection-entry-group-icon {
    display: grid;
    place-items: center;
    width: 1.8rem;
    height: 1.8rem;
    border-radius: 50%;
    background: var(--nh-surface-2);
    color: var(--nh-accent);
  }
  .connection-entry-group-icon :deep(.nh-ic) {
    width: 1rem;
    height: 1rem;
  }
  .connection-entry-group h2 {
    margin: 0;
    font-size: 0.95rem;
  }
  .connection-entry-group-count {
    display: grid;
    place-items: center;
    min-width: 1.5rem;
    height: 1.5rem;
    padding: 0 0.35rem;
    border-radius: 999px;
    background: var(--nh-surface-2);
    color: var(--nh-text-muted);
    font-size: 0.72rem;
  }
  .connection-entry-group-chevron {
    margin-inline-start: auto;
    color: var(--nh-text-muted);
    transition: transform var(--nh-motion);
  }
  .connection-entry-group-chevron.is-collapsed {
    transform: rotate(-90deg);
  }
  .connection-entry-group-list {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(min(100%, 23rem), 1fr));
    align-items: stretch;
    gap: 0.6rem;
  }
  .connection-entry-empty {
    padding: 2rem 1rem;
    color: var(--nh-text-muted);
    text-align: center;
  }
</style>
