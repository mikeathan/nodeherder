<script setup>
import { watch, ref, onMounted, onUnmounted } from "vue";
import "../../assets/css/device.styles.css";
import DeviceFormatter from "../../modules/sensors/device-formatter";
const props = defineProps({
  payload: Object,
});

let deviceFormatter = ref(DeviceFormatter);
watch(
  () => props.payload,
  (newpayload) => {
    console.log("Watch props.payload update " + newpayload.name);
    deviceFormatter.update(newpayload);
  }
);

onMounted(() => {
  console.log("mounted: " + props.payload.name);
  deviceFormatter = new DeviceFormatter();
});

onUnmounted(() => {
  console.log("unmounted: " + props.payload.name);
  deviceFormatter.dispose();
});
</script>

<template>
  <div class="card-text">
    <span style="margin-right: 4.5rem">
      {{ deviceFormatter.lastSeen }}
    </span>
    <span
      :class="`fa fa-fw ${deviceFormatter.linkQualityIconClass}`"
      class="device-info-span"
    ></span>
    <span class="device-info-span">
      {{ deviceFormatter.linkQuality }}
    </span>
    <span :class="`fa fa-fw ${deviceFormatter.batteryIconClass}`"></span>
  </div>
</template>
