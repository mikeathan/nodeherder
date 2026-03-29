<script setup lang="ts">
  import { ref, computed } from 'vue';
  import type { AssistantMessage } from '@/types/assistant.type';
  import BaseBubble from './BaseBubble.vue';

  const props = defineProps<{
    message: AssistantMessage;
  }>();

  const emit = defineEmits<{
    (e: 'rerun', content: string): void;
  }>();

  const isEditing = ref(false);
  const editedContent = ref('');

  const hasChanges = computed(
    () => editedContent.value.trim() !== props.message.content && editedContent.value.trim().length > 0
  );

  const startEdit = () => {
    editedContent.value = props.message.content;
    isEditing.value = true;
  };

  const cancelEdit = () => {
    isEditing.value = false;
  };

  const updateAndRun = () => {
    if (hasChanges.value) {
      emit('rerun', editedContent.value.trim());
    }
    isEditing.value = false;
  };

  const rerunMessage = () => {
    emit('rerun', props.message.content);
  };
</script>

<template>
  <BaseBubble alignment="right" :textToCopy="isEditing ? editedContent : message.content">
    <template #content>
      <div
        v-if="!isEditing"
        class="message-content user-bubble p-3 px-4 text-base line-height-3 text-color border-round-3xl w-full"
        style="word-break: break-word; white-space: pre-wrap">
        {{ message.content }}
      </div>

      <div v-else class="w-full flex flex-column gap-2 mt-2">
        <Textarea
          v-model="editedContent"
          autoResize
          rows="3"
          class="w-full text-base p-3 surface-ground border-round-xl border-1 surface-border"
          @keydown.enter.exact.prevent="updateAndRun" />
        <div class="flex justify-content-end gap-2">
          <Button label="Cancel" text size="small" class="text-color-secondary p-button-sm" @click="cancelEdit" />
          <Button
            label="Update"
            size="small"
            class="p-button-sm p-button-rounded"
            :disabled="!hasChanges"
            @click="updateAndRun" />
        </div>
      </div>
    </template>

    <template #actions>
      <Button
        v-if="!isEditing"
        icon="pi pi-pencil"
        class="text-color-secondary p-button-sm p-0 flex align-items-center justify-content-center opacity-70 hover:opacity-100 transition-opacity"
        text
        rounded
        style="width: 28px; height: 28px"
        v-tooltip.bottom="'Edit Message'"
        @click="startEdit" />
      <Button
        v-if="!isEditing"
        icon="pi pi-refresh"
        class="text-color-secondary p-button-sm p-0 flex align-items-center justify-content-center opacity-70 hover:opacity-100 transition-opacity"
        text
        rounded
        style="width: 28px; height: 28px"
        v-tooltip.bottom="'Rerun Message'"
        @click="rerunMessage" />
    </template>
  </BaseBubble>
</template>

<style>
  .user-bubble {
    background-color: rgba(0, 0, 0, 0.06);
  }

  html.dark .user-bubble {
    background-color: rgba(255, 255, 255, 0.08);
  }
</style>
