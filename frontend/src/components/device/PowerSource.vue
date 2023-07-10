<script setup>
import { getPowerSourceIcon } from "../../modules/sensors/device-formatter";
const props = defineProps({
  power_source: String,
  value: Number,
});
</script>
<template>
  <span :class="`fa fa-fw ${getPowerSourceIcon(power_source, value)}`"></span>
</template>
