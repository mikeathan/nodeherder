<script setup lang="ts">
  import { ref, onUnmounted } from 'vue';
  import { DialogComponents } from '@/mixins/useDialogComponents';
  import { OpenDialogEvent, useDialogEvents, emitCloseDialog } from '@/mixins/useDialogsEventBus';

  const cleanup = useDialogEvents({
    openDialog: (event: OpenDialogEvent) => {
      openDialog(event);
    },

    closeDialog: () => {
      closeDialog();
    },
  });

  function closeDialog() {
    if (currentDialogComponent.value != null) {
      currentDialogComponent.value = null;
    }
  }

  function openDialog(event: OpenDialogEvent): void {
    if (currentDialogComponent.value == null) {
      currentDialogComponent.value = event;
    }
  }

  const currentDialogComponent = ref<OpenDialogEvent | null>(null);

  onUnmounted(() => {
    cleanup();
    currentDialogComponent.value = null;
  });
</script>

<template>
  <component
    v-if="currentDialogComponent != null"
    :is="DialogComponents[currentDialogComponent.type]"
    v-bind="currentDialogComponent.props"
    v-on="currentDialogComponent.events" />
</template>
