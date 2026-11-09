<script setup lang="ts">
  /*
   * One labelled setting: name and explanation on the leading side, the control on the
   * trailing side. Stacks on narrow screens. The control keeps its own accessible name;
   * `forId` ties the visible label to a native input when one is rendered.
   */
  defineProps<{ label: string; help?: string; forId?: string }>();
</script>

<template>
  <div class="nh-setting-row">
    <div class="nh-setting-copy">
      <label v-if="forId" :for="forId" class="nh-setting-label">{{ label }}</label>
      <span v-else class="nh-setting-label">{{ label }}</span>
      <p v-if="help" class="nh-help">{{ help }}</p>
      <slot name="help" />
    </div>
    <div class="nh-setting-control">
      <slot />
      <slot name="error" />
    </div>
  </div>
</template>

<style scoped>
  .nh-setting-row {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
    padding: 0.85rem 0;
  }
  .nh-setting-row + .nh-setting-row {
    border-top: var(--nh-border-w) solid var(--nh-border);
  }
  .nh-setting-copy {
    min-width: 0;
    flex: 1 1 16rem;
  }
  .nh-setting-label {
    display: block;
    font-size: 0.9rem;
    font-weight: 600;
    color: var(--nh-text);
  }
  .nh-help {
    margin: 0.3rem 0 0;
    max-width: 44ch;
  }
  .nh-setting-control {
    display: flex;
    align-items: center;
    justify-content: flex-end;
    flex-wrap: wrap;
    gap: 0.5rem;
    min-height: var(--nh-ctl-h);
    min-width: 0;
  }
  @media (max-width: 640px) {
    .nh-setting-row {
      flex-direction: column;
      gap: 0.6rem;
    }
    /* In a column container flex-basis is a height, so the row basis must be released. */
    .nh-setting-copy {
      flex: none;
    }
    .nh-setting-control {
      width: 100%;
      justify-content: flex-start;
    }
  }
</style>
