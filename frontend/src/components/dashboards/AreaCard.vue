<script setup lang="ts">
  /* One Home area: title, tiles, and (while arranging) move/rename/add/delete tools. */
  import { computed } from 'vue';
  import type { HomeBlock, HomeCard } from '@/types/home.type';
  import type { HomeSourcePresentation } from '@/domain/home-cards';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import HomeCardList from './HomeCardList.vue';
  import HomeBlockHeader from './HomeBlockHeader.vue';

  const props = defineProps<{
    group: HomeBlock;
    sources: ReadonlyMap<string, HomeSourcePresentation>;
    editing: boolean;
    busy: boolean;
    index: number;
    count: number;
    sourceError?: string | null;
  }>();
  defineEmits<{
    (e: 'move', to: number): void;
    (e: 'rename'): void;
    (e: 'remove'): void;
    (e: 'add-tiles'): void;
    (e: 'remove-tile', id: string): void;
    (e: 'move-card', card: HomeCard): void;
    (e: 'transfer-card', placement: { id: string; index: number }): void;
    (e: 'reorder-cards', cards: HomeCard[]): void;
  }>();

  const headingId = computed(() => `area-${props.group.id}`);
</script>

<template>
  <section class="nh-area nh-home-edit-surface" :class="{ 'is-editing': editing }" :aria-labelledby="headingId" :data-area="group.name">
    <HomeBlockHeader class="nh-area-head" :title="group.name" :heading-id="headingId" :editing="editing" :busy="busy"
      :index="index" :count="count" :drag-label="`Drag ${group.name} area`" @move="$emit('move', $event)">
      <UiButton size="sm" variant="ghost" icon="edit" :label="`Rename ${group.name}`"
        :disabled="busy" @click="$emit('rename')" />
      <UiButton size="sm" variant="ghost" icon="delete" :label="`Delete ${group.name}`"
        :disabled="busy" class="is-danger" @click="$emit('remove')" />
    </HomeBlockHeader>
    <HomeCardList
      :cards="group.cards"
      :sources="sources"
      :editing="editing"
      :busy="busy"
      :source-error="sourceError"
      accepts-cards
      @remove="$emit('remove-tile', $event)"
      @move="$emit('move-card', $event)"
      @transfer="$emit('transfer-card', $event)"
      @reorder="$emit('reorder-cards', $event)">
      <button v-if="editing" type="button" class="nh-tile-add" :disabled="busy" @click="$emit('add-tiles')">
        <UiIcon name="add" />
        Add tiles
      </button>
      <p v-else-if="!group.cards.length" class="nh-area-empty">
        No tiles yet. Choose
        <b>Edit layout</b>
        to add some.
      </p>
    </HomeCardList>
  </section>
</template>

<style scoped>
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
