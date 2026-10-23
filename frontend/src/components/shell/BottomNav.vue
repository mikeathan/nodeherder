<script setup lang="ts">
  import { computed } from 'vue';
  import { useRoute } from 'vue-router';
  import { PRIMARY_NAV_ITEMS, isNavActive } from './navigation';
  import UiIcon from '@/components/ui/UiIcon.vue';
  defineProps<{ open: boolean }>();
  defineEmits<{ more: [] }>();
  const route = useRoute();
  const moreCurrent = computed(() => !PRIMARY_NAV_ITEMS.some((item) => isNavActive(item, route.name)));
</script>
<template>
  <nav class="nh-bottom-nav" aria-label="Primary">
    <RouterLink v-for="item in PRIMARY_NAV_ITEMS" :key="item.route" :to="{ name: item.route }"
      :aria-current="isNavActive(item, route.name) ? 'page' : undefined">
      <UiIcon :name="item.icon" /><span>{{ item.label }}</span>
    </RouterLink>
    <button type="button" :aria-current="moreCurrent ? 'page' : undefined" :aria-expanded="open" aria-controls="nh-nav" @click="$emit('more')">
      <UiIcon name="menu" /><span>More</span>
    </button>
  </nav>
</template>
<style scoped>
  .nh-bottom-nav { display: none; }
  @media (max-width: 719px) {
    .nh-bottom-nav { position: fixed; bottom: 0; left: 0; right: 0; z-index: 50; display: grid; grid-template-columns: repeat(5, minmax(0, 1fr));
      padding-bottom: var(--nh-safe-bottom); background: var(--nh-surface); border-top: var(--nh-border-w) solid var(--nh-border-strong); }
    a, button { color: var(--nh-text-muted); background: none; border: 0; text-decoration: none; font: inherit; font-size: 0.65rem;
      display: flex; flex-direction: column; gap: 0.2rem; align-items: center; justify-content: center; min-height: var(--nh-bottom-nav-h); min-width: 44px; cursor: pointer; }
    [aria-current='page'] { color: var(--nh-accent); background: var(--nh-accent-soft); }
    a:focus-visible, button:focus-visible { outline: 3px solid var(--nh-focus); outline-offset: -3px; }
  }
</style>
