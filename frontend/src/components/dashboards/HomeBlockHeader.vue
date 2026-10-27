<script setup lang="ts">
  import UiButton from '@/components/ui/UiButton.vue';

  defineProps<{
    title: string;
    headingId?: string;
    editing: boolean;
    busy: boolean;
    index: number;
    count: number;
    dragLabel: string;
  }>();
  defineEmits<{ move: [to: number] }>();
</script>

<template>
  <header class="nh-block-head">
    <UiButton v-if="editing" size="sm" variant="ghost" icon="drag" :label="dragLabel"
      class="nh-area-handle" :disabled="busy" />
    <h2 :id="headingId">{{ title }}</h2>
    <div v-if="editing" class="nh-block-tools">
      <UiButton size="sm" variant="ghost" icon="moveUp" :label="`Move ${title} earlier`"
        :disabled="busy || index === 0" @click="$emit('move', index - 1)" />
      <UiButton size="sm" variant="ghost" icon="moveDown" :label="`Move ${title} later`"
        :disabled="busy || index === count - 1" @click="$emit('move', index + 1)" />
      <slot />
    </div>
  </header>
</template>

<style scoped>
  .nh-block-head {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    margin: 0 0 var(--nh-home-heading-gap, 0.6rem) 0.15rem;
    min-height: var(--nh-home-heading-height, 2rem);
  }
  .nh-block-head h2 {
    color: var(--nh-text);
    font-size: 1.05rem;
    font-weight: 650;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .nh-area-handle:not(:disabled) {
    color: var(--nh-text-muted);
    cursor: grab;
    touch-action: none;
    flex: none;
  }
  .nh-area-handle:not(:disabled):active { cursor: grabbing; }
  .nh-block-tools { margin-left: auto; display: flex; gap: 0.1rem; }
</style>
