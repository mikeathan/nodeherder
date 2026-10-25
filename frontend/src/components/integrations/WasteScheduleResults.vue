<script setup lang="ts">
  import { computed } from 'vue';
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
        {{ resource.source_retrieved_at ? new Date(resource.source_retrieved_at).toLocaleString() : 'not supplied' }}.
      </p>
      <p class="connection-muted">
        Refresh is manual. Dates use the UK calendar; limited coverage does not mean later collections are cancelled.
      </p>
      <p v-for="notice in resource.notices" :key="notice">{{ notice }}</p>
    </div>
    <div class="connection-grid">
      <UiCard v-for="service in resource.cards" :key="service.service_id" :title="service.name" icon="schedule"
        ><p class="connection-date">
          {{
            service.nextDate ??
            (service.date_state === 'unavailable' ? 'Dates unavailable' : 'No upcoming date supplied')
          }}
        </p>
        <p v-if="service.bin_colour" class="connection-muted">{{ service.bin_colour }} bin</p>
        <p v-if="service.subscription === 'required'" class="connection-muted">
          This service requires a subscription. Check that your address is subscribed.
        </p>
        <ul v-if="service.laterDates.length" class="connection-list">
          <li v-for="date in service.laterDates" :key="date">{{ date }}</li>
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
