<script setup lang="ts">
  /*
   * Circular power control for an entity view. It uses the same draft/confirmation
   * path as every other control and says On/Off/Unknown instead of presenting a
   * missing state as "off" (NH-02, FE-03).
   */
  import { computed } from 'vue';
  import type { Device, Expose } from '@/types/device';
  import { binaryOffValue, binaryOnValue, binaryStateOf } from '@/domain/exposes';
  import { useExposeEdit } from '@/composables/useExposeEdit';
  import UiIcon from '@/components/ui/UiIcon.vue';
  const props = defineProps<{ device: Device; expose: Expose }>();
  const edit = useExposeEdit(
    () => props.device,
    () => props.expose
  );
  const { pending, disabled } = edit;
  const on = computed(() => binaryStateOf(props.expose));
  const text = computed(() => (on.value === null ? 'Unknown' : on.value ? 'On' : 'Off'));
  function toggle() {
    edit.commit(on.value === true ? binaryOffValue(props.expose) : binaryOnValue(props.expose));
  }
</script>
<template>
  <div class="entity-power-control">
    <button
      type="button"
      role="switch"
      class="entity-power"
      :class="{ 'is-on': on === true, 'is-unknown': on === null }"
      :aria-checked="on === true"
      :aria-label="`Power, ${text}`"
      :disabled="disabled"
      @click="toggle">
      <UiIcon name="power" />
    </button>
    <span class="entity-power-label">Power</span>
    <span class="entity-power-state" role="status">{{ pending ? 'Sending…' : text }}</span>
  </div>
</template>
<style scoped>
  .entity-power-control {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.4rem;
    min-width: 0;
    font-size: 0.75rem;
  }
  .entity-power {
    display: grid;
    place-items: center;
    width: 3rem;
    height: 3rem;
    min-width: 44px;
    min-height: 44px;
    padding: 0;
    border-radius: 50%;
    border: 1px solid var(--nh-border-strong);
    background: var(--nh-surface-2);
    color: var(--nh-text-muted);
    cursor: pointer;
  }
  .entity-power.is-on {
    background: var(--nh-accent);
    border-color: var(--nh-accent);
    color: var(--nh-accent-contrast);
  }
  .entity-power.is-unknown {
    border-style: dashed;
  }
  .entity-power:focus-visible {
    outline: 3px solid var(--nh-focus);
    outline-offset: 3px;
  }
  .entity-power:disabled {
    opacity: 0.45;
    cursor: not-allowed;
  }
  .entity-power-label {
    color: var(--nh-text);
  }
  .entity-power-state {
    min-height: 1rem;
    color: var(--nh-text-muted);
    font-size: 0.7rem;
  }
</style>
