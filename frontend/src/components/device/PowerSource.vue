<script setup>
import { getPowerSourceIcon } from "../../modules/sensors/device-formatter";
const props = defineProps({
  power_source: String,
  value: Number,
});
</script>
<template>
  <!-- todo if power_source is not define set it to unknow same for value  -->
  <span :title="`${power_source} ${value}`" :class="`fa fa-fw ${getPowerSourceIcon(power_source, value)}`"></span>
</template>
