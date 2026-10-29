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
    <p v-if="source === 'integrations' && choices.length" class="nh-muted">Choose one result to add as a card.</p>
    <fieldset v-if="choices.length" class="nh-checks-wrap">
      <legend class="nh-label">Values to show</legend>
      <div class="nh-checks" :class="{ 'is-integration': source === 'integrations' }">
        <label
          v-for="choice in choices"
          :key="choice.key"
          class="nh-check"
          :class="{ 'is-selected': source === 'integrations' && picked[0] === choice.key, 'is-disabled': busy || choice.alreadyShown }">
          <input
            v-if="source === 'integrations'"
            type="radio"
            name="integration-result"
            :checked="picked[0] === choice.key"
            :disabled="busy || choice.alreadyShown"
            @change="picked = [choice.key]" />
          <input v-else v-model="picked" type="checkbox" :value="choice.key" :disabled="busy || choice.alreadyShown" />
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
  .nh-checks.is-integration {
    gap: 0.65rem;
  }
  .is-integration .nh-check {
    min-height: 4.5rem;
    padding: 0.75rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface-2);
    transition: border-color var(--nh-motion), background var(--nh-motion);
  }
  .is-integration .nh-check:hover:not(.is-disabled) {
    border-color: var(--nh-border-strong);
  }
  .is-integration .nh-check.is-selected {
    border-color: var(--nh-accent);
    background: var(--nh-accent-soft);
  }
  .is-integration .nh-check:focus-within {
    outline: 2px solid var(--nh-accent);
    outline-offset: 2px;
  }
  .is-integration .nh-check.is-disabled {
    cursor: not-allowed;
    opacity: 0.6;
  }
  .is-integration .nh-check input {
    appearance: none;
    display: grid;
    place-content: center;
    flex: none;
    margin: 0;
    border: 2px solid var(--nh-border-strong);
    border-radius: 50%;
    background: var(--nh-surface);
  }
  .is-integration .nh-check input:checked {
    border-color: var(--nh-accent);
  }
  .is-integration .nh-check input:checked::before {
    content: '';
    width: 0.5rem;
    height: 0.5rem;
    border-radius: 50%;
    background: var(--nh-accent);
  }
  @media (forced-colors: active) {
    .is-integration .nh-check input {
      appearance: auto;
    }
    .is-integration .nh-check input::before {
      display: none;
    }
  }
  .nh-check small {
    display: block;
    font-size: 0.75rem;
    color: var(--nh-text-muted);
  }
</style>
