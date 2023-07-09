<script setup>
import { getPowerSourceIcon } from "../../modules/sensors/device-formatter";
const props = defineProps({
  value: Number,
});
</script>
<template>
  <span :class="`fa fa-fw ${getPowerSourceIcon(value)}`"></span>
</template>
