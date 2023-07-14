<script setup>
import LinkQuality from "../device/LinkQuality.vue";
import PowerSource from "../device/PowerSource.vue";
import LastSeen from "../device/LastSeen.vue";
const props = defineProps({
  payload: Object,
});
</script>

<template>
  <div class="card-footer">
    <div className="row justify-content-between flex-nowrap">
      <LastSeen :value="this.payload.stats.last_seen"></LastSeen>
      <div className="col-auto text-truncate">
        <LinkQuality :value="this.payload.stats.linkquality" />
        <PowerSource :power_source="this.payload.power_source" :value="this.payload.stats.battery" />
      </div>
    </div>
  </div>
</template>
