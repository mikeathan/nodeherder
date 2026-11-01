<script setup lang="ts">
  /*
   * Protocol, signal and power for a device (spec 007 FR-07). Presentation values
   * and thresholds live in domain/device-chips.ts; this component only renders.
   */
  import { computed } from 'vue';
  import { Device } from '@/types/device';
  import { resolveProtocol } from '@/domain/devices';
  import { linkChip, powerChip } from '@/domain/devices/chips';
  import { protocolIcon } from '@/components/ui/icons';
  import UiChip from '@/components/ui/UiChip.vue';

  const props = defineProps<{ device: Device }>();
  const protocol = computed(() => resolveProtocol(props.device.connection_type));
  const link = computed(() => linkChip(props.device));
  const power = computed(() => powerChip(props.device));
</script>

<template>
  <UiChip :icon-path="protocolIcon(device.connection_type)" :class="`is-proto-${protocol.id}`">{{ protocol.label }}</UiChip>
  <UiChip v-if="link" :icon="link.icon" :tone="link.tone" :title="link.title">{{ link.label }}</UiChip>
  <UiChip v-if="power" :icon="power.icon" :tone="power.tone">{{ power.label }}</UiChip>
</template>

<style scoped>
  .is-proto-mqtt :deep(.nh-ic) {
    color: var(--nh-kind-zigbee);
  }
  .is-proto-http :deep(.nh-ic),
  .is-proto-esphome :deep(.nh-ic) {
    color: var(--nh-kind-wifi);
  }
</style>
