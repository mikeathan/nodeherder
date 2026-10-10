<script setup lang="ts">
  /* Slide-over customiser opened from the header palette button. */
  import UiDialog from '@/components/ui/UiDialog.vue';
  import AppearanceForm from './AppearanceForm.vue';

  defineProps<{ open: boolean }>();
  defineEmits<{ (e: 'close'): void }>();
</script>

<template>
  <UiDialog :open="open" title="Appearance" @close="$emit('close')">
    <p class="nh-muted">Saved in this browser only.</p>
    <AppearanceForm />
  </UiDialog>
</template>
