<script setup lang="ts">
  import { computed } from 'vue';
  import type { HomeBlock, HomeCard } from '@/types/home.type';
  import { homeCardTitle, type HomeSourcePresentation } from '@/domain/home-cards';
  import HomeCardList from './HomeCardList.vue';
  import HomeBlockHeader from './HomeBlockHeader.vue';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{
    block: HomeBlock;
    sources: ReadonlyMap<string, HomeSourcePresentation>;
    editing: boolean;
    busy: boolean;
    index: number;
    count: number;
    sourceError?: string | null;
  }>();
  defineEmits<{ remove: [id: string]; deleteBlock: [card: HomeCard, title: string]; moveCard: [card: HomeCard]; moveBlock: [to: number] }>();
  const title = computed(() => homeCardTitle(props.block.cards[0], props.sources));
</script>

<template>
  <div class="nh-home-edit-surface" :class="{ 'is-editing': editing }">
    <HomeBlockHeader v-if="editing" :title="title" :editing="editing" :busy="busy" :index="index"
      :count="count" :drag-label="`Drag ${title} card`" @move="$emit('moveBlock', $event)">
      <UiButton size="sm" variant="ghost" icon="delete" :label="`Delete ${title}`" :disabled="busy"
        class="is-danger" @click="$emit('deleteBlock', block.cards[0], title)" />
    </HomeBlockHeader>
    <HomeCardList
      :cards="block.cards"
      :sources="sources"
      :source-error="sourceError"
      :editing="editing"
      :busy="busy"
      @remove="$emit('remove', $event)"
      @move="$emit('moveCard', $event)" />
  </div>
</template>
