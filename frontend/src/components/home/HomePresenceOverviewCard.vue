<script setup lang="ts">
  import UiButton from '@/components/ui/UiButton.vue';
  import HomePresenceOverview from './HomePresenceOverview.vue';

  defineProps<{ editing?: boolean; busy?: boolean }>();
  defineEmits<{ remove: [] }>();
</script>

<template>
  <section class="nh-overview-card" aria-label="Presence overview">
    <header>
      <strong>Presence overview</strong>
      <UiButton v-if="editing" size="sm" variant="ghost" icon="delete" label="Remove presence overview"
        :disabled="busy" @click="$emit('remove')" />
    </header>
    <HomePresenceOverview />
  </section>
</template>

<style scoped>
  .nh-overview-card { display: grid; gap: 0.4rem; min-width: 0; }
  header { display: flex; justify-content: space-between; align-items: center; gap: 0.5rem; }
</style>
