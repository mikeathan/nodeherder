<script setup lang="ts">
  import { computed } from 'vue';
  import { useHub } from '@/composables/useHub';
  import { supportsSpatialMapping } from '@/domain/spatial';
  import PresenceRailTile from './PresenceRailTile.vue';

  const { devices } = useHub();
  const sensors = computed(() => devices.value.filter((device) => supportsSpatialMapping(device)));
</script>

<template>
  <section class="nh-presence-rail-wrap" aria-label="Presence sensors">
    <div class="nh-presence-rail" role="list">
      <PresenceRailTile v-for="sensor in sensors" :key="sensor.id" :device="sensor" role="listitem" />
    </div>
    <p v-if="!sensors.length" class="nh-muted">No spatial sensors are available yet.</p>
  </section>
</template>

<style scoped>
  .nh-presence-rail-wrap { min-width: 0; }
  .nh-presence-rail {
    display: grid;
    gap: 0.25rem;
  }
  .nh-presence-rail > * { min-width: 0; }
</style>
