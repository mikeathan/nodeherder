<script setup lang="ts">
  /* One Home area: title, tiles, and (while arranging) move/rename/add/delete tools. */
  import { computed } from 'vue';
  import { DashboardGroup } from '@/types/settings.type';
  import { groupEntities } from '@/domain/dashboard';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import EntityTile from '@/components/entity/EntityTile.vue';

  const props = defineProps<{ group: DashboardGroup; editing: boolean; index: number; count: number }>();
  defineEmits<{
    (e: 'move', to: number): void;
    (e: 'rename'): void;
    (e: 'remove'): void;
    (e: 'add-tiles'): void;
    (e: 'remove-tile', deviceId: string, expose: string): void;
  }>();

  const tiles = computed(() => groupEntities(props.group));
  const headingId = computed(() => `area-${props.index}`);
</script>

<template>
  <section class="nh-area" :class="{ 'is-editing': editing }" :aria-labelledby="headingId" :data-area="group.name">
    <header class="nh-area-head">
      <span v-if="editing" class="nh-area-handle" title="Drag to move" aria-hidden="true"><UiIcon name="drag" /></span>
      <h2 :id="headingId">{{ group.name }}</h2>
      <span class="nh-area-count">{{ tiles.length }}</span>
      <div v-if="editing" class="nh-area-tools">
        <UiButton size="sm" variant="ghost" icon="moveUp" :label="`Move ${group.name} earlier`" :disabled="index === 0" @click="$emit('move', index - 1)" />
        <UiButton size="sm" variant="ghost" icon="moveDown" :label="`Move ${group.name} later`" :disabled="index === count - 1" @click="$emit('move', index + 1)" />
        <UiButton size="sm" variant="ghost" icon="edit" :label="`Rename ${group.name}`" @click="$emit('rename')" />
        <UiButton size="sm" variant="ghost" icon="delete" :label="`Delete ${group.name}`" class="is-danger" @click="$emit('remove')" />
      </div>
    </header>
    <div class="nh-tiles">
      <EntityTile
        v-for="t in tiles"
        :key="`${t.deviceId}|${t.expose}`"
        :device-id="t.deviceId"
        :expose="t.expose"
        :editing="editing"
        @remove="$emit('remove-tile', t.deviceId, t.expose)" />
      <button v-if="editing" type="button" class="nh-tile-add" @click="$emit('add-tiles')">
        <UiIcon name="add" /> Add tiles
      </button>
      <p v-else-if="!tiles.length" class="nh-area-empty">No tiles yet. Choose <b>Edit layout</b> to add some.</p>
    </div>
  </section>
</template>

<style scoped>
  .nh-area {
    min-width: 0;
    border-radius: var(--nh-radius-l);
    transition:
      outline-color var(--nh-motion),
      background var(--nh-motion);
  }
  .nh-area.is-editing {
    outline: 2px dashed var(--nh-border-strong);
    outline-offset: 0.35rem;
    background: color-mix(in srgb, var(--nh-surface) 40%, transparent);
  }
  .nh-area-head {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0 0 0.6rem 0.15rem;
    min-height: 2rem;
  }
  .nh-area-head h2 {
    font-size: 1.05rem;
    font-weight: 650;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .nh-area-count {
    font-size: 0.75rem;
    color: var(--nh-text-muted);
  }
  .nh-area-handle {
    display: grid;
    place-items: center;
    width: 2rem;
    height: 2rem;
    border-radius: var(--nh-radius-s);
    color: var(--nh-text-muted);
    cursor: grab;
    touch-action: none;
    flex: none;
  }
  .nh-area-handle:hover {
    background: var(--nh-surface-2);
    color: var(--nh-text);
  }
  .nh-area-tools {
    margin-left: auto;
    display: flex;
    gap: 0.1rem;
  }
  .nh-tiles {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(min(100%, 9rem), 1fr));
    gap: var(--nh-gap);
  }
  .nh-tile-add {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 0.4rem;
    min-height: var(--nh-tile-h);
    border-radius: var(--nh-radius-l);
    border: 2px dashed var(--nh-border-strong);
    background: transparent;
    color: var(--nh-text-muted);
    cursor: pointer;
    font-weight: 600;
  }
  .nh-tile-add:hover {
    color: var(--nh-accent);
    border-color: var(--nh-accent);
  }
  .nh-area-empty {
    grid-column: 1 / -1;
    margin: 0;
    padding: 1rem;
    border-radius: var(--nh-radius-l);
    border: var(--nh-border-w) dashed var(--nh-border);
    color: var(--nh-text-muted);
    font-size: 0.875rem;
  }
</style>
