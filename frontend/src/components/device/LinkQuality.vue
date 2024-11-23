<script setup lang="ts">
const props = defineProps({
  value: Number
});
</script>
<template>
  <span key="linkquality" className="me-" v-if="value !== undefined">
    <i :class="`fa fa-fw fa-signal fa-fw`"></i>
    {{ value + " LQI" }}
  </span>
</template>
