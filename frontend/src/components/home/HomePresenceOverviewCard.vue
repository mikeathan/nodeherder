<script setup lang="ts">
  import { computed } from 'vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import { useHub } from '@/composables/useHub';
  import { supportsSpatialMapping } from '@/domain/spatial';
  import HomePresenceOverview from './HomePresenceOverview.vue';

  defineProps<{ editing?: boolean; busy?: boolean }>();
  const { devices } = useHub();
  const count = computed(() => devices.value.filter((device) => supportsSpatialMapping(device)).length);
  defineEmits<{ remove: [] }>();
</script>

<template>
  <section class="nh-overview-card" aria-label="Presence overview">
    <header class="nh-overview-head">
      <span class="nh-overview-icon"><UiIcon name="presence" /></span>
      <span class="nh-overview-title">
        <strong>Presence overview</strong>
        <span class="nh-muted">{{ count }} sensor{{ count === 1 ? '' : 's' }}</span>
      </span>
      <UiButton v-if="editing" size="sm" variant="ghost" icon="delete" label="Remove presence overview"
        :disabled="busy" @click="$emit('remove')" />
    </header>
    <HomePresenceOverview />
  </section>
</template>

<style scoped>
  .nh-overview-card {
    container-type: inline-size;
    display: grid;
    gap: 0.55rem;
    min-width: 0;
    padding: 0.65rem 0.75rem;
    border-radius: var(--nh-radius-l);
    background: var(--nh-surface);
    border: var(--nh-border-w) solid var(--nh-border);
    box-shadow: var(--nh-shadow-1);
  }
  .nh-overview-head { display: flex; align-items: center; gap: 0.65rem; min-width: 0; }
  .nh-overview-icon {
    display: grid;
    place-items: center;
    width: 2.25rem;
    height: 2.25rem;
    flex: none;
    border-radius: 50%;
    background: var(--nh-surface-2);
    color: var(--nh-text-muted);
  }
  .nh-overview-title { display: grid; min-width: 0; }
  .nh-overview-title strong { overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
  .nh-overview-title .nh-muted { font-size: 0.8rem; }
</style>
