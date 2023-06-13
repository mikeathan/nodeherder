<script setup>
import { watch, ref, onMounted, onUnmounted, computed, reactive } from "vue";
import "../../assets/css/device.styles.css";
import DeviceFormatter from "../../modules/sensors/device-formatter";
const props = defineProps({
  payload: Object,
});

let formatter = ref(DeviceFormatter);

watch(
  () => props.payload,
  (newpayload) => {
    console.log("Watch props.payload update " + newpayload.name);
    formatter.update(newpayload);
  }
);

onMounted(() => {
  console.log("mounted: " + props.payload.name);
  formatter = new DeviceFormatter();
});

onUnmounted(() => {
  console.log("unmounted: " + props.payload.name);
  formatter.dispose();
});
</script>

<template>
  <div class="card-text">
    <span style="margin-right: 4.5rem">
      {{ formatter.lastSeen }}
    </span>
    <span :class="`fa fa-fw ${formatter.linkQualityIconClass}`" class="device-info-span"></span>
    <span class="device-info-span">
      {{ formatter.linkQuality }}
    </span>
    <span :class="`fa fa-fw ${formatter.batteryIconClass}`"></span>
  </div>
</template>
