<script setup lang="ts">
  import LinkQuality from '../../device/LinkQuality.vue';
  import PowerSource from '../../device/PowerSource.vue';
  import LastSeen from '../../device/LastSeen.vue';
  import { PropType } from 'vue';
  import { getOfflineIcon } from '@/modules/formatters/device.formatter';
  import Icon from '../../controls/Icon.vue';
  import { Device } from '@/types/device';
  import { computed } from 'vue';
import { getPowerSourceValue } from '@/contracts/device';

  const props = defineProps({
    device: Object as PropType<Device>,
  });

  const device = computed(() => props.device ?? ({} as Device));
</script>
<template>
  <div class="grid justify-content-between align-items-center">
    <LastSeen :timestamp="device.last_seen" />
    <div class="white-space-nowrap overflow-hidden text-overflow-ellipsis" v-if="device.availability === 'online'">
      <LinkQuality :value="device.exposes['linkquality']?.data" />
      <PowerSource :power_source="device.power_source" :value="getPowerSourceValue(device)" />
    </div>
    <div class="col-fixed white-space-nowrap overflow-hidden text-overflow-ellipsis" v-else>
      <Icon :icon="getOfflineIcon()" />
    </div>
  </div>
</template>
