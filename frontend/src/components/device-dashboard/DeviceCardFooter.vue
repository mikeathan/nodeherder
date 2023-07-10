<script setup>
import { watch, ref, onMounted, onUnmounted } from "vue";
import LinkQuality from "../device/LinkQuality.vue";
import PowerSource from "../device/PowerSource.vue";
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
  <div class="card-footer">
    <div className="row justify-content-between flex-nowrap">
      <div title="last update" className="col text-truncate">
        {{ lastSeen }}
      </div>

      <div className="col-auto text-truncate">
        <LinkQuality :value="this.payload.stats.linkquality" />
        <PowerSource
          :power_source="this.payload.power_source"
          :value="this.payload.stats.battery"
        />
      </div>
    </div>
  </div>
</template>
