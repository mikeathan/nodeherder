<script setup lang="ts">
  import { computed, ref, watch } from 'vue';
  import type { HomeCard } from '@/types/home.type';
  import type { HomeSourcePresentation } from '@/domain/home-cards';
  import { useHub } from '@/composables/useHub';
  import {
    homePickerChoices,
    homePickerOptions,
    homePickerEmptyMessage,
    selectedHomeCards,
  } from '@/domain/home-picker';
  import UiDialog from '@/components/ui/UiDialog.vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiSelect from '@/components/ui/UiSelect.vue';

  const props = defineProps<{
    open: boolean;
    areaName: string;
    existing: readonly HomeCard[];
    sources: ReadonlyMap<string, HomeSourcePresentation>;
    busy: boolean;
    sourceError?: string | null;
  }>();
  const emit = defineEmits<{ close: []; add: [cards: HomeCard[]] }>();
  const { devices } = useHub();
  const source = ref<string | null>(null);
  const picked = ref<string[]>([]);
  const options = computed(() => homePickerOptions(devices.value));
  const choices = computed(() => homePickerChoices(source.value, devices.value, props.sources, props.existing));
  const emptyMessage = computed(() => homePickerEmptyMessage(source.value));
  const addLabel = computed(() => {
    if (picked.value.length === 0) return 'Add tiles';
    if (picked.value.length === 1) return 'Add 1 tile';
    return `Add ${picked.value.length} tiles`;
  });

  function resetSelection() {
    source.value = null;
    picked.value = [];
  }
  function clearPickedValues() {
    picked.value = [];
  }
  function add() {
    emit('add', selectedHomeCards(choices.value, picked.value));
  }

  watch(
    () => props.open,
    (open) => {
      if (open) resetSelection();
    }
  );
  watch(source, clearPickedValues);
</script>

<template>
  <UiDialog :open="open" :title="`Add tiles to ${areaName || 'Home'}`" wide @close="emit('close')">
    <label class="nh-label" for="home-tile-source">Device or integration</label>
    <UiSelect
      id="home-tile-source"
      v-model="source"
      :options="options"
      placeholder="Choose a source"
      :disabled="busy" />
    <p v-if="source === 'integrations' && sourceError" class="nh-alert is-warning" role="status">{{ sourceError }}</p>
    <fieldset v-if="choices.length" class="nh-checks-wrap">
      <legend class="nh-label">Values to show</legend>
      <div class="nh-checks">
        <label v-for="choice in choices" :key="choice.key" class="nh-check">
          <input v-model="picked" type="checkbox" :value="choice.key" :disabled="busy || choice.alreadyShown" />
          <span>
            {{ choice.label }}
            <small>{{ choice.alreadyShown ? 'Already shown here' : choice.detail }}</small>
          </span>
        </label>
      </div>
    </fieldset>
    <p v-else-if="source" class="nh-muted">{{ emptyMessage }}</p>
    <template #footer>
      <UiButton :disabled="busy" @click="emit('close')">Cancel</UiButton>
      <UiButton variant="primary" :disabled="busy || !picked.length" @click="add">{{ addLabel }}</UiButton>
    </template>
  </UiDialog>
</template>

<style scoped>
  .nh-checks-wrap {
    border: 0;
    padding: 0;
    margin: 1rem 0 0;
    min-width: 0;
  }
  .nh-checks {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(12rem, 1fr));
    gap: 0.15rem 1rem;
    margin-top: 0.5rem;
  }
  .nh-check {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-height: 2.75rem;
    cursor: pointer;
  }
  .nh-check input {
    accent-color: var(--nh-accent);
    width: 1.1rem;
    height: 1.1rem;
  }
  .nh-check small {
    display: block;
    font-size: 0.75rem;
    color: var(--nh-text-muted);
  }
</style>
