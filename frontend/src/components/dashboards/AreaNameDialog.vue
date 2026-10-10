<script setup lang="ts">
  /* Create or rename a Home area, with inline validation (replaces alert() checks). */
  import { ref, watch } from 'vue';
  import UiDialog from '@/components/ui/UiDialog.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiField from '@/components/ui/UiField.vue';

  const props = defineProps<{ open: boolean; current?: string; validate: (name: string, current?: string) => string | null }>();
  const emit = defineEmits<{ (e: 'close'): void; (e: 'save', name: string): void }>();

  const name = ref('');
  const error = ref<string | null>(null);
  watch(
    () => props.open,
    (open) => {
      if (!open) return;
      name.value = props.current ?? '';
      error.value = null;
    }
  );

  function submit() {
    error.value = props.validate(name.value, props.current);
    if (error.value) return;
    emit('save', name.value.trim());
  }
</script>

<template>
  <UiDialog :open="open" :title="current ? 'Rename area' : 'New area'" @close="emit('close')">
    <form id="area-name-form" @submit.prevent="submit">
      <UiField label="Area name" for="area-name" :error="error" help="For example Kitchen or Upstairs.">
        <input
          id="area-name"
          v-model="name"
          class="nh-input"
          autocomplete="off"
          maxlength="80"
          autofocus
          :aria-invalid="!!error || undefined"
          :aria-describedby="error ? 'area-name-err' : 'area-name-help'" />
      </UiField>
    </form>
    <template #footer>
      <UiButton @click="emit('close')">Cancel</UiButton>
      <UiButton variant="primary" type="submit" form="area-name-form">{{ current ? 'Rename' : 'Create area' }}</UiButton>
    </template>
  </UiDialog>
</template>
