<script setup lang="ts">
  /*
   * Device-page entry point for the presence workspace. It shows one coarse, honest status
   * (offline, not set up, clear, or how many slots report a position) and opens the Presence
   * tab. It deliberately does not run the tracking pipeline: filtering, association and hold
   * state belong to the workspace, so the chip cannot disagree with it about tracks.
   */
  import { computed } from 'vue';
  import type { Device } from '@/types/device';
  import type { SpatialFrame } from '@/types/spatial';
  import { store } from '@/store';
  import { hasCompleteSpatialMapping, mergeSpatialMappings } from '@/domain/spatial';
  import { spatialHeadline } from '@/domain/spatial/summary';
  import { useDeviceSettings } from '@/composables/useDeviceSettings';
  import { useHub } from '@/composables/useHub';
  import UiIcon from '@/components/ui/UiIcon.vue';

  const props = defineProps<{ device: Device }>();
  const emit = defineEmits<{ open: [] }>();

  const { settings } = useDeviceSettings(() => props.device.id);
  const { isConnected, statusOf } = useHub();
  const configured = computed(() =>
    hasCompleteSpatialMapping(mergeSpatialMappings(settings.value?.spatial?.mappings))
  );
  const online = computed(() => isConnected.value && statusOf(props.device) === 'online');
  const activeSlots = computed(() => {
    const frame = store.getters['hub/spatialFrame'](props.device.id) as SpatialFrame | undefined;
    return frame ? frame.targets.filter((target) => target.active && target.x && target.y).length : 0;
  });
  const headline = computed(() =>
    spatialHeadline({ online: online.value, configured: configured.value, activeSlots: activeSlots.value })
  );
</script>

<template>
  <button
    type="button"
    class="nh-chip nh-presence-chip"
    :class="`is-${headline.tone}`"
    :aria-label="`Presence: ${headline.label}. Open the presence view`"
    @click="emit('open')">
    <UiIcon name="signal" />
    <span>Presence · {{ headline.label }}</span>
  </button>
</template>

<style scoped>
  .nh-presence-chip {
    border: 0;
    cursor: pointer;
    font: inherit;
    font-size: 0.75rem;
    font-weight: 500;
  }
  .nh-presence-chip:hover {
    filter: brightness(1.08);
  }
  .nh-presence-chip:focus-visible {
    outline: 2px solid var(--nh-accent);
    outline-offset: 2px;
  }
</style>
