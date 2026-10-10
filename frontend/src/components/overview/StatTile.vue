<script setup lang="ts">
  import type { RouteLocationRaw } from 'vue-router';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import { IconName } from '@/components/ui/icons';

  defineProps<{ to: RouteLocationRaw; icon: IconName; value: string; label: string; hint?: string; tone?: 'ok' | 'warn' | 'danger' }>();
</script>

<template>
  <RouterLink :to="to" class="nh-stat" :class="tone && `is-${tone}`">
    <span class="nh-stat-num">{{ value }}</span>
    <span class="nh-stat-ic"><UiIcon :name="icon" /></span>
    <span class="nh-stat-label">{{ label }}</span>
    <span v-if="hint" class="nh-stat-hint">{{ hint }}</span>
  </RouterLink>
</template>

<style scoped>
  .nh-stat {
    display: grid;
    grid-template-columns: 1fr auto;
    grid-template-areas: 'num ic' 'label label' 'hint hint';
    padding: 0.85rem 1rem;
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface);
    border: var(--nh-border-w) solid transparent;
    text-decoration: none;
    box-shadow: var(--nh-shadow-1);
    min-width: 0;
  }
  [data-mode='dark'] .nh-stat {
    border-color: var(--nh-border);
  }
  .nh-stat:hover {
    border-color: var(--nh-accent);
  }
  .nh-stat-num {
    grid-area: num;
    font-size: 1.9rem;
    font-weight: 800;
    line-height: 1.1;
    font-variant-numeric: tabular-nums;
  }
  .nh-stat-ic {
    grid-area: ic;
    color: var(--nh-accent);
  }
  .nh-stat-ic .nh-ic {
    width: 1.6rem;
    height: 1.6rem;
  }
  .nh-stat-label {
    grid-area: label;
    font-size: 0.82rem;
    color: var(--nh-text-muted);
    font-weight: 500;
  }
  .nh-stat-hint {
    grid-area: hint;
    font-size: 0.72rem;
    color: var(--nh-text-muted);
  }
  .nh-stat.is-warn .nh-stat-num,
  .nh-stat.is-warn .nh-stat-ic {
    color: var(--nh-warn);
  }
  .nh-stat.is-danger .nh-stat-num,
  .nh-stat.is-danger .nh-stat-ic {
    color: var(--nh-danger);
  }
  .nh-stat.is-ok .nh-stat-ic {
    color: var(--nh-ok);
  }
</style>
