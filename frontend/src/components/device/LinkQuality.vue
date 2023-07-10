<script setup>
const props = defineProps({
  value: Number,
});
</script>
<template>
  <span key="linkquality" className="me-1" v-if="value !== undefined">
    <i :class="`fa fa-fw fa-signal fa-fw`"></i>
    {{ value + " LQI" }}
  </span>
</template>
