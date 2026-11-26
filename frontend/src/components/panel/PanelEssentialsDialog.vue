<script setup lang="ts">
  /*
   * Night essentials picker. Options come from the panel projection so a choice always maps
   * to a control the panel can render; the limit is enforced here and in the domain helper.
   */
  import { ref, watch } from 'vue';
  import { PANEL_ESSENTIAL_LIMIT, togglePanelEssential, type PanelEssential } from '@/domain/panel-essentials';
  import type { PanelAmbientLight } from '@/domain/panel-ambient';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiDialog from '@/components/ui/UiDialog.vue';

  const props = defineProps<{
    open: boolean;
    options: readonly (PanelEssential & { room: string })[];
    selected: readonly PanelEssential[];
    ambient: PanelAmbientLight | null;
    ambientOptions: readonly { deviceId: string; expose: string; label: string; room: string }[];
  }>();
  const emit = defineEmits<{
    (e: 'close'): void;
    (e: 'save', value: PanelEssential[]): void;
    (e: 'save-ambient', value: PanelAmbientLight | null): void;
  }>();

  const draft = ref<PanelEssential[]>([]);
  const ambientKey = ref('');
  const threshold = ref(20);
  watch(
    () => props.open,
    (open) => {
      if (!open) return;
      draft.value = [...props.selected];
      ambientKey.value = props.ambient ? `${props.ambient.deviceId}|${props.ambient.expose}` : '';
      threshold.value = props.ambient?.threshold ?? 20;
    },
    { immediate: true }
  );

  const key = (item: PanelEssential) => `${item.deviceId}|${item.expose}`;
  const isSelected = (item: PanelEssential) => draft.value.some((candidate) => key(candidate) === key(item));
  const limitReached = () => draft.value.length >= PANEL_ESSENTIAL_LIMIT;

  function toggle(item: PanelEssential) {
    draft.value = togglePanelEssential(draft.value, { deviceId: item.deviceId, expose: item.expose, label: item.label });
  }
  function save() {
    emit('save', draft.value);
    const option = props.ambientOptions.find((candidate) => `${candidate.deviceId}|${candidate.expose}` === ambientKey.value);
    emit(
      'save-ambient',
      option && Number.isFinite(Number(threshold.value)) && threshold.value >= 0
        ? { deviceId: option.deviceId, expose: option.expose, threshold: Number(threshold.value) }
        : null
    );
    emit('close');
  }
</script>

<template>
  <UiDialog :open="open" title="Night essentials" wide @close="$emit('close')">
    <p class="nh-muted">
      Choose up to {{ PANEL_ESSENTIAL_LIMIT }} controls for the simplified night panel.
      {{ draft.length }} of {{ PANEL_ESSENTIAL_LIMIT }} selected.
    </p>
    <fieldset class="nh-ambient">
      <legend>Ambient light (optional)</legend>
      <p class="nh-muted">Go to night when this sensor reads below the threshold; the schedule still applies.</p>
      <div class="nh-ambient-row">
        <label>
          <span>Light sensor</span>
          <select v-model="ambientKey">
            <option value="">Off — use schedule only</option>
            <option v-for="option in ambientOptions" :key="`${option.deviceId}|${option.expose}`" :value="`${option.deviceId}|${option.expose}`">
              {{ option.label }} · {{ option.room }}
            </option>
          </select>
        </label>
        <label>
          <span>Below (lx)</span>
          <input v-model.number="threshold" type="number" min="0" max="100000" step="1" :disabled="!ambientKey" />
        </label>
      </div>
    </fieldset>

    <ul class="nh-essential-list">
      <li v-for="item in options" :key="key(item)">
        <label :class="{ 'is-selected': isSelected(item) }">
          <input
            type="checkbox"
            :checked="isSelected(item)"
            :disabled="!isSelected(item) && limitReached()"
            @change="toggle(item)" />
          <span class="nh-essential-tx">
            <b>{{ item.label }}</b>
            <small>{{ item.room }}</small>
          </span>
        </label>
      </li>
    </ul>
    <template #footer>
      <UiButton @click="$emit('close')">Cancel</UiButton>
      <UiButton variant="primary" :disabled="!draft.length" @click="save">Save essentials</UiButton>
    </template>
  </UiDialog>
</template>

<style scoped>
  .nh-ambient {
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    padding: 0.65rem 0.75rem 0.75rem;
    margin: 0.8rem 0 0;
  }
  .nh-ambient legend {
    font-size: 0.78rem;
    font-weight: 650;
    padding: 0 0.25rem;
  }
  .nh-ambient p {
    font-size: 0.72rem;
    margin: 0.2rem 0 0.5rem;
  }
  .nh-ambient-row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.6rem;
  }
  .nh-ambient-row label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    font-size: 0.75rem;
    color: var(--nh-text-muted);
    min-width: 0;
    flex: 1 1 12rem;
  }
  .nh-ambient-row select,
  .nh-ambient-row input {
    min-height: var(--nh-ctl-h);
    border: var(--nh-border-w) solid var(--nh-border-strong);
    border-radius: var(--nh-radius-s);
    background: var(--nh-surface);
    color: var(--nh-text);
    font: inherit;
    padding: 0 0.55rem;
  }
  .nh-essential-list {
    list-style: none;
    margin: 0.6rem 0 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    max-height: 55vh;
    overflow: auto;
  }
  label {
    display: flex;
    align-items: center;
    gap: 0.65rem;
    min-height: var(--nh-touch-h);
    padding: 0.45rem 0.6rem;
    border: var(--nh-border-w) solid var(--nh-border);
    border-radius: var(--nh-radius-m);
    cursor: pointer;
  }
  label.is-selected {
    border-color: color-mix(in srgb, var(--nh-accent) 55%, transparent);
    background: var(--nh-accent-soft);
  }
  input {
    width: 1.1rem;
    height: 1.1rem;
    accent-color: var(--nh-accent);
    flex: none;
  }
  .nh-essential-tx {
    display: flex;
    flex-direction: column;
    min-width: 0;
  }
  .nh-essential-tx b {
    font-size: 0.85rem;
  }
  .nh-essential-tx small {
    color: var(--nh-text-muted);
    font-size: 0.72rem;
  }
</style>
