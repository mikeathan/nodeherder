<script setup lang="ts">
  import type { Device } from '@/types/device';
  import { PRESENCE_COPY } from '@/domain/spatial/presence';
  import PresenceRailTile from './PresenceRailTile.vue';

  withDefaults(defineProps<{ sensors: readonly Device[]; showNames?: boolean }>(), { showNames: true });
</script>

<template>
  <section class="nh-presence-rail-wrap" :aria-label="PRESENCE_COPY.overview.sensorsAriaLabel">
    <div class="nh-presence-rail" role="list">
      <PresenceRailTile
        v-for="sensor in sensors"
        :key="sensor.id"
        :device="sensor"
        :show-name="showNames"
        role="listitem" />
    </div>
    <p v-if="!sensors.length" class="nh-muted">{{ PRESENCE_COPY.overview.noSensors }}</p>
  </section>
</template>

<style scoped>
  .nh-presence-rail-wrap { min-width: 0; }
  .nh-presence-rail {
    display: grid;
    gap: 0;
  }
  .nh-presence-rail > * { min-width: 0; }
</style>
