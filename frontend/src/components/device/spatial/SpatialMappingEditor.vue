<script setup lang="ts">
  /*
   * Presentational target mapping editor. Drafts, validation and option
   * selection live in useSpatialView.ts; this component only renders fields.
   */
  import type { Expose } from '@/types/device';
  import { exposeOptionLabel } from '@/domain/exposes';
  import { SPATIAL_ROLES, spatialRoleLabel, type SpatialMapping, type SpatialRole, type SpatialRowOptions } from '@/domain/spatial';

  defineProps<{
    mappings: readonly SpatialMapping[];
    rowOptions: readonly SpatialRowOptions[];
    warning: (mapping: SpatialMapping, index: number) => string;
    unavailableLabel: (id: string) => string;
  }>();
  const emit = defineEmits<{
    (event: 'update', row: number, role: SpatialRole, value: string): void;
  }>();

  function updateMapping(row: number, role: SpatialRole, event: Event) {
    // Some native menus report input before change; own the draft immediately
    // so a render between those events cannot restore the previous selection.
    emit('update', row, role, (event.target as HTMLSelectElement).value);
  }

  function optionLabel(entity: Expose): string {
    return exposeOptionLabel(entity);
  }

  function isUnavailable(mapping: SpatialMapping, options: Expose[], role: SpatialRole): boolean {
    const selected = mapping[role];
    return !!selected && !options.some((entity) => entity.name === selected);
  }
</script>

<template>
  <p class="nh-muted nh-mapping-hint">
    For numbered target labels, each selector lists only the matching field for that target and axis. Leave unused
    targets unmapped. To unmap a target, set X, Y and Active to Not mapped, then save spatial settings.
  </p>
  <fieldset v-for="(mapping, index) in mappings" :key="index" class="nh-spatial-map">
    <legend>Target {{ index + 1 }}</legend>
    <label v-for="role in SPATIAL_ROLES" :key="role">
      {{ spatialRoleLabel(role) }}
      <select
        :value="mapping[role]"
        class="nh-input"
        @input="updateMapping(index, role, $event)"
        @change="updateMapping(index, role, $event)">
        <option value="">Not mapped</option>
        <option v-if="isUnavailable(mapping, rowOptions[index][role], role)" :value="mapping[role]">
          Unavailable or mismatched: {{ unavailableLabel(mapping[role]) }}
        </option>
        <option v-for="entity in rowOptions[index][role]" :key="entity.name" :value="entity.name">{{ optionLabel(entity) }}</option>
      </select>
    </label>
    <p v-if="warning(mapping, index)" role="status" class="nh-spatial-warning">{{ warning(mapping, index) }}</p>
  </fieldset>
</template>

<style scoped>
  .nh-mapping-hint {
    margin: 0.5rem 0 0;
  }
  .nh-spatial-map {
    display: flex;
    flex-wrap: wrap;
    gap: var(--nh-space);
    margin-block: 0.75rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    padding: var(--nh-space);
  }
  .nh-spatial-map label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
  .nh-spatial-warning {
    flex-basis: 100%;
    margin: 0;
  }
</style>
