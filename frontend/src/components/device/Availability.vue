<script setup>
const props = defineProps({
  status: String,
});
</script>
<template>
  <small key="Availability" v-if="status === 'offline'">
    <i class="fa-solid text-danger fa-triangle-exclamation fa-fade"></i>
    Offline
  </small>
</template>
