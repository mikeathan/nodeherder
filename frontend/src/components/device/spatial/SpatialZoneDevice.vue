<script setup lang="ts">
  /*
   * Maps a drawn zone to explicitly chosen firmware entities and shows requested versus
   * confirmed values. Rectangles use four coordinate entities; polygons use one text entity
   * in the firmware's `x:y;x:y;...` format. Translation/readiness rules live in
   * domain/spatial/device-zones.ts.
   */
  import { computed, ref, watch } from 'vue';
  import type { Device } from '@/types/device';
  import type { SpatialZone, ZoneEntityRole } from '@/types/spatial';
  import { ZONE_COORDINATE_ROLES, ZONE_POLYGON_ROLE } from '@/types/spatial';
  import {
    ZONE_APPLY_MESSAGES,
    occupancyFlagOptions,
    serializeZonePolygon,
    zoneCommandPlan,
    zoneCommandRows,
    zoneCoordinateOptions,
    zonePolygonOptions,
    zonePolygonPlan,
  } from '@/domain/spatial/device-zones';
  import { exposeDisplayLabel, formatExposeValue } from '@/domain/exposes';
  import { useDeviceCommand } from '@/composables/useDeviceCommand';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{ device: Device; zone: SpatialZone; online: boolean }>();
  const emit = defineEmits<{ occupancy: [id: string]; map: [role: ZoneEntityRole, id: string] }>();
  const { send, isPending, canControl } = useDeviceCommand();
  const flags = computed(() => occupancyFlagOptions(props.device));
  const occupancy = computed(() => props.device.exposes[props.zone.occupancyEntity ?? '']);
  const coordinates = computed(() => zoneCoordinateOptions(props.device.exposes));
  const polygons = computed(() => zonePolygonOptions(props.device.exposes));
  const plan = computed(() =>
    props.zone.shape === 'polygon' ? zonePolygonPlan(props.zone, props.device.exposes) : zoneCommandPlan(props.zone, props.device.exposes)
  );
  const commandRows = computed(() => zoneCommandRows(plan.value.commands, props.device.exposes));
  const polygonValue = computed(() => serializeZonePolygon(props.zone.points));
  const busy = computed(() => plan.value.commands.some((command) => isPending(props.device.id, command.entity)));
  const notice = ref('');
  watch(
    () => JSON.stringify(props.zone),
    () => { notice.value = ''; }
  );

  /** Sends the validated command set once; a partial submission is reported, never retried. */
  function apply() {
    if (plan.value.error || !props.online || !canControl(props.device) || busy.value) return;
    for (const command of plan.value.commands) {
      if (!send(props.device, command.entity, command.value)) {
        notice.value = ZONE_APPLY_MESSAGES.partial;
        return;
      }
    }
    notice.value = props.zone.shape === 'polygon' ? ZONE_APPLY_MESSAGES.polygonRequested : ZONE_APPLY_MESSAGES.requested;
  }
</script>

<template>
  <section aria-label="Physical zone configuration">
    <h4>Physical zone state</h4>
    <label>
      Reported occupancy entity
      <select
        class="nh-input"
        :value="zone.occupancyEntity ?? ''"
        @change="emit('occupancy', ($event.target as HTMLSelectElement).value)">
        <option value="">Not mapped</option>
        <option v-for="expose in flags" :key="expose.name" :value="expose.name">{{ exposeDisplayLabel(expose) }}</option>
      </select>
    </label>
    <p v-if="occupancy">
      Device-reported occupancy: {{ formatExposeValue(occupancy) }}{{ online ? '' : ' (offline; may be stale)' }}
    </p>

    <template v-if="zone.shape === 'rectangle'">
      <h4>Apply rectangle to device</h4>
      <p>
        Map this rectangle to the correct firmware slot. Each coordinate is sent separately; intermediate bounds can
        briefly affect occupancy. Configure firmware rectangle mode before applying.
      </p>
      <details>
        <summary>Map physical zone coordinates</summary>
        <label v-for="role in ZONE_COORDINATE_ROLES" :key="role">
          {{ role }} entity
          <select
            class="nh-input"
            :value="zone.entities?.[role] ?? ''"
            @change="emit('map', role, ($event.target as HTMLSelectElement).value)">
            <option value="">Not mapped</option>
            <option
              v-if="zone.entities?.[role] && !coordinates.some((expose) => expose.name === zone.entities?.[role])"
              :value="zone.entities[role]"
              disabled>
              Unavailable: {{ zone.entities[role] }}
            </option>
            <option v-for="expose in coordinates" :key="expose.name" :value="expose.name">{{ exposeDisplayLabel(expose) }}</option>
          </select>
        </label>
      </details>
    </template>

    <template v-else>
      <h4>Apply polygon to device</h4>
      <p>{{ ZONE_APPLY_MESSAGES.polygonMode }}</p>
      <label>
        Polygon text entity
        <select
          class="nh-input"
          :value="zone.entities?.[ZONE_POLYGON_ROLE] ?? ''"
          @change="emit('map', ZONE_POLYGON_ROLE, ($event.target as HTMLSelectElement).value)">
          <option value="">Not mapped</option>
          <option
            v-if="zone.entities?.[ZONE_POLYGON_ROLE] && !polygons.some((expose) => expose.name === zone.entities?.[ZONE_POLYGON_ROLE])"
            :value="zone.entities[ZONE_POLYGON_ROLE]"
            disabled>
            Unavailable: {{ zone.entities[ZONE_POLYGON_ROLE] }}
          </option>
          <option v-for="expose in polygons" :key="expose.name" :value="expose.name">{{ exposeDisplayLabel(expose) }}</option>
        </select>
      </label>
      <p class="nh-muted">
        Value to send: <code>{{ polygonValue.error || polygonValue.value }}</code>
      </p>
    </template>

    <p v-if="plan.error" class="nh-muted">{{ plan.error }}</p>
    <ul v-else>
      <li v-for="row in commandRows" :key="row.role">
        {{ row.role }}: request {{ row.request }} · reported {{ row.reported }}
        {{ isPending(device.id, row.entity) ? '(awaiting confirmation)' : '' }}
      </li>
    </ul>
    <UiButton :disabled="!!plan.error || !online || !canControl(device) || busy" @click="apply">
      {{ zone.shape === 'polygon' ? 'Apply polygon to device' : 'Apply rectangle to device' }}
    </UiButton>
    <p v-if="notice" role="status">{{ notice }}</p>
  </section>
</template>

<style scoped>
  label {
    display: grid;
    gap: 0.3rem;
    margin-block: 0.5rem;
  }
  select {
    width: 100%;
    min-width: 0;
  }
  ul {
    padding-left: 1.25rem;
  }
  code {
    overflow-wrap: anywhere;
  }
</style>
