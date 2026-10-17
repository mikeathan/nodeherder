<script setup lang="ts">
  import { ref } from 'vue';
  import UiButton from '@/components/ui/UiButton.vue';
  defineProps<{ label: string; options: { name: string; label: string }[]; manual: boolean; disabled: boolean }>();
  const emit = defineEmits<{ (event: 'add', name: string): void }>();
  const name = ref('');
  function add() {
    emit('add', name.value.trim());
    name.value = '';
  }
</script>

<template>
  <div class="sensor-picker">
    <label>
      <span>{{ label }}</span>
      <input v-if="manual" v-model="name" :aria-label="label" placeholder="e.g. target_distance" :disabled="disabled" />
      <select v-else v-model="name" :aria-label="label" :disabled="disabled">
        <option value="">Choose a sensor</option>
        <option v-for="option in options" :key="option.name" :value="option.name">{{ option.label }}</option>
      </select>
    </label>
    <UiButton :disabled="disabled || !name.trim()" @click="add">Add adjustment</UiButton>
  </div>
</template>

<style scoped>
  .sensor-picker { display: flex; align-items: end; gap: .75rem; margin-top: 1rem; }
  label { display: grid; gap: .4rem; flex: 1; min-width: 0; font-size: .85rem; }
  input, select { width: 100%; min-height: 2.75rem; padding: .65rem; border: 1px solid var(--nh-border); border-radius: var(--nh-radius-s); background: var(--nh-surface); color: var(--nh-text); font: inherit; }
  @media (max-width: 640px) { .sensor-picker { align-items: stretch; flex-direction: column; } }
</style>
