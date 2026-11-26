<script setup lang="ts">
  /*
   * Night essentials picker. Options come from the panel projection so a choice always maps
   * to a control the panel can render; the limit is enforced here and in the domain helper.
   */
  import { ref, watch } from 'vue';
  import { PANEL_ESSENTIAL_LIMIT, togglePanelEssential, type PanelEssential } from '@/domain/panel-essentials';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiDialog from '@/components/ui/UiDialog.vue';

  const props = defineProps<{
    open: boolean;
    options: readonly (PanelEssential & { room: string })[];
    selected: readonly PanelEssential[];
  }>();
  const emit = defineEmits<{ (e: 'close'): void; (e: 'save', value: PanelEssential[]): void }>();

  const draft = ref<PanelEssential[]>([]);
  watch(
    () => props.open,
    (open) => {
      if (open) draft.value = [...props.selected];
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
    emit('close');
  }
</script>

<template>
  <UiDialog :open="open" title="Night essentials" wide @close="$emit('close')">
    <p class="nh-muted">
      Choose up to {{ PANEL_ESSENTIAL_LIMIT }} controls for the simplified night panel.
      {{ draft.length }} of {{ PANEL_ESSENTIAL_LIMIT }} selected.
    </p>
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
