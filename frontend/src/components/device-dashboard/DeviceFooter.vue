<script setup>
import { onMounted } from "vue";
import "../../assets/css/device.styles.css";
import {
  formatLastSeen,
  getBatteryIcon,
  formatLinkQuality,
  getLinkQualityIcon,
} from "../../modules/sensors/device";

const props = defineProps({
  payload: Object,
});



onMounted(() => {
  console.log("device footer mounted");

});
var lastSeenTimerId = undefined;
function lastSeen(payload) {
  return formatLastSeen(payload)
};

</script>

<template>
  <div class="card-text">
    <span style="margin-right: 4.5rem">
      {{ lastSeen(payload) }}
    </span>
    <span :class="`fa fa-fw ${getLinkQualityIcon()}`" class="device-info-span"></span>
    <span class="device-info-span">
      {{ formatLinkQuality(payload) }}
    </span>
    <span :class="`fa fa-fw ${getBatteryIcon(payload)}`"></span>
  </div>
</template>
