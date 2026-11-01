<script setup lang="ts">
  /*
   * Zone list and numeric geometry editor. Bounds/resize math and slot limits
   * live in domain/spatial/workspace.ts; this component emits requested edits.
   */
  import { computed, ref } from 'vue';
  import type { SpatialPoint, SpatialZone, ZoneKind, ZoneShape } from '@/types/spatial';
  import { ZONE_KINDS } from '@/types/spatial';
  import {
    MAX_SPATIAL_NAME_LENGTH,
    MIN_SPATIAL_DIMENSION_MM,
    SPATIAL_COORDINATE_LIMIT_MM,
    ZONE_BOUND_FIELDS,
    ZONE_LIMITS,
    polygonBounds,
    resizeZoneBounds,
    zoneKindLabel,
  } from '@/domain/spatial/workspace';
  import UiButton from '@/components/ui/UiButton.vue';
  import SpatialContextHelp from './SpatialContextHelp.vue';
  import SpatialPointEditor from './SpatialPointEditor.vue';

  const props = defineProps<{ zones: SpatialZone[]; selected?: SpatialZone }>();
  const emit = defineEmits<{
    select: [id: string];
    add: [kind: ZoneKind, shape: ZoneShape];
    update: [patch: Partial<SpatialZone>];
    remove: [];
  }>();
  const shape = ref<ZoneShape>('rectangle');
  const bounds = computed(() => polygonBounds(props.selected?.points ?? [{ x: 0, y: 0 }]));

  /** Applies one numeric rectangle bound through the shared geometry helper. */
  function resize(key: 'x' | 'y' | 'width' | 'height', event: Event) {
    if (!props.selected) return;
    emit('update', { points: resizeZoneBounds(props.selected.points, key, Number((event.target as HTMLInputElement).value)) });
  }

  /** Replaces polygon corners from the point editor. */
  function updatePoints(points: SpatialPoint[]) {
    emit('update', { points });
  }
</script>
<template>
  <section aria-label="Zone editor" class="nh-zone-editor">
    <h3>Zone editor</h3>
    <p class="nh-muted">Create room overlays. Geometric previews are separate from the sensor’s reported occupancy.</p>
    <SpatialContextHelp topic="zones" />
    <label>
      New zone shape
      <select v-model="shape" class="nh-input">
        <option value="rectangle">Rectangle</option>
        <option value="polygon">Polygon</option>
      </select>
    </label>
    <div class="nh-editor-actions">
      <UiButton
        v-for="kind in ZONE_KINDS"
        :key="kind"
        :disabled="zones.filter((zone) => zone.kind === kind).length >= ZONE_LIMITS[kind]"
        @click="emit('add', kind, shape)">
        Add {{ zoneKindLabel(kind).toLowerCase() }} zone
      </UiButton>
    </div>
    <p v-if="!zones.length">Create your first zone: add a detection zone, name it, then position it on the canvas.</p>
    <label v-else>
      Zone slots
      <select
        class="nh-input"
        :value="selected?.id ?? ''"
        @change="emit('select', ($event.target as HTMLSelectElement).value)">
        <option value="">Select a zone</option>
        <option v-for="zone in zones" :key="zone.id" :value="zone.id">{{ zone.name }} · {{ zone.kind }}</option>
      </select>
    </label>
    <div v-if="selected" class="nh-zone-fields">
      <label>
        Zone name
        <input
          :value="selected.name"
          class="nh-input"
          :maxlength="MAX_SPATIAL_NAME_LENGTH"
          @change="emit('update', { name: ($event.target as HTMLInputElement).value })" />
      </label>
      <template v-if="selected.shape === 'rectangle'">
        <label v-for="field in ZONE_BOUND_FIELDS" :key="field.key">
          {{ field.label }} (mm)
          <input
            class="nh-input"
            type="number"
            :min="field.key === 'width' || field.key === 'height' ? MIN_SPATIAL_DIMENSION_MM : -SPATIAL_COORDINATE_LIMIT_MM"
            :max="SPATIAL_COORDINATE_LIMIT_MM"
            :value="bounds[field.key]"
            @change="resize(field.key, $event)" />
        </label>
      </template>
      <SpatialPointEditor v-else :points="selected.points" label="Zone" @update="updatePoints" />
      <UiButton variant="danger" @click="emit('remove')">Remove zone</UiButton>
      <p class="nh-muted">
        Drag the selected shape or its corner handles. Numeric fields provide the same keyboard controls. Use Undo to
        restore a removed overlay.
      </p>
    </div>
  </section>
</template>
<style scoped>
  .nh-zone-editor,
  .nh-zone-fields {
    display: grid;
    gap: 0.65rem;
    min-width: 0;
  }
  label {
    display: grid;
    gap: 0.25rem;
  }
  input,
  select {
    min-width: 0;
    width: 100%;
  }
  .nh-editor-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }
</style>
