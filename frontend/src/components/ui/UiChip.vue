<script setup lang="ts">
  /* Small status label (.nh-chip in base.css). */
  import UiIcon from './UiIcon.vue';
  import { IconName } from './icons';

  defineProps<{ tone?: 'default' | 'ok' | 'warn' | 'danger' | 'info' | 'accent' | 'muted'; icon?: IconName; iconPath?: string }>();
</script>

<template>
  <span class="nh-chip" :class="tone && tone !== 'default' ? `is-${tone}` : undefined">
    <UiIcon v-if="icon || iconPath" :name="icon" :path="iconPath" />
    <slot />
  </span>
</template>
