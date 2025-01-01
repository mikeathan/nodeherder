<script setup>
  import LinkQuality from '../device/LinkQuality.vue';
  import PowerSource from '../device/PowerSource.vue';
  import LastSeen from '../device/LastSeen.vue';
  import { getOfflineIcon } from '@/modules/formatters/device.formatter';
  import Icon from '../controls/Icon.vue';

  const props = defineProps({
    device: Object,
  });
</script>
<template>
  <div
    class="grid justify-content-between align-items-center">
    <LastSeen
      :timestamp="device.properties.last_seen"></LastSeen>
    <div
      class="text-truncate"
      v-if="device.properties.availability === 'online'">
      <LinkQuality :value="device.properties.linkquality" />
      <PowerSource
        :power_source="device.power_source"
        :value="device.properties.battery" />
    </div>
    <div className="col-auto text-truncate" v-else>
      <Icon :icon="getOfflineIcon()" />
    </div>
  </div>
</template>
