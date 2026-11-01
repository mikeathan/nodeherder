<script setup lang="ts">
  /*
   * Numeric corner editor for polygon overlays. Bounds and append defaults come
   * from domain/spatial/workspace.ts; edits are emitted without mutation.
   */
  import type { SpatialPoint } from '@/types/spatial';
  import {
    MAX_SPATIAL_VERTICES,
    MIN_SPATIAL_VERTICES,
    SPATIAL_COORDINATE_LIMIT_MM,
    appendSpatialPoint,
  } from '@/domain/spatial/workspace';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{ points: SpatialPoint[]; label: string }>();
  const emit = defineEmits<{ update: [points: SpatialPoint[]] }>();

  /** Updates one axis of one corner and emits a fresh list. */
  function update(index: number, axis: 'x' | 'y', event: Event) {
    const points = props.points.map((point) => ({ ...point }));
    points[index][axis] = Number((event.target as HTMLInputElement).value);
    emit('update', points);
  }

  /** Removes one corner while preserving the minimum polygon size. */
  function remove(index: number) {
    emit('update', props.points.filter((_, current) => current !== index));
  }

  /** Appends one starter corner while preserving the maximum polygon size. */
  function add() {
    emit('update', appendSpatialPoint(props.points));
  }
</script>
<template>
  <fieldset>
    <legend>{{ label }} corners (mm)</legend>
    <div v-for="(point, index) in points" :key="index" class="nh-point-row">
      <span>{{ index + 1 }}</span>
      <label>
        X
        <input
          class="nh-input"
          type="number"
          :min="-SPATIAL_COORDINATE_LIMIT_MM"
          :max="SPATIAL_COORDINATE_LIMIT_MM"
          :aria-label="`${label} corner ${index + 1} X`"
          :value="point.x"
          @change="update(index, 'x', $event)" />
      </label>
      <label>
        Y
        <input
          class="nh-input"
          type="number"
          :min="-SPATIAL_COORDINATE_LIMIT_MM"
          :max="SPATIAL_COORDINATE_LIMIT_MM"
          :aria-label="`${label} corner ${index + 1} Y`"
          :value="point.y"
          @change="update(index, 'y', $event)" />
      </label>
      <UiButton
        size="sm"
        :disabled="points.length <= MIN_SPATIAL_VERTICES"
        :label="`Remove ${label} corner ${index + 1}`"
        @click="remove(index)">
        Remove
      </UiButton>
    </div>
    <UiButton :disabled="points.length >= MAX_SPATIAL_VERTICES" @click="add">Add corner</UiButton>
  </fieldset>
</template>
<style scoped>
  fieldset {
    min-width: 0;
    border: 1px solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    padding: 0.65rem;
  }
  .nh-point-row {
    display: grid;
    grid-template-columns: auto minmax(0, 1fr) minmax(0, 1fr) auto;
    gap: 0.4rem;
    align-items: end;
    margin-bottom: 0.5rem;
  }
  input {
    width: 100%;
    min-width: 0;
  }
  label {
    display: grid;
    gap: 0.2rem;
  }
</style>
