<script setup lang="ts">
  /* Presentational range and rotation controls for the spatial view. */
  import { SPATIAL_ORIENTATION_OPTIONS, SPATIAL_SCALE_OPTIONS } from '@/domain/spatial';

  defineProps<{ scale: number; orientation: number }>();
  const emit = defineEmits<{ (event: 'update:scale', value: number): void; (event: 'update:orientation', value: number): void }>();

  function selectScale(event: Event) {
    emit('update:scale', Number((event.target as HTMLSelectElement).value));
  }
  function selectOrientation(event: Event) {
    emit('update:orientation', Number((event.target as HTMLSelectElement).value));
  }
</script>

<template>
  <div class="nh-spatial-options">
    <label>
      Scale
      <select :value="scale" class="nh-input" @change="selectScale">
        <option v-for="option in SPATIAL_SCALE_OPTIONS" :key="option.value" :value="option.value">{{ option.label }}</option>
      </select>
    </label>
    <label>
      Orientation
      <select :value="orientation" class="nh-input" @change="selectOrientation">
        <option v-for="option in SPATIAL_ORIENTATION_OPTIONS" :key="option.value" :value="option.value">{{ option.label }}</option>
      </select>
    </label>
  </div>
</template>

<style scoped>
  .nh-spatial-options {
    display: flex;
    flex-wrap: wrap;
    gap: var(--nh-space);
    margin-block: 0.75rem;
  }
  .nh-spatial-options label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }
</style>
