<script setup lang="ts">
  /* One expose as a labelled row: kind icon, name, optional help, and its value or control. */
  import { computed, ref, useId } from 'vue';
  import { Device, Expose } from '@/types/device';
  import { controlOf, exposeControlUnavailable, exposeKind, exposeSemanticName, exposeDisplayLabel, isBinaryOn } from '@/domain/exposes';
  import { EntityHelpKinds, resolveEntityHelp } from '@/domain/entityHelp';
  import { exposeIcon } from '@/components/ui/icons';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import ExposeControl from './ExposeControl.vue';

  const props = defineProps<{ device: Device; expose: Expose; showDescription?: boolean; showHelp?: boolean; card?: boolean }>();
  const label = computed(() => exposeDisplayLabel(props.expose));
  const icon = computed(() => exposeIcon(exposeSemanticName(props.expose), isBinaryOn(props.expose)));
  const cardSlider = computed(() => props.card && controlOf(props.expose) === 'slider');

  const help = computed(() => resolveEntityHelp(props.expose));
  const showInlineHelp = computed(() => props.showDescription && help.value?.kind === EntityHelpKinds.Provider);
  const inlineHelp = computed(() => (showInlineHelp.value ? help.value : null));
  const helpToggle = computed(() => {
    if (!props.showHelp || !help.value || showInlineHelp.value) return null;
    return help.value;
  });
  const helpOpen = ref(false);
  const helpPanelId = `entity-help-${useId().replace(/[^a-zA-Z0-9_-]/g, '')}`;
</script>

<template>
  <div class="nh-row" :class="[`kind-${exposeKind(exposeSemanticName(expose))}`, { 'is-card': card, 'has-slider': cardSlider }]">
    <UiIcon :path="icon" class="nh-row-ic" />
    <div v-if="!cardSlider" class="nh-row-label">
      <div class="nh-row-title">
        <span class="nh-row-name">{{ label }}</span>
        <button
          v-if="helpToggle"
          type="button"
          class="nh-row-help-toggle"
          :aria-label="`About ${label}`"
          :aria-expanded="helpOpen"
          :aria-controls="helpPanelId"
          @click="helpOpen = !helpOpen">
          <UiIcon name="info" />
        </button>
      </div>
      <small v-if="inlineHelp">{{ inlineHelp.text }}</small>
      <p v-if="helpToggle" v-show="helpOpen" :id="helpPanelId" class="nh-row-help">
        {{ helpToggle.text }}
        <a
          v-if="helpToggle.source"
          :href="helpToggle.source.url"
          target="_blank"
          rel="noopener noreferrer"
          >{{ helpToggle.source.label }}</a
        >
      </p>
      <small v-if="exposeControlUnavailable(expose)">{{ exposeControlUnavailable(expose) }}</small>
    </div>
    <ExposeControl :device="device" :expose="expose" :label="label" :card="card" />
  </div>
</template>

<style scoped>
  .nh-row {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.4rem 0.65rem;
    padding: 0.45rem var(--nh-space);
    min-height: calc(var(--nh-ctl-h) + 0.25rem);
  }
  .nh-row + .nh-row {
    border-top: var(--nh-border-w) solid var(--nh-border);
  }
  .nh-row-ic {
    color: var(--nh-kc, var(--nh-text-muted));
  }
  .nh-row-label {
    flex: 1 1 6rem;
    display: flex;
    flex-direction: column;
    min-width: 0;
    font-size: 0.9rem;
  }
  .nh-row-title {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    min-width: 0;
  }
  .nh-row-name {
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .nh-row-help-toggle {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex: none;
    padding: 0.1rem;
    border: 0;
    border-radius: var(--nh-radius-s);
    background: none;
    color: var(--nh-text-muted);
    cursor: pointer;
  }
  .nh-row-help-toggle:hover {
    color: var(--nh-text);
  }
  .nh-row-help-toggle:focus-visible {
    outline: 2px solid var(--nh-accent);
    outline-offset: 1px;
  }
  .nh-row-help-toggle .nh-ic {
    width: 1em;
    height: 1em;
  }
  .nh-row-help {
    margin: 0.25rem 0 0;
    max-width: 44ch;
    color: var(--nh-text-muted);
    line-height: 1.5;
    white-space: normal;
  }
  .nh-row-help a {
    display: inline-block;
    margin-top: 0.15rem;
  }
  .nh-row-label small {
    color: var(--nh-text-muted);
  }
  .nh-row.is-card {
    display: grid;
    grid-template-columns: 1.1rem minmax(0, 1fr) auto;
    column-gap: 0.65rem;
    min-height: 3rem;
    padding-block: 0.6rem;
  }
  .is-card .nh-row-ic {
    width: 1.1rem;
    height: 1.1rem;
  }
  .is-card.has-slider {
    align-items: start;
    padding-bottom: 0.25rem;
  }
  .is-card.has-slider .nh-row-ic {
    margin-top: 0.1rem;
  }
  .is-card.has-slider :deep(.nh-xc) {
    grid-column: 2 / -1;
  }
</style>
