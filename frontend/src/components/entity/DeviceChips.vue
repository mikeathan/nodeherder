<script setup lang="ts">
  /*
   * Protocol, signal and power for a device (spec 007 FR-07). Signal uses Zigbee link
   * quality or Wi-Fi RSSI when reported; nothing is shown for values a device does not report.
   */
  import { computed } from 'vue';
  import { Device } from '@/types/device';
  import { batteryLevel, hasLowBattery, hasWeakLink, isBatteryPowered, linkQuality, resolveProtocol, wifiSignal } from '@/domain/devices';
  import { protocolIcon } from '@/components/ui/icons';
  import UiChip from '@/components/ui/UiChip.vue';

  const props = defineProps<{ device: Device }>();
  const protocol = computed(() => resolveProtocol(props.device.connection_type));
  const lqi = computed(() => linkQuality(props.device));
  const rssi = computed(() => wifiSignal(props.device));
  const battery = computed(() => batteryLevel(props.device));
</script>

<template>
  <UiChip :icon-path="protocolIcon(device.connection_type)" :class="`is-proto-${protocol.id}`">{{ protocol.label }}</UiChip>
  <UiChip v-if="lqi !== null" icon="signal" :tone="hasWeakLink(device) ? 'warn' : 'default'" :title="'Link quality (0–255)'">{{ lqi }} LQI</UiChip>
  <UiChip v-else-if="rssi !== null" icon="wifi" title="Wi-Fi signal">{{ rssi }} dBm</UiChip>
  <UiChip v-if="isBatteryPowered(device)" :icon="hasLowBattery(device) ? 'batteryLow' : 'battery'" :tone="hasLowBattery(device) ? 'warn' : 'default'">
    {{ battery === null ? 'Battery' : `${Math.round(battery)} %` }}
  </UiChip>
  <UiChip v-else-if="device.power_source" tone="muted">Mains</UiChip>
</template>

<style scoped>
  .is-proto-mqtt :deep(.nh-ic) {
    color: var(--nh-kind-zigbee);
  }
  .is-proto-http :deep(.nh-ic) {
    color: var(--nh-kind-wifi);
  }
</style>
