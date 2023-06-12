<script setup>
import { watch, ref, onMounted } from "vue";
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

var lastSeenTimerId = undefined;
let lastSeenUpdater = ref("");
watch(
  () => props.payload,
  (newpayload) => {
    console.log("Watch props.payload update " + newpayload.name);

    if (lastSeenTimerId != undefined) {
      clearInterval(lastSeenTimerId);
      console.log("clearInterval " + lastSeenTimerId);
    }
    lastSeenTimerId = setInterval(function () {
      lastSeenUpdater.value = formatLastSeen(newpayload);
    }, 1000);

    console.log("setInterval " + newpayload.name + " id: " + lastSeenTimerId);
  }
);
onMounted(() => {
  lastSeenUpdater.value = formatLastSeen(props.payload);
});
</script>

<template>
  <div class="card-text">
    <span style="margin-right: 4.5rem">
      {{ lastSeenUpdater }}
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
