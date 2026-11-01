<script setup lang="ts">
  /*
   * Room outline, visual items and sensor placement editor. Bounds, field
   * labels and starter geometry live in domain/spatial/workspace.ts.
   */
  import { ref } from 'vue';
  import type { RoomObject, SpatialPoint, SpatialWorkspace } from '@/types/spatial';
  import {
    MAX_SPATIAL_NAME_LENGTH,
    MAX_SPATIAL_OBJECTS,
    ROOM_FURNITURE_NAMES,
    ROOM_OBJECT_FIELDS,
    SENSOR_FIELDS,
    STARTER_ROOM_OUTLINE,
    rectanglePoints,
  } from '@/domain/spatial/workspace';
  import UiButton from '@/components/ui/UiButton.vue';
  import SpatialContextHelp from './SpatialContextHelp.vue';
  import SpatialPointEditor from './SpatialPointEditor.vue';

  defineProps<{ workspace: SpatialWorkspace; selected?: RoomObject }>();
  const emit = defineEmits<{
    update: [patch: Partial<RoomObject>];
    sensor: [key: keyof SpatialWorkspace['sensor'], value: number];
    add: [kind: RoomObject['kind'], name?: string];
    select: [id: string];
    remove: [];
    outline: [points: SpatialPoint[]];
  }>();
  const furniture = ref<string>(ROOM_FURNITURE_NAMES[0]);
  const roomLimitReached = (workspace: SpatialWorkspace) => workspace.objects.length >= MAX_SPATIAL_OBJECTS;
  const starterOutline = () =>
    rectanglePoints(STARTER_ROOM_OUTLINE.x, STARTER_ROOM_OUTLINE.y, STARTER_ROOM_OUTLINE.width, STARTER_ROOM_OUTLINE.height);
</script>
<template>
  <section aria-label="Room builder" class="nh-room-builder">
    <h3>Room builder</h3>
    <SpatialContextHelp topic="room" />
    <UiButton v-if="!workspace.outline.length" @click="emit('outline', starterOutline())">Add room outline</UiButton>
    <template v-else>
      <SpatialPointEditor :points="workspace.outline" label="Room" @update="emit('outline', $event)" />
      <UiButton @click="emit('outline', [])">Remove outline</UiButton>
    </template>
    <div class="nh-room-actions">
      <UiButton :disabled="roomLimitReached(workspace)" @click="emit('add', 'wall')">Add wall</UiButton>
      <UiButton :disabled="roomLimitReached(workspace)" @click="emit('add', 'door')">Add door</UiButton>
    </div>
    <label>
      Furniture type
      <select v-model="furniture" class="nh-input">
        <option v-for="name in ROOM_FURNITURE_NAMES" :key="name">{{ name }}</option>
      </select>
    </label>
    <UiButton :disabled="roomLimitReached(workspace)" @click="emit('add', 'furniture', furniture)">Add furniture</UiButton>
    <label v-if="workspace.objects.length">
      Room item
      <select class="nh-input" :value="selected?.id ?? ''" @change="emit('select', ($event.target as HTMLSelectElement).value)">
        <option value="">Select an item</option>
        <option v-for="item in workspace.objects" :key="item.id" :value="item.id">{{ item.name }} · {{ item.kind }}</option>
      </select>
    </label>
    <template v-if="selected">
      <label>
        Item name
        <input
          class="nh-input"
          :value="selected.name"
          :maxlength="MAX_SPATIAL_NAME_LENGTH"
          @change="emit('update', { name: ($event.target as HTMLInputElement).value })" />
      </label>
      <label v-for="field in ROOM_OBJECT_FIELDS" :key="field.key">
        {{ field.label }}
        <input
          class="nh-input"
          type="number"
          :value="selected[field.key]"
          @change="emit('update', { [field.key]: Number(($event.target as HTMLInputElement).value) })" />
      </label>
      <UiButton variant="danger" @click="emit('remove')">Remove room item</UiButton>
    </template>
    <fieldset>
      <legend>Sensor placement</legend>
      <label v-for="field in SENSOR_FIELDS" :key="field.key">
        {{ field.label }}
        <input
          class="nh-input"
          type="number"
          :value="workspace.sensor[field.key]"
          @change="emit('sensor', field.key, Number(($event.target as HTMLInputElement).value))" />
      </label>
    </fieldset>
  </section>
</template>
<style scoped>
  .nh-room-builder {
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
  fieldset {
    min-width: 0;
    border: 1px solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    display: grid;
    gap: 0.5rem;
  }
  .nh-room-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }
</style>
