<script setup>
import { watch, ref, onMounted, onUnmounted } from "vue";
import "../../assets/css/device.styles.css";
import DeviceFormatter from "../../modules/sensors/device-formatter";
const props = defineProps({
  payload: Object,
});

let formatter = new DeviceFormatter();
let lastSeen = ref("");

watch(
  () => props.payload,
  (newpayload) => {
    console.log("Watch props.payload update " + newpayload.name);
    formatter.formatPayload(newpayload, (v) => {
      lastSeen.value = v;
    });
    lastSeen.value = formatter.lastSeen;
  },
  { immediate: true }
);

onMounted(() => {
  console.log("mounted: " + props.payload.name);
});

onUnmounted(() => {
  console.log("unmounted: " + props.payload.name);
  formatter.dispose();
});
</script>

<template>
  <div class="card-text">
    <span style="margin-right: 4.5rem">
      {{ lastSeen }}
    </span>
    <span
      :class="`fa fa-fw ${formatter.linkQualityIconClass}`"
      class="device-info-span"
    ></span>
    <span class="device-info-span">
      {{ formatter.linkQuality }}
    </span>
    <span :class="`fa fa-fw ${formatter.batteryIconClass}`"></span>
  </div>
</template>
