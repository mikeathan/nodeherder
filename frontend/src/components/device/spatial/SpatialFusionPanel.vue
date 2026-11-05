<script setup lang="ts">
  /*
   * Presentation-only fusion across all spatial-capable devices. Values come from the existing
   * derived presence devices; nothing is persisted and no automation is fed from this view.
   */
  import { computed } from 'vue';
  import type { FusedSensor } from '@/domain/spatial/fusion';
  import { fuseSensors } from '@/domain/spatial/fusion';
  import { derivedPresenceDeviceId, derivedPresenceValues, derivedZoneSignals } from '@/domain/spatial/derived';
  import type { Device } from '@/types/device';
  import { RouteName } from '@/types/router';
  import { useHub } from '@/composables/useHub';
  import { supportsSpatialMapping } from '@/domain/spatial';

  const { devices, findDevice, statusOf, isConnected } = useHub();

  const sensors = computed<FusedSensor[]>(() => {
    const result: FusedSensor[] = [];
    for (const device of devices.value) {
      if (!supportsSpatialMapping(device)) continue;
      const derived = findDevice(derivedPresenceDeviceId(device.id));
      if (!derived) continue;
      const values = derivedPresenceValues(derived);
      const zones = [...derivedZoneSignals(derived).entries()].map(([zoneId, signal]) => {
        const entity = Object.values(derived.exposes).find(
          (candidate) => candidate.attributes?.zoneId === zoneId && candidate.name.endsWith('_occupied')
        );
        return {
          name: typeof entity?.attributes?.zoneName === 'string' ? entity.attributes.zoneName : zoneId,
          occupied: signal.occupied,
        };
      });
      result.push({
        id: device.id,
        name: device.friendly_name,
        online: !!isConnected.value && statusOf(derived) === 'online',
        occupied: values.occupied,
        targetCount: values.targetCount,
        zones,
      });
    }
    return result.sort((left, right) => left.name.localeCompare(right.name));
  });
  const fusion = computed(() => fuseSensors(sensors.value));
  const occupiedZoneNames = (sensor: FusedSensor) =>
    sensor.zones.filter((zone) => zone.occupied === true).map((zone) => zone.name);
  const routeFor = (device: Device) => ({ name: RouteName.DevicePage, params: { id: device.id } });
</script>

<template>
  <section aria-label="Multi-sensor presence">
    <h3>Multi-sensor presence</h3>
    <p class="nh-muted">
      Combined only for display. Automations keep using each sensor's own derived entities, counts are reported
      targets rather than people, and an offline sensor is never treated as clear.
    </p>
    <p v-if="!sensors.length" class="nh-muted">
      No spatial sensor with derived presence yet. Save a spatial mapping and let each sensor report once.
    </p>
    <template v-else>
      <p class="nh-fusion-summary">
        <strong>{{ fusion.occupied === null ? 'Unknown' : fusion.occupied ? 'Occupied' : 'Clear' }}</strong>
        · {{ fusion.targetCount }} reported target{{ fusion.targetCount === 1 ? '' : 's' }} ·
        {{ fusion.online }}/{{ fusion.total }} sensor{{ fusion.total === 1 ? '' : 's' }} online
      </p>
      <ul v-if="fusion.notes.length" class="nh-fusion-notes">
        <li v-for="note in fusion.notes" :key="note">{{ note }}</li>
      </ul>
      <ul class="nh-fusion-sensors">
        <li v-for="sensor in sensors" :key="sensor.id">
          <div class="nh-fusion-heading">
            <strong>{{ sensor.name }}</strong>
            <span class="nh-muted">{{ sensor.online ? 'online' : 'offline' }}</span>
          </div>
          <p>
            {{ sensor.occupied === null ? 'Unknown' : sensor.occupied ? 'Occupied' : 'Clear' }}
            <template v-if="sensor.targetCount !== null">· {{ sensor.targetCount }} reported target{{ sensor.targetCount === 1 ? '' : 's' }}</template>
          </p>
          <p v-if="occupiedZoneNames(sensor).length" class="nh-muted">
            Occupied zones: {{ occupiedZoneNames(sensor).join(', ') }}
          </p>
          <RouterLink v-if="findDevice(sensor.id)" :to="routeFor(findDevice(sensor.id)!)">Open {{ sensor.name }}</RouterLink>
        </li>
      </ul>
    </template>
  </section>
</template>

<style scoped>
  .nh-fusion-summary {
    font-size: 1.05rem;
  }
  .nh-fusion-notes {
    color: var(--nh-warning, #b9770e);
    padding-inline-start: 1.2rem;
  }
  .nh-fusion-sensors {
    list-style: none;
    padding: 0;
    display: grid;
    gap: 0.6rem;
  }
  .nh-fusion-sensors > li {
    padding: 0.6rem 0.75rem;
    border: 1px solid var(--nh-border);
    border-radius: var(--nh-radius);
  }
  .nh-fusion-heading {
    display: flex;
    justify-content: space-between;
    gap: 0.5rem;
  }
</style>
