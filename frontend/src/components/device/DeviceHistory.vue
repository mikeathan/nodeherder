<script setup lang="ts">
  import { computed, ref } from 'vue';
  import { findDevice } from '@/composables/useHub';
  import { supportsSpatialMapping } from '@/domain/spatial';
  import { copyWorkspace } from '@/domain/spatial/workspace';
  import { useDeviceSettings } from '@/composables/useDeviceSettings';
  import DeviceMetrics from './DeviceMetrics.vue';
  import SpatialActivityHistory from './spatial/SpatialActivityHistory.vue';
  import SpatialPositions from './spatial/SpatialPositions.vue';
  import UiButton from '@/components/ui/UiButton.vue';

  const props = defineProps<{ id: string }>();
  const device = computed(() => findDevice(props.id));
  const spatial = computed(() => supportsSpatialMapping(device.value));
  const { settings } = useDeviceSettings(() => props.id);
  const workspace = computed(() => copyWorkspace(settings.value.spatial?.workspace));
  const mode = ref<'charts' | 'activity' | 'positions'>('charts');
</script>

<template>
  <section class="device-history-modes" aria-label="History views">
    <div v-if="spatial" class="nh-history-modes" role="group" aria-label="History view">
      <UiButton size="sm" :variant="mode === 'charts' ? 'primary' : 'ghost'" :pressed="mode === 'charts'" @click="mode = 'charts'">Charts</UiButton>
      <UiButton size="sm" :variant="mode === 'activity' ? 'primary' : 'ghost'" :pressed="mode === 'activity'" @click="mode = 'activity'">Zone activity</UiButton>
      <UiButton size="sm" :variant="mode === 'positions' ? 'primary' : 'ghost'" :pressed="mode === 'positions'" @click="mode = 'positions'">Positions</UiButton>
    </div>
    <DeviceMetrics v-if="mode === 'charts' || !spatial" :id="id" />
    <SpatialActivityHistory v-else-if="mode === 'activity' && device" :device="device" :workspace="workspace" />
    <SpatialPositions v-else-if="mode === 'positions'" :device-id="id" />
  </section>
</template>

<style scoped>
  .nh-history-modes { display: flex; flex-wrap: wrap; gap: 0.4rem; margin-block-end: 0.75rem; }
</style>
