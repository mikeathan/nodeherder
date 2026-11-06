<script setup lang="ts">
  import { computed } from 'vue';
  import { useHub } from '@/composables/useHub';
  import { supportsSpatialMapping } from '@/domain/spatial';
  import { derivedPresenceDeviceId, derivedPresenceValues } from '@/domain/spatial/derived';
  import { RouteName } from '@/types/router';

  const { devices, findDevice, statusOf, isConnected } = useHub();
  const sensors = computed(() =>
    devices.value
      .filter((device) => supportsSpatialMapping(device))
      .map((device) => {
        const derived = findDevice(derivedPresenceDeviceId(device.id));
        return derived
          ? { device, values: derivedPresenceValues(derived), online: !!isConnected.value && statusOf(derived) === 'online' }
          : null;
      })
      .filter((entry): entry is NonNullable<typeof entry> => entry !== null)
  );
  // A single sensor is already represented by its live view; the overview only adds value when
  // more than one can disagree.
  const visible = computed(() => sensors.value.length >= 2);
  const occupied = computed(() => sensors.value.filter((sensor) => sensor.values.occupied === true).length);
  const unknowns = computed(() => sensors.value.filter((sensor) => sensor.values.occupied === null).length);
</script>

<template>
  <section v-if="visible" class="nh-presence-overview" aria-label="All presence sensors">
    <header>
      <strong>Presence overview</strong>
      <span class="nh-muted">{{ sensors.length }} sensors · {{ occupied }} reporting presence<template v-if="unknowns"> · {{ unknowns }} unknown</template></span>
    </header>
    <ul>
      <li v-for="sensor in sensors" :key="sensor.device.id">
        <RouterLink :to="{ name: RouteName.DevicePage, params: { id: sensor.device.id } }">{{ sensor.device.friendly_name }}</RouterLink>
        <span :class="{ 'is-occupied': sensor.values.occupied }">
          {{ sensor.values.occupied === null ? 'Unknown' : sensor.values.occupied ? 'Occupied' : 'Clear' }}
        </span>
        <span v-if="sensor.values.targetCount !== null" class="nh-muted">{{ sensor.values.targetCount }} target{{ sensor.values.targetCount === 1 ? '' : 's' }}</span>
        <span v-if="!sensor.online" class="nh-muted">derived offline</span>
      </li>
    </ul>
  </section>
</template>

<style scoped>
  .nh-presence-overview { border: 1px solid var(--nh-border); border-radius: var(--nh-radius); padding: 0.65rem 0.8rem; margin-block-end: 1rem; }
  header { display: flex; justify-content: space-between; gap: 0.75rem; flex-wrap: wrap; }
  ul { list-style: none; padding: 0; margin: 0.5rem 0 0; display: grid; gap: 0.25rem; }
  li { display: flex; gap: 0.6rem; flex-wrap: wrap; align-items: baseline; }
  .is-occupied { color: var(--nh-accent); font-weight: 600; }
</style>
