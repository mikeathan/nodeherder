<script setup lang="ts">
  import UiIcon from '@/components/ui/UiIcon.vue';
  import type { ConnectionCatalogType } from '@/types/integrations';
  import {
    catalogStatusLabel,
    catalogStatusTone,
    integrationTypeIcon,
  } from '@/domain/integrations/catalog';

  defineProps<{
    type: ConnectionCatalogType;
    providerCount: number;
    selected: boolean;
  }>();
  defineEmits<{ select: [] }>();
</script>

<template>
  <button
    type="button"
    class="catalog-type"
    :class="{ 'is-selected': selected }"
    :aria-pressed="selected"
    @click="$emit('select')">
    <span class="catalog-type-icon"><UiIcon :name="integrationTypeIcon(type.id)" /></span>
    <span class="catalog-type-copy">
      <span class="catalog-type-heading">
        <strong>{{ type.name }}</strong>
        <span class="catalog-status" :data-tone="catalogStatusTone(type.worst_status)">
          {{ catalogStatusLabel(type.worst_status) }}
        </span>
      </span>
      <span class="catalog-type-description">{{ type.description }}</span>
      <span class="catalog-type-meta">
        {{ type.configured_count ?? 0 }} configured · {{ providerCount }}
        {{ providerCount === 1 ? 'provider' : 'providers' }}
      </span>
    </span>
  </button>
</template>

<style scoped>
  .catalog-type {
    display: flex;
    align-items: flex-start;
    gap: 0.8rem;
    width: 100%;
    padding: 1rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface);
    color: inherit;
    font: inherit;
    text-align: start;
    cursor: pointer;
  }
  .catalog-type:hover {
    border-color: var(--nh-accent);
  }
  .catalog-type.is-selected {
    border-color: var(--nh-accent);
    box-shadow: inset 0 0 0 1px var(--nh-accent);
  }
  .catalog-type:focus-visible {
    outline: 3px solid var(--nh-focus);
    outline-offset: 2px;
  }
  .catalog-type-icon {
    display: grid;
    place-items: center;
    flex: none;
    width: 2.6rem;
    height: 2.6rem;
    border-radius: 50%;
    background: var(--nh-surface-2);
    color: var(--nh-accent);
  }
  .catalog-type-icon :deep(.nh-ic) {
    width: 1.35rem;
    height: 1.35rem;
  }
  .catalog-type-copy {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    min-width: 0;
  }
  .catalog-type-heading {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }
  .catalog-type-heading strong {
    font-size: 0.98rem;
  }
  .catalog-type-description,
  .catalog-type-meta {
    color: var(--nh-text-muted);
    font-size: 0.8rem;
    overflow-wrap: anywhere;
  }
  .catalog-status {
    flex: none;
    padding: 0.15rem 0.45rem;
    border-radius: 999px;
    background: var(--nh-surface-2);
    color: var(--nh-text-muted);
    font-size: 0.68rem;
    white-space: nowrap;
  }
  .catalog-status[data-tone='ready'] {
    color: var(--nh-ok);
  }
  .catalog-status[data-tone='pending'] {
    color: var(--nh-info);
  }
  .catalog-status[data-tone='disabled'] {
    color: var(--nh-text-muted);
  }
  .catalog-status[data-tone='error'] {
    color: var(--nh-danger);
  }
</style>
