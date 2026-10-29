<script setup lang="ts">
  import { ref, watch } from 'vue';
  import draggable from 'vuedraggable';
  import type { HomeCard } from '@/types/home.type';
  import type { HomeSourcePresentation } from '@/domain/home-cards';
  import HomeCardItem from './HomeCardItem.vue';

  const props = defineProps<{
    cards: readonly HomeCard[];
    sources: ReadonlyMap<string, HomeSourcePresentation>;
    editing: boolean;
    busy: boolean;
    acceptsCards?: boolean;
    sourceError?: string | null;
  }>();
  const emit = defineEmits<{
    remove: [id: string];
    move: [card: HomeCard];
    transfer: [placement: { id: string; index: number }];
    reorder: [cards: HomeCard[]];
  }>();
  const list = ref<HomeCard[]>([]);
  // vuedraggable mutates only an owned interaction array. Server-accepted cards remain untouched.
  // Resync after failed writes too: accepted props may retain exactly the same identity on failure.
  watch(
    [() => props.cards, () => props.busy],
    () => {
      list.value = [...props.cards];
    },
    { immediate: true }
  );
  function changed(event: { added?: { element: HomeCard; newIndex: number }; moved?: { element: HomeCard } }) {
    if (event.added) emit('transfer', { id: event.added.element.id, index: event.added.newIndex });
    else if (event.moved) emit('reorder', [...list.value]);
    // The destination owns a cross-group transaction. Never save the source's transient removal.
  }
</script>

<template>
  <!-- Fallback dragging keeps the overlaid handle active across nested area grids. -->
  <draggable
    v-model="list"
    item-key="id"
    class="nh-tiles"
    :group="{ name: 'home-cards', pull: true, put: !!acceptsCards }"
    handle=".nh-card-move-handle"
    :force-fallback="true"
    :disabled="!editing || busy"
    :animation="160"
    ghost-class="is-ghost"
    @change="changed">
    <template #item="{ element }">
      <HomeCardItem
        :card="element"
        :sources="sources"
        :editing="editing"
        :busy="busy"
        :source-error="sourceError"
        @remove="emit('remove', element.id)"
        @move="emit('move', element)" />
    </template>
    <template #footer><slot /></template>
  </draggable>
</template>

<style scoped>
  .nh-tiles {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(min(100%, 11rem), 1fr));
    gap: var(--nh-gap);
    min-width: 0;
    min-height: 0.5rem;
  }
  .nh-tiles :deep(.nh-home-card.is-ghost) {
    opacity: 0.4;
  }
  .nh-tiles :deep(.is-integration) {
    grid-column: 1 / -1;
  }
</style>
