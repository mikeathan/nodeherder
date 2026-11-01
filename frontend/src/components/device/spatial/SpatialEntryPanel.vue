<script setup lang="ts">
  /*
   * Entry / assumed-presence mapping and inspected reported values. Field
   * definitions and option filtering live in domain/spatial/entry.ts.
   */
  import { computed } from 'vue';
  import type { Device } from '@/types/device';
  import type { EntryRole, SpatialWorkspace } from '@/types/spatial';
  import { ENTRY_FIELD_DEFINITIONS, entryEntityOptions } from '@/domain/spatial/entry';
  import { exposeDisplayLabel } from '@/domain/exposes';
  import ExposeRow from '@/components/entity/ExposeRow.vue';
  import SpatialContextHelp from './SpatialContextHelp.vue';

  const props = defineProps<{ device: Device; workspace: SpatialWorkspace; online: boolean }>();
  const emit = defineEmits<{ map: [role: EntryRole, id: string] }>();
  const options = computed(() =>
    ENTRY_FIELD_DEFINITIONS.map((field) => entryEntityOptions(props.device, field))
  );

  /** Reads the selected value from a mapping select. */
  function map(role: EntryRole, event: Event) {
    emit('map', role, (event.target as HTMLSelectElement).value);
  }
</script>
<template>
  <section aria-label="Entry and assumed presence">
    <h3>Entry / Exit & Assumed Presence</h3>
    <SpatialContextHelp topic="entry" />
    <p>
      Map the device’s discovered entities to inspect and tune its own occupancy hold. Entry overlays alone do not
      enable this feature.
    </p>
    <p v-if="!online" role="status">Device offline. Reported values below may be stale.</p>
    <details>
      <summary>Configure entry entity mapping</summary>
      <label v-for="(item, index) in ENTRY_FIELD_DEFINITIONS" :key="item.role" class="nh-entry-field">
        {{ item.label }} entity
        <select class="nh-input" :value="workspace.entryEntities[item.role]" @change="map(item.role, $event)">
          <option value="">Not mapped</option>
          <option
            v-if="
              workspace.entryEntities[item.role] &&
              !options[index].some((expose) => expose.name === workspace.entryEntities[item.role])
            "
            :value="workspace.entryEntities[item.role]"
            disabled>
            Unavailable: {{ workspace.entryEntities[item.role] }}
          </option>
          <option v-for="expose in options[index]" :key="expose.name" :value="expose.name">{{ exposeDisplayLabel(expose) }}</option>
        </select>
      </label>
    </details>
    <template v-for="item in ENTRY_FIELD_DEFINITIONS" :key="item.role">
      <div v-if="device.exposes[workspace.entryEntities[item.role]]">
        <h4>{{ item.label }}</h4>
        <ExposeRow :device="device" :expose="device.exposes[workspace.entryEntities[item.role]]" />
      </div>
      <p v-else class="nh-muted">{{ item.label }}: not mapped or unavailable.</p>
    </template>
    <p class="nh-muted">
      Save spatial settings to retain mappings. Device controls send changes immediately and wait for reported
      confirmation.
    </p>
  </section>
</template>
<style scoped>
  .nh-entry-field {
    display: grid;
    gap: 0.3rem;
    margin-block: 0.75rem;
  }
  select {
    min-width: 0;
    max-width: 100%;
  }
</style>
