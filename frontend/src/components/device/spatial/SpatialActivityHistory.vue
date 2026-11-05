<script setup lang="ts">
  /*
   * Persisted zone activity from the hub's derived presence entities. Raw positions are not
   * stored, so this never implies a position replay or reconstructs places.
   */
  import { computed, ref } from 'vue';
  import type { Device } from '@/types/device';
  import type { SpatialWorkspace } from '@/types/spatial';
  import { useDerivedPresence } from '@/composables/spatial/useDerivedPresence';
  import { useHistoryQuery } from '@/composables/useHistoryQuery';
  import { buildZoneActivity, zoneActivityRequest, zoneHistoryEntities } from '@/domain/spatial/activity';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{ device: Device; workspace: SpatialWorkspace }>();
  const derived = useDerivedPresence(() => props.device);
  const zones = computed(() => props.workspace.zones);
  const range = ref<'1h' | '24h' | '7d'>('24h');
  const to = ref(Date.now());
  const spans = { '1h': 3_600_000, '24h': 86_400_000, '7d': 604_800_000 } as const;
  const from = computed(() => to.value - spans[range.value]);
  const entities = computed(() => zoneHistoryEntities(derived.derived.value, zones.value));
  const request = computed(() => zoneActivityRequest(derived.derived.value, zones.value, from.value, to.value));
  const history = useHistoryQuery(() => request.value);
  const activity = computed(() =>
    buildZoneActivity(history.entry.value?.series ?? [], zones.value, entities.value, from.value, to.value)
  );
  const status = computed(() => history.entry.value?.status ?? 'loading');

  function choose(next: '1h' | '24h' | '7d') {
    range.value = next;
    to.value = Date.now();
  }
  function refresh() {
    to.value = Date.now();
    history.refresh();
  }
</script>

<template>
  <section aria-label="Zone activity history">
    <h3>Zone activity</h3>
    <p class="nh-muted">
      Built from the hub’s recorded presence entities. Raw positions are not stored, so this shows zone occupancy and
      counts, not a position replay. Unknown time is shown as partial, never as empty.
    </p>
    <div class="nh-activity-toolbar">
      <div role="group" aria-label="History range">
        <UiButton
          v-for="option in (['1h', '24h', '7d'] as const)"
          :key="option"
          size="sm"
          :variant="range === option ? 'primary' : 'ghost'"
          :pressed="range === option"
          @click="choose(option)">
          {{ option === '1h' ? '1 hour' : option === '24h' ? '24 hours' : '7 days' }}
        </UiButton>
      </div>
      <UiButton size="sm" variant="ghost" :disabled="!request" @click="refresh">Refresh</UiButton>
    </div>

    <p v-if="!derived.derived.value" class="nh-muted">
      No derived presence device yet. Save a spatial mapping for this device first.
    </p>
    <p v-else-if="!request" class="nh-muted">
      No drawn zones have derived occupancy entities yet.
    </p>
    <p v-else-if="status === 'loading'" class="nh-muted">Loading recorded activity…</p>
    <p v-else-if="status === 'error'" class="nh-muted">{{ history.entry.value?.error || 'Could not load activity.' }}</p>
    <p v-else-if="!activity.length" class="nh-muted">No recorded activity in this range.</p>
    <ul v-else class="nh-activity-list">
      <li v-for="zone in activity" :key="zone.zoneId">
        <div class="nh-activity-heading">
          <strong>{{ zone.zoneName }}</strong>
          <span class="nh-muted">{{ zone.partial ? 'Partial data' : 'Complete range' }}</span>
        </div>
        <p>
          Occupied {{ zone.sharePct === null ? 'unknown' : `${Math.round(zone.sharePct)}%` }} of known time ·
          {{ zone.transitions }} transition{{ zone.transitions === 1 ? '' : 's' }} ·
          peak count {{ zone.peakCount ?? 'unknown' }} ·
          last occupied {{ zone.lastOccupiedAt ? new Date(zone.lastOccupiedAt).toLocaleString() : 'never in range' }}
        </p>
        <ol class="nh-activity-timeline" aria-label="Occupancy timeline">
          <li
            v-for="(bucket, index) in zone.buckets"
            :key="index"
            :title="bucket === null ? 'No data' : `${Math.round(bucket)}% occupied`"
            :class="{ 'is-empty': bucket === null, 'is-hot': (bucket ?? 0) > 50 }"
            :style="{ opacity: bucket === null ? 0.2 : 0.25 + (bucket / 100) * 0.75 }" />
        </ol>
      </li>
    </ul>
  </section>
</template>

<style scoped>
  .nh-activity-toolbar {
    display: flex;
    gap: 0.75rem;
    align-items: center;
    justify-content: space-between;
    flex-wrap: wrap;
  }
  .nh-activity-list {
    list-style: none;
    padding: 0;
    display: grid;
    gap: 0.9rem;
  }
  .nh-activity-heading {
    display: flex;
    justify-content: space-between;
    gap: 0.5rem;
  }
  .nh-activity-timeline {
    display: grid;
    grid-template-columns: repeat(48, minmax(0, 1fr));
    gap: 2px;
    list-style: none;
    padding: 0;
    margin: 0.4rem 0 0;
  }
  .nh-activity-timeline li {
    height: 0.9rem;
    border-radius: 2px;
    background: var(--nh-accent);
  }
  .nh-activity-timeline li.is-hot {
    background: var(--nh-danger, #c0392b);
  }
  .nh-activity-timeline li.is-empty {
    background: var(--nh-border);
  }
</style>
