<script setup lang="ts">
  import { computed } from 'vue';
  import type { HomeIntegrationCard } from '@/types/home.type';
  import type { HomeSourcePresentation } from '@/domain/home-cards';
  import { homeIntegrationCard } from '@/domain/home-cards';
  import { RouteName } from '@/types/router';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import IntegrationStatus from '@/components/integrations/IntegrationStatus.vue';

  const props = defineProps<{
    card: HomeIntegrationCard;
    sources: ReadonlyMap<string, HomeSourcePresentation>;
    editing?: boolean;
    busy?: boolean;
    panel?: boolean;
    sourceError?: string | null;
  }>();
  defineEmits<{ remove: [] }>();
  const item = computed(() => homeIntegrationCard(props.card, props.sources));
  const destination = computed(() => ({
    name: RouteName.Integration,
    params: { id: props.card.instance_id },
    hash: '#results',
    query: { returnTo: props.panel ? '/panel' : '/groupdashboard' },
  }));
  const status = computed(() => (props.sourceError ? 'Offline · showing saved information' : item.value.status));
</script>

<template>
  <article
    class="nh-integration-tile"
    :class="{
      'is-summary': item.rows.length > 0,
      'is-panel': panel,
      'has-warning': item.warning || sourceError,
    }"
    :data-home-card="card.id"
    :aria-label="item.title">
    <div class="nh-integration-body">
      <span class="nh-integration-badge">
        <component v-if="item.icon" :is="item.icon.component" v-bind="item.icon.props" />
        <UiIcon v-else name="connections" />
      </span>
      <div class="nh-integration-copy">
        <strong v-if="editing" class="nh-integration-name">{{ item.title }}</strong>
        <RouterLink v-else class="nh-integration-name nh-integration-link" :to="destination">
          {{ item.title }}
        </RouterLink>
        <span v-if="!item.rows.length" class="nh-integration-value">{{ item.value }}</span>
        <span v-if="item.detail" class="nh-integration-value">{{ item.detail }}</span>
        <span class="nh-integration-origin">{{ item.origin }}</span>
      </div>
    </div>
    <ul v-if="item.rows.length" class="nh-integration-rows">
      <li v-for="row in item.rows" :key="row.id">
        <span>
          <component v-if="row.icon" :is="row.icon.component" v-bind="row.icon.props" />
          {{ row.title }}
        </span>
        <span>{{ row.value }}</span>
      </li>
    </ul>
    <IntegrationStatus class="nh-integration-status" :status="status" :checked="item.checked"
      :error="item.error || !!sourceError" :warning="item.warning" />
    <button
      v-if="editing"
      class="nh-tile-del"
      type="button"
      :disabled="busy"
      :aria-label="`Remove ${item.title} from Home`"
      @click="$emit('remove')">
      <UiIcon name="close" />
    </button>
  </article>
</template>

<style scoped>
  .nh-integration-tile {
    position: relative;
    min-width: 0;
    min-height: var(--nh-tile-h);
    padding: 0.7rem 0.75rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface);
    box-shadow: var(--nh-shadow-1);
  }
  .nh-integration-body {
    display: flex;
    align-items: flex-start;
    gap: 0.65rem;
  }
  .nh-integration-badge {
    display: grid;
    place-items: center;
    width: 2.5rem;
    height: 2.5rem;
    flex: none;
    border-radius: 50%;
    background: var(--nh-surface-2);
    color: var(--nh-accent);
  }
  .nh-integration-badge :deep(svg) {
    width: 1.5rem;
    height: 2rem;
  }
  .nh-integration-copy {
    min-width: 0;
    display: flex;
    flex-direction: column;
  }
  .nh-integration-name {
    font-size: 0.875rem;
    font-weight: 600;
    color: var(--nh-text);
    text-decoration: none;
    overflow: hidden;
    white-space: nowrap;
    text-overflow: ellipsis;
  }
  .nh-integration-link::after {
    content: '';
    position: absolute;
    inset: 0;
    border-radius: var(--nh-radius-l);
  }
  .nh-integration-link:focus-visible {
    outline: none;
  }
  .nh-integration-link:focus-visible::after {
    outline: 3px solid var(--nh-accent);
    outline-offset: 3px;
  }
  .nh-integration-value {
    font-size: 0.82rem;
    color: var(--nh-text-muted);
    overflow-wrap: anywhere;
  }
  .nh-integration-origin {
    font-size: 0.7rem;
    color: var(--nh-text-muted);
    overflow-wrap: anywhere;
  }
  .nh-integration-status {
    font-size: 0.67rem;
    color: var(--nh-text-muted);
    margin: 0.5rem 0 0;
    overflow-wrap: anywhere;
  }
  .nh-integration-rows {
    list-style: none;
    padding: 0;
    margin: 0.75rem 0 0;
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    font-size: 0.8rem;
  }
  .nh-integration-rows li,
  .nh-integration-rows li > span:first-child {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }
  .nh-integration-rows li {
    justify-content: space-between;
  }
  .nh-integration-rows :deep(svg) {
    width: 1.1rem;
    height: 1.5rem;
  }
  .nh-tile-del {
    position: absolute;
    top: -0.5rem;
    right: -0.5rem;
    z-index: 1;
    width: 1.75rem;
    height: 1.75rem;
    padding: 0;
    display: grid;
    place-items: center;
    border: 2px solid var(--nh-surface);
    border-radius: 50%;
    background: var(--nh-danger);
    color: white;
    cursor: pointer;
  }
  .nh-tile-del .nh-ic {
    width: 0.9rem;
    height: 0.9rem;
  }
  .is-panel {
    padding: 1.1rem;
    grid-column: 1 / -1;
    inline-size: 100%;
    max-inline-size: 24rem;
    justify-self: start;
  }
  .is-panel .nh-integration-name {
    white-space: normal;
    overflow-wrap: anywhere;
  }
  .is-panel .nh-integration-name,
  .is-panel .nh-integration-value {
    font-size: 1.05rem;
  }
  .is-panel .nh-integration-status {
    font-size: 0.8rem;
  }
  .is-panel .nh-integration-rows {
    font-size: 1rem;
    gap: 1rem;
  }
  .is-panel .nh-integration-rows li {
    column-gap: 1.5rem;
  }
  .is-panel .nh-integration-rows li > span:last-child {
    white-space: nowrap;
  }
</style>
