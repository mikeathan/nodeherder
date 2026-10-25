<script setup lang="ts">
  import { computed } from 'vue';
  import WasteBinIcon from './WasteBinIcon.vue';
  import UiCard from '@/components/ui/UiCard.vue';
  import { presentWasteSchedule } from '@/domain/waste-schedule';
  import type { WasteSchedule } from '@/types/integrations.type';
  const props = defineProps<{ resources: WasteSchedule[] }>();
  const resources = computed(() => props.resources.map(presentWasteSchedule));
</script>
<template>
  <p v-if="!resources.length" class="connection-muted">
    No collection dates have been saved yet. Check the pending lookup or use Refresh.
  </p>
  <section v-for="resource in resources" :key="resource.resource_id" class="connection-stack">
    <details class="waste-schedule-status">
      <summary>
        <span class="waste-status-label" :class="`is-${resource.freshness}`">{{ resource.statusLabel }}</span>
        <time :datetime="resource.last_checked_at" class="waste-checked">{{ resource.checkedLabel }}</time>
        <span class="waste-details-label">Details</span>
      </summary>
      <div class="waste-status-content">
        <dl>
          <div>
            <dt>Last checked</dt>
            <dd>{{ resource.checkedDetail }}</dd>
          </div>
          <div>
            <dt>Provider verification</dt>
            <dd>{{ resource.sourceDetail }}</dd>
          </div>
        </dl>
        <p>
          Refresh is manual. Dates use the UK calendar; limited coverage does not mean later collections are cancelled.
        </p>
        <p v-for="notice in resource.notices" :key="notice">{{ notice }}</p>
        <dl aria-label="Provider collection descriptions">
          <div v-for="service in resource.cards" :key="service.service_id">
            <dt>{{ service.binLabel ?? service.title }}</dt>
            <dd>
              <span>{{ service.name }}</span
              ><small>{{ service.coverageLabel }}</small>
            </dd>
          </div>
        </dl>
      </div>
    </details>
    <div class="connection-grid waste-collection-grid">
      <UiCard
        v-for="service in resource.cards"
        :key="service.service_id"
        :title="service.title"
        class="waste-collection-card">
        <template #tools><WasteBinIcon :colour="service.bin_colour" /></template>
        <p class="connection-date">
          {{
            service.nextDate ??
            (service.date_state === 'unavailable' ? 'Dates unavailable' : 'No upcoming date supplied')
          }}
        </p>
        <p v-if="service.subscription === 'required'" class="connection-muted">
          This service requires a subscription. Check that your address is subscribed.
        </p>
        <ul v-if="service.laterDates.length" class="connection-list">
          <li v-for="date in service.laterDates" :key="date">{{ date }}</li>
        </ul>
      </UiCard>
    </div>
  </section>
</template>

<style scoped>
  /* Shared card stacking margins do not belong in this grid. Each desktop row shares its
     header track, keeping wrapped names and the date content aligned without truncation. */
  .waste-collection-grid > .waste-collection-card {
    margin: 0;
    min-width: 0;
    display: grid;
    grid-template-rows: minmax(4.75rem, auto) 1fr;
  }
  .waste-collection-card :deep(.nh-card-head) {
    min-width: 0;
    display: grid;
    grid-template-columns: minmax(0, 1fr) 3rem;
    align-items: center;
    gap: 0.75rem;
    padding: 0.5rem 1rem;
  }
  .waste-collection-card :deep(.nh-card-title) {
    flex: 1;
    font-size: 1rem;
    min-width: 0;
    margin: 0;
    overflow-wrap: anywhere;
  }
  .waste-collection-card :deep(.nh-card-tools) {
    margin: 0;
    justify-self: end;
  }
  .waste-collection-card .connection-date {
    margin: 0;
    font-size: 1.1rem;
    line-height: 1.4;
    font-weight: 650;
  }
  .waste-schedule-status {
    padding: 0;
    border-radius: var(--nh-radius-m);
    font-size: 0.85rem;
  }
  .waste-schedule-status > summary {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    min-height: 44px;
    margin: 0;
    padding: 0.5rem 0.75rem;
    list-style: none;
    font-weight: 500;
  }
  .waste-schedule-status > summary::-webkit-details-marker {
    display: none;
  }
  .waste-status-label::before {
    content: '';
    display: inline-block;
    width: 0.4rem;
    height: 0.4rem;
    margin-right: 0.45rem;
    border-radius: 50%;
    background: var(--nh-text-muted);
  }
  .waste-status-label.is-fresh::before {
    background: var(--nh-success, #34a76a);
  }
  .waste-status-label.is-stale::before {
    background: var(--nh-warning, #d39434);
  }
  .waste-checked {
    color: var(--nh-text-muted);
  }
  .waste-details-label {
    margin-left: auto;
    color: var(--nh-accent);
  }
  .waste-details-label::after {
    content: '⌄';
    margin-left: 0.35rem;
  }
  .waste-schedule-status[open] .waste-details-label::after {
    content: '⌃';
  }
  .waste-status-content {
    padding: 0 0.75rem 0.75rem;
    color: var(--nh-text-muted);
  }
  .waste-status-content dl {
    display: grid;
    gap: 0.5rem;
    margin: 0.5rem 0;
  }
  .waste-status-content dl > div {
    display: grid;
    grid-template-columns: minmax(7rem, 1fr) 2fr;
    gap: 0.5rem;
  }
  .waste-status-content dt {
    font-weight: 600;
  }
  .waste-status-content dd {
    margin: 0;
    overflow-wrap: anywhere;
  }
  .waste-status-content dd small {
    display: block;
    margin-top: 0.2rem;
  }
  @media (max-width: 480px) {
    .waste-schedule-status > summary {
      gap: 0.5rem;
      font-size: 0.75rem;
    }
    .waste-status-content dl > div {
      grid-template-columns: 1fr;
      gap: 0.1rem;
    }
  }
  @supports (grid-template-rows: subgrid) {
    .waste-collection-grid > .waste-collection-card {
      grid-row: span 2;
      grid-template-rows: subgrid;
      row-gap: 0;
    }
  }
</style>
