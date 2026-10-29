<script setup lang="ts">
  import type { HomeCard } from '@/types/home.type';
  import type { HomeSourcePresentation } from '@/domain/home-cards';
  import EntityTile from '@/components/entity/EntityTile.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import IntegrationTile from './IntegrationTile.vue';

  defineProps<{
    card: HomeCard;
    sources: ReadonlyMap<string, HomeSourcePresentation>;
    editing?: boolean;
    busy?: boolean;
    sourceError?: string | null;
  }>();
  defineEmits<{ remove: []; move: [] }>();
</script>

<template>
  <div
    class="nh-home-card"
    :class="{ 'is-integration': card.kind !== 'device' }"
    :data-card-id="card.id">
    <EntityTile
      v-if="card.kind === 'device'"
      :device-id="card.device_id"
      :expose="card.expose"
      :editing="editing"
      :busy="busy"
      @remove="$emit('remove')" />
    <IntegrationTile
      v-else
      :card="card"
      :sources="sources"
      :editing="editing"
      :busy="busy"
      :source-error="sourceError"
      @remove="$emit('remove')" />
    <div v-if="editing" class="nh-card-layout-tools">
      <UiButton size="sm" variant="ghost" icon="drag" label="Drag card" class="nh-card-move-handle" :disabled="busy" />
      <UiButton size="sm" variant="ghost" icon="moveUp" label="Move card" :disabled="busy" @click="$emit('move')" />
    </div>
  </div>
</template>

<style scoped>
  .nh-home-card {
    min-width: 0;
    position: relative;
  }
  .nh-home-card:not(.is-integration) {
    display: flex;
  }
  .nh-home-card:not(.is-integration) :deep(.nh-tile) {
    flex: 1;
  }
  .nh-home-card.is-integration {
    grid-column: 1 / -1;
  }
  .nh-card-layout-tools {
    position: absolute;
    z-index: 2;
    right: 0.25rem;
    bottom: 0.25rem;
    display: flex;
    border-radius: var(--nh-radius-s);
    background: var(--nh-surface);
    opacity: 0;
  }
  .nh-home-card:hover .nh-card-layout-tools,
  .nh-home-card:focus-within .nh-card-layout-tools {
    opacity: 1;
  }
  @media (hover: none) {
    .nh-card-layout-tools {
      opacity: 1;
    }
  }
</style>
