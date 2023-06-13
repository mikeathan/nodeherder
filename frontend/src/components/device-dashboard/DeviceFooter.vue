<script setup>
import { watch, ref, onMounted, onUnmounted } from "vue";
import "../../assets/css/device.styles.css";
import PayloadFormatter from "../../modules/payload-formatter";
const props = defineProps({
  payload: Object,
});

var lastSeenTimerId = undefined;
let lastSeenUpdater = ref("");
let payloadFormatter = ref(PayloadFormatter);
watch(
  () => props.payload,
  (newpayload) => {
    console.log("Watch props.payload update " + newpayload.name);
    payloadFormatter.update(newpayload);
  }
);

onMounted(() => {
  console.log("mounted: " + props.payload.name);
  payloadFormatter = new PayloadFormatter();
  payloadFormatter.update(props.payload);
});
onUnmounted(() => {
  console.log("unmounted: " + props.payload.name);
  payloadFormatter.dispose();
});
</script>

<template>
  <div class="card-text">
    <span style="margin-right: 4.5rem">
      {{ payloadFormatter.lastSeen }}
    </span>
    <span
      :class="`fa fa-fw ${payloadFormatter.linkQualityIconClass}`"
      class="device-info-span"
    ></span>
    <span class="device-info-span">
      {{ payloadFormatter.linkQuality }}
    </span>
    <span :class="`fa fa-fw ${payloadFormatter.batteryIconClass}`"></span>
  </div>

  <!-- <div class="card-text">
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
  </div> -->
</template>
