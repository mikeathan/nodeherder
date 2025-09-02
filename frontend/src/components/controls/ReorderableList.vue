<script setup lang="ts">
  import { defineProps, defineEmits, reactive, watch } from 'vue';
  import draggable from 'vuedraggable';

  interface Props<T> {
    items: T[];
    itemKey: string;
    showHandle?: boolean;
    showBorder?: boolean;
  }

  const props = defineProps<Props<any>>();
  const emit = defineEmits<{
    (e: 'update:items', value: any[]): void;
    (e: 'delete-item', item: any): void;
  }>();
  const localItems = reactive([...props.items]);

  watch(
    () => props.items,
    (newItems) => {
      localItems.splice(0, localItems.length, ...newItems);
    }
  );

  function onReorder() {
    emit('update:items', [...localItems]);
  }
</script>

<template>
  <draggable
    v-model="localItems"
    :item-key="props.itemKey"
    :handle="props.showHandle ? '.drag-handle' : undefined"
    tag="div"
    @end="onReorder">
    <template #item="{ element, index }">
      <div class="flex flex-col sm:flex-row sm:items-center" :class="{ 'item-border': props.showBorder }">
        <!-- Slot content -->
        <div class="flex-1 sm:flex sm:items-center">
          <slot :item="element" :index="index" />
        </div>
        <!-- Action Panel  -->
        <div class="mt-3">
          <!-- Drag Handle -->
          <i class="p-1 drag-handle pi pi-bars cursor-grab" style="color: var(--p-button-text-primary-color)" />
          <!-- Delete Button -->
          <i
            @click="$emit('delete-item', element)"
            class="p-1 pi pi-trash cursor-pointer"
            style="color: var(--p-button-text-primary-color)" />
        </div>
      </div>
    </template>
  </draggable>
</template>

<style scoped>
  .item-border {
    border: 1px solid var(--p-button-secondary-border-color, transparent);
  }
  .drag-handle {
    cursor: grab;
  }
  .click-handle {
    cursor: pointer;
  }
</style>
