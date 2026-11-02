<script setup lang="ts">
  import { computed, nextTick, ref, watch } from 'vue';
  import UiPager from '@/components/ui/UiPager.vue';
  import { fuelPriceValue } from '@/domain/integrations/fuel';
  import type { FuelPrices } from '@/types/integrations/fuel';
  import { formatLocalDateTime } from '@/utils/date.utils';
  import { formatQuantity } from '@/utils/units';

  const PAGE_SIZE = 10;
  const props = defineProps<{ fuel: FuelPrices }>();
  const stations = computed(() => props.fuel.stations);
  const page = ref(1);
  const pageCount = computed(() => Math.max(1, Math.ceil(stations.value.length / PAGE_SIZE)));
  const visibleStations = computed(() =>
    stations.value.slice((page.value - 1) * PAGE_SIZE, page.value * PAGE_SIZE)
  );
  const resultsStart = ref<HTMLElement>();
  const radius = computed(() => formatQuantity(props.fuel.radius_metres, 'distance'));
  const stale = computed(() => props.fuel.freshness === 'stale');
  const pageStatus = computed(() => {
    const total = stations.value.length;
    return `Page ${page.value} of ${pageCount.value} · ${total} ${total === 1 ? 'station' : 'stations'}`;
  });
  function changePage(next: number, scroll = false) {
    const clamped = Math.min(Math.max(1, next), pageCount.value);
    if (clamped === page.value) return;
    page.value = clamped;
    if (scroll) void nextTick(() => resultsStart.value?.scrollIntoView({ block: 'start' }));
  }
  watch(
    [() => props.fuel.resource_id, () => props.fuel.revision],
    () => {
      page.value = 1;
    }
  );
  watch(pageCount, (count) => {
    if (page.value > count) page.value = count;
  });
</script>

<template>
  <div class="fuel-results">
    <section class="fuel-panel" aria-label="Fuel-price summary">
      <p class="fuel-muted">{{ fuel.origin.label }} · within {{ radius }}</p>
      <div class="fuel-summary">
        <strong>{{ stations.length }}</strong>
        <span>{{ stations.length === 1 ? 'station' : 'stations' }}</span>
      </div>
      <p class="fuel-muted">Prices are rounded to the nearest penny per litre.</p>
      <p class="fuel-muted">Captured {{ formatLocalDateTime(fuel.captured_at) }}</p>
      <p class="fuel-muted">Last checked {{ formatLocalDateTime(fuel.last_checked_at) }}</p>
      <p v-if="stale" role="status">This fuel-price data is out of date.</p>
    </section>

    <UiPager
      :page="page"
      :page-count="pageCount"
      :status="pageStatus"
      label="Fuel station pages"
      announce-status
      @update:page="changePage($event)" />

    <div ref="resultsStart" class="fuel-page-start" />

    <p v-if="!stations.length" class="fuel-empty">No nearby stations were returned for this radius.</p>

    <article v-for="station in visibleStations" :key="station.station_id" class="fuel-station">
      <header>
        <div>
          <h2>{{ station.name }}</h2>
          <p>{{ [station.brand, station.address, station.postcode].filter(Boolean).join(' · ') }}</p>
        </div>
        <span v-if="station.distance_metres != null" class="fuel-distance">
          {{ formatQuantity(station.distance_metres, 'distance') }}
        </span>
      </header>
      <div v-if="station.prices.length" class="fuel-prices">
        <div v-for="price in station.prices" :key="price.grade">
          <span>{{ price.label }}</span>
          <strong>{{ fuelPriceValue(price) }}</strong>
          <small v-if="price.updated_at">Updated {{ formatLocalDateTime(price.updated_at) }}</small>
        </div>
      </div>
      <p v-else class="fuel-muted">No price was returned for this station.</p>
    </article>

    <UiPager
      :page="page"
      :page-count="pageCount"
      :status="pageStatus"
      label="Fuel station pages, bottom"
      @update:page="changePage($event, true)" />

    <p class="fuel-muted">{{ fuel.attribution }}</p>
  </div>
</template>

<style scoped>
  .fuel-results {
    display: grid;
    gap: 1rem;
  }
  .fuel-panel,
  .fuel-station {
    min-width: 0;
    padding: 1.25rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface);
  }
  .fuel-summary {
    display: flex;
    align-items: baseline;
    gap: 0.45rem;
    margin: 0.5rem 0 0.75rem;
  }
  .fuel-summary strong {
    font-size: 2rem;
    line-height: 1;
  }
  .fuel-summary span {
    color: var(--nh-text-muted);
  }
  .fuel-muted {
    margin: 0.5rem 0 0;
    color: var(--nh-text-muted);
    font-size: 0.8rem;
  }
  .fuel-empty {
    padding: 1.5rem;
    border: var(--nh-border-w) dashed var(--nh-border);
    border-radius: var(--nh-radius-l);
    color: var(--nh-text-muted);
    text-align: center;
  }
  .fuel-page-start {
    scroll-margin-top: 6rem;
  }
  .fuel-station header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
  }
  .fuel-station h2 {
    margin: 0;
    font-size: 1.1rem;
  }
  .fuel-station header p {
    margin: 0.35rem 0 0;
    color: var(--nh-text-muted);
    font-size: 0.82rem;
    overflow-wrap: anywhere;
  }
  .fuel-distance {
    flex: none;
    padding: 0.25rem 0.55rem;
    border-radius: 999px;
    background: var(--nh-surface-2);
    color: var(--nh-text-muted);
    font-size: 0.75rem;
    white-space: nowrap;
  }
  .fuel-prices {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(8rem, 1fr));
    gap: 0.65rem;
    margin-top: 1rem;
  }
  .fuel-prices > div {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    padding: 0.75rem;
    border-radius: var(--nh-radius-m);
    background: var(--nh-surface-2);
  }
  .fuel-prices span,
  .fuel-prices small {
    color: var(--nh-text-muted);
    font-size: 0.75rem;
  }
  .fuel-prices strong {
    font-size: 1.25rem;
  }
  @media (max-width: 700px) {
    .fuel-panel,
    .fuel-station {
      padding: 1rem;
    }
    .fuel-prices {
      grid-template-columns: repeat(2, 1fr);
    }
  }
</style>
