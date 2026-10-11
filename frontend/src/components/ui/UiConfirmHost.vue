<script setup lang="ts">
  /* Renders confirm() requests (composables/useConfirm.ts). Mounted once in the shell. */
  import UiDialog from './UiDialog.vue';
  import UiButton from './UiButton.vue';
  import { confirmState, settleConfirm } from '@/composables/useConfirm';
</script>

<template>
  <UiDialog :open="!!confirmState" :title="confirmState?.title ?? ''" @close="settleConfirm(false)">
    <p>{{ confirmState?.message }}</p>
    <template #footer>
      <UiButton @click="settleConfirm(false)">Cancel</UiButton>
      <UiButton :variant="confirmState?.danger ? 'danger' : 'primary'" autofocus @click="settleConfirm(true)">
        {{ confirmState?.confirmLabel ?? 'Confirm' }}
      </UiButton>
    </template>
  </UiDialog>
</template>
