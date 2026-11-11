<script setup lang="ts">
  import { computed } from 'vue';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import { useHub } from '@/composables/useHub';
  import { PRESENCE_COPY, presenceOverviewModel } from '@/domain/spatial/presence';
  import HomePresenceOverview from './HomePresenceOverview.vue';

  const props = defineProps<{ editing?: boolean; busy?: boolean }>();
  const { devices, statusOf } = useHub();
  const model = computed(() => presenceOverviewModel(devices.value, !!props.editing, statusOf));
  defineEmits<{ remove: [] }>();
</script>

<template>
  <section class="nh-overview-card" :aria-label="PRESENCE_COPY.overview.ariaLabel">
    <header v-if="model.showHeader" class="nh-overview-head">
      <span class="nh-overview-icon"><UiIcon name="presence" /></span>
      <span class="nh-overview-title" :class="{ 'is-single': model.isSingle }">
        <strong>{{ model.title }}</strong>
        <span class="nh-muted">{{ model.subtitle }}</span>
      </span>
      <UiButton
        v-if="editing"
        size="sm"
        variant="ghost"
        icon="delete"
        :label="PRESENCE_COPY.overview.remove"
        :disabled="busy"
        @click="$emit('remove')" />
    </header>
    <HomePresenceOverview :sensors="model.sensors" :show-names="model.showNames" />
  </section>
</template>

<style scoped>
  .nh-overview-card {
    container-type: inline-size;
    display: grid;
    gap: 0.35rem;
    min-width: 0;
    padding: 0.6rem 0.75rem;
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
  .nh-overview-title strong {
    font-size: 0.72rem;
    font-weight: 650;
    letter-spacing: 0.06em;
    text-transform: uppercase;
    color: var(--nh-text-muted);
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
  .nh-overview-title .nh-muted { font-size: 0.68rem; }
  .nh-overview-title.is-single strong {
    font-size: 0.875rem;
    font-weight: 600;
    letter-spacing: 0;
    text-transform: none;
    color: var(--nh-text);
  }
  .nh-overview-title.is-single .nh-muted { font-size: 0.7rem; }
</style>
