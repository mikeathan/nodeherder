<script setup lang="ts">
  /* Page title block with optional back link, subtitle and actions. One <h1> per screen. */
  import type { RouteLocationRaw } from 'vue-router';
  import UiIcon from './UiIcon.vue';

  defineProps<{ title: string; subtitle?: string; back?: RouteLocationRaw; backLabel?: string }>();
</script>

<template>
  <header class="nh-page-head">
    <div class="nh-page-titles">
      <RouterLink v-if="back" :to="back" class="nh-back"><UiIcon name="back" />{{ backLabel ?? 'Back' }}</RouterLink>
      <div class="nh-page-title-row">
        <slot name="before" />
        <div class="nh-page-title-text">
          <h1 class="nh-h1">{{ title }}</h1>
          <p v-if="subtitle || $slots.subtitle" class="nh-sub"><slot name="subtitle">{{ subtitle }}</slot></p>
        </div>
      </div>
    </div>
    <div v-if="$slots.actions" class="nh-page-actions"><slot name="actions" /></div>
  </header>
</template>

<style scoped>
  .nh-page-head {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    justify-content: space-between;
    gap: 0.75rem;
    margin-bottom: calc(var(--nh-space) * 1.1);
  }
  .nh-page-titles {
    min-width: 0;
    flex: 1 1 16rem;
  }
  .nh-page-title-row {
    display: flex;
    align-items: center;
    gap: 0.85rem;
    min-width: 0;
  }
  .nh-page-title-text {
    min-width: 0;
  }
  .nh-h1 {
    font-size: 1.6rem;
    font-weight: 750;
    letter-spacing: -0.015em;
    line-height: 1.2;
    overflow-wrap: anywhere;
  }
  .nh-sub {
    margin: 0.25rem 0 0;
    color: var(--nh-text-muted);
    font-size: 0.9rem;
  }
  .nh-page-actions {
    display: flex;
    flex-wrap: wrap;
    gap: 0.45rem;
  }
  .nh-back {
    display: inline-flex;
    align-items: center;
    gap: 0.25rem;
    color: var(--nh-text-muted);
    text-decoration: none;
    font-size: 0.85rem;
    margin-bottom: 0.35rem;
    min-height: 1.75rem;
  }
  .nh-back:hover {
    color: var(--nh-accent);
  }
  @media (max-width: 640px) {
    .nh-h1 {
      font-size: 1.35rem;
    }
  }
</style>
