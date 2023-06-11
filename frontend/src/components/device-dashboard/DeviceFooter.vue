<script setup>
import {
  formatLastSeen,
  getBatteryIcon,
  formatLinkQuality,
  getLinkQualityIcon,
} from "../../modules/sensors/device";
import "../../assets/css/device.styles.css";
const props = defineProps({
  payload: Object,
});
</script>

<template>
  <div class="card-text">
    <span style="margin-right: 4.5rem">
      {{ formatLastSeen(payload) }}
    </span>
    <span
      :class="`fa fa-fw ${getLinkQualityIcon()}`"
      class="device-info-span"
    ></span>
    <span class="device-info-span">
      {{ formatLinkQuality(payload) }}
    </span>
    <span :class="`fa fa-fw ${getBatteryIcon(payload)}`"></span>
  </div>
</template>
