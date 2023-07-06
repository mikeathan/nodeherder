<script setup>
import { watch, ref, onMounted, onUnmounted } from "vue";
import "../../assets/css/device.styles.css";
import "../../assets/css/dark.css";
import DeviceFormatter from "../../modules/sensors/device-formatter";
const props = defineProps({
  payload: Object,
});

let formatter = new DeviceFormatter();
let lastSeen = ref("");

watch(
  () => props.payload,
  (newpayload) => {
    console.log("Watch props.payload update");
    formatter.formatPayload(newpayload.stats, (v) => {
      lastSeen.value = v;
    });
    lastSeen.value = formatter.lastSeen;
  },
  { immediate: true }
);

onMounted(() => {
  console.log("mounted");
});

onUnmounted(() => {
  console.log("unmounted");
  formatter.dispose();
});
</script>

<template>
  <div class="card-text">
    <div className="row justify-content-between flex-nowrap">
      <div title="last update" className="col text-truncate">
        {{ lastSeen }}
      </div>

      <div className="col-auto text-truncate">
        <span key="linkquality" className="me-1">
          <i :class="`fa fa-fw ${formatter.linkQualityIconClass}`"></i>
          {{ formatter.linkQuality }}
        </span>
        <span :class="`fa fa-fw ${formatter.powerSourceIconClass}`"></span>
      </div>
    </div>
  </div>
</template>
