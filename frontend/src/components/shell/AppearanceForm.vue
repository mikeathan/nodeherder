<script setup lang="ts">
  /*
   * Appearance customiser (spec 007 FR-02, AC-02). Shared by the header panel and
   * Settings → Appearance; both edit the same useThemeSettings state, applied instantly.
   */
  import { computed } from 'vue';
  import { useThemeSettings } from '@/composables/useThemeSettings';
  import { ACCENTS, Density, ThemeMode, THEMES } from '@/theme/theme-settings';
  import UiSegmented from '@/components/ui/UiSegmented.vue';
  import UiToggle from '@/components/ui/UiToggle.vue';
  import UiButton from '@/components/ui/UiButton.vue';

  const { settings, update, reset } = useThemeSettings();

  const modes: { value: ThemeMode; label: string }[] = [
    { value: 'system', label: 'System' },
    { value: 'light', label: 'Light' },
    { value: 'dark', label: 'Dark' },
  ];
  const densities: { value: Density; label: string }[] = [
    { value: 'compact', label: 'Compact' },
    { value: 'comfortable', label: 'Comfortable' },
    { value: 'spacious', label: 'Spacious' },
  ];
  const customAccent = computed(() => {
    const accent = settings.value.accent;
    return accent !== 'theme' && !ACCENTS.includes(accent) ? accent : null;
  });
</script>

<template>
  <div class="nh-appearance">
    <div class="nh-field">
      <span class="nh-label">Mode</span>
      <UiSegmented :model-value="settings.mode" :options="modes" label="Mode" @update:model-value="update({ mode: $event })" />
    </div>

    <fieldset class="nh-fieldset">
      <legend class="nh-label">Theme</legend>
      <div class="nh-swatches">
        <button
          v-for="t in THEMES"
          :key="t.id"
          type="button"
          class="nh-swatch"
          :aria-pressed="settings.theme === t.id"
          @click="update({ theme: t.id })">
          <span :style="{ background: t.swatch[0] }" /><span :style="{ background: t.swatch[1] }" />
          <em>{{ t.name }}</em>
        </button>
      </div>
    </fieldset>

    <fieldset class="nh-fieldset">
      <legend class="nh-label">Accent colour</legend>
      <div class="nh-swatches is-accent">
        <button type="button" class="nh-swatch is-theme" :aria-pressed="settings.accent === 'theme'" @click="update({ accent: 'theme' })">
          <em>Theme</em>
        </button>
        <button
          v-for="hex in ACCENTS"
          :key="hex"
          type="button"
          class="nh-swatch"
          :aria-pressed="settings.accent === hex"
          :aria-label="`Accent ${hex}`"
          @click="update({ accent: hex })">
          <span :style="{ background: hex }" />
        </button>
        <label class="nh-swatch is-custom" :aria-pressed="customAccent !== null" title="Custom colour">
          <span class="sr-only">Custom accent colour</span>
          <input type="color" :value="customAccent ?? '#1b7fb0'" @input="update({ accent: ($event.target as HTMLInputElement).value })" />
        </label>
      </div>
    </fieldset>

    <div class="nh-field">
      <span class="nh-label">Density</span>
      <UiSegmented :model-value="settings.density" :options="densities" label="Density" @update:model-value="update({ density: $event })" />
    </div>

    <div class="nh-field">
      <label class="nh-label" for="ap-radius">Corner roundness · {{ settings.radius }}</label>
      <input id="ap-radius" class="nh-range" type="range" min="0" max="20" :value="settings.radius" @input="update({ radius: Number(($event.target as HTMLInputElement).value) })" />
    </div>

    <div class="nh-field">
      <label class="nh-label" for="ap-scale">Text size · {{ settings.fontScale }} %</label>
      <input id="ap-scale" class="nh-range" type="range" min="90" max="125" step="5" :value="settings.fontScale" @input="update({ fontScale: Number(($event.target as HTMLInputElement).value) })" />
    </div>

    <div class="nh-row-field">
      <span>Background texture</span>
      <UiToggle :model-value="settings.effects" label="Background texture" @update:model-value="update({ effects: $event })" />
    </div>

    <div>
      <UiButton size="sm" icon="refresh" @click="reset">Reset to defaults</UiButton>
    </div>
  </div>
</template>

<style scoped>
  .nh-appearance {
    display: flex;
    flex-direction: column;
    gap: 1.1rem;
  }
  .nh-fieldset {
    border: 0;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.45rem;
    min-width: 0;
  }
  .nh-fieldset legend {
    padding: 0;
    margin-bottom: 0.45rem;
  }
  .nh-swatches {
    display: flex;
    flex-wrap: wrap;
    gap: 0.45rem;
  }
  .nh-swatch {
    display: inline-flex;
    align-items: center;
    padding: 0.25rem;
    border-radius: var(--nh-radius-m);
    border: 2px solid var(--nh-border);
    background: var(--nh-surface);
    color: var(--nh-text);
    cursor: pointer;
    min-height: 2.3rem;
  }
  .nh-swatch span {
    width: 1.25rem;
    height: 1.6rem;
    display: block;
  }
  .nh-swatch span:first-child {
    border-radius: 4px 0 0 4px;
  }
  .nh-swatch span:nth-child(2) {
    border-radius: 0 4px 4px 0;
  }
  .nh-swatch em {
    font-style: normal;
    font-size: 0.78rem;
    padding: 0 0.4rem;
  }
  .nh-swatches.is-accent .nh-swatch span {
    width: 1.6rem;
    border-radius: 50%;
  }
  .nh-swatch[aria-pressed='true'] {
    border-color: var(--nh-accent);
    box-shadow: 0 0 0 2px var(--nh-accent-soft);
  }
  .nh-swatch.is-custom {
    position: relative;
    width: 2.3rem;
    background: conic-gradient(red, yellow, lime, cyan, blue, magenta, red);
  }
  .nh-swatch.is-custom input {
    position: absolute;
    inset: 0;
    opacity: 0;
    cursor: pointer;
    width: 100%;
    height: 100%;
  }
  .nh-range {
    width: 100%;
    accent-color: var(--nh-accent);
  }
  .nh-row-field {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
  }
</style>
