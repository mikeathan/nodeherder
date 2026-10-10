<script setup lang="ts">
  /* Empty, loading-failed and "nothing here yet" states with a next step. */
  import UiIcon from './UiIcon.vue';
  import { IconName } from './icons';

  defineProps<{ icon?: IconName; title: string; text?: string; tone?: 'default' | 'error' }>();
</script>

<template>
  <div class="nh-empty" :class="{ 'is-error': tone === 'error' }" :role="tone === 'error' ? 'alert' : undefined">
    <UiIcon :name="icon ?? 'info'" />
    <h2>{{ title }}</h2>
    <p v-if="text">{{ text }}</p>
    <div v-if="$slots.default" class="nh-inline"><slot /></div>
  </div>
</template>

<style scoped>
  .nh-empty {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 0.5rem;
    padding: 2.5rem 1rem;
    color: var(--nh-text-muted);
  }
  .nh-empty > .nh-ic {
    width: 2.5rem;
    height: 2.5rem;
    color: var(--nh-accent);
  }
  .nh-empty h2 {
    color: var(--nh-text);
    font-size: 1.05rem;
  }
  .nh-empty p {
    margin: 0 0 0.5rem;
    max-width: 28rem;
  }
  .nh-empty.is-error > .nh-ic {
    color: var(--nh-danger);
  }
</style>
