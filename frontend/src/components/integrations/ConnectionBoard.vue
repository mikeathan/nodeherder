<script setup lang="ts">
  // Ordering interactions are shared; each widget remains responsible only for its presentation.
  import { nextTick, ref } from 'vue';
  import draggable from 'vuedraggable';
  import type { ConnectionBoardEntry, ConnectionInstance } from '@/types/integrations.type';
  import ConnectionWidget from './ConnectionWidget.vue';
  const props = defineProps<{
    entries: ConnectionBoardEntry[];
    instances: Map<string, ConnectionInstance>;
    editing: boolean;
    busy: boolean;
  }>();
  const emit = defineEmits<{
    reorder: [entries: ConnectionBoardEntry[]];
    move: [id: string, offset: number];
    hide: [id: string];
    refresh: [item: ConnectionInstance];
    open: [id: string, event: MouseEvent];
  }>();
  const root = ref<HTMLElement | null>(null);
  async function move(id: string, offset: number) {
    emit('move', id, offset);
    await nextTick();
    const widget = Array.from(root.value?.querySelectorAll<HTMLElement>('[data-instance-id]') ?? []).find(
      (item) => item.dataset.instanceId === id
    );
    // At an edge the just-used arrow becomes disabled; focus the remaining usable arrow.
    (
      widget?.querySelector<HTMLButtonElement>(`[data-move="${offset}"]:not(:disabled)`) ??
      widget?.querySelector<HTMLButtonElement>('[data-move]:not(:disabled)')
    )?.focus();
  }
</script>
<template>
  <div ref="root">
    <draggable
      :model-value="entries"
      item-key="instance_id"
      tag="div"
      class="connection-board-grid"
      :disabled="!editing || busy"
      handle=".connection-widget-drag"
      :animation="150"
      @update:model-value="emit('reorder', $event)">
      <template #item="{ element, index }">
        <ConnectionWidget
          v-if="props.instances.has(element.instance_id)"
          :item="props.instances.get(element.instance_id)!"
          :editing="editing"
          :busy="busy"
          :first="index === 0"
          :last="index === entries.length - 1"
          @open="emit('open', element.instance_id, $event)"
          @refresh="emit('refresh', props.instances.get(element.instance_id)!)"
          @hide="emit('hide', element.instance_id)"
          @move="move(element.instance_id, $event)" />
      </template>
    </draggable>
  </div>
</template>
<style scoped>
  /* auto-fill retains empty tracks, so one connection never stretches across the entire page. */
  .connection-board-grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(min(100%, 19rem), 1fr));
    gap: 1rem;
    align-items: start;
  }
  @media (max-width: 640px) {
    .connection-board-grid {
      grid-template-columns: minmax(0, 1fr);
    }
  }
</style>
