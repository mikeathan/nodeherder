<script setup>
import { ref } from 'vue';

const props = defineProps({
  status: String,
  showAnimation: ref(true)
});
</script>
<template>
  <i class="fa fa-circle" style="color: green"></i>
  <div key="Availability" v-if="status === 'offline'">
    <i class="fa-solid text-danger fa-triangle-exclamation fa-fade"></i>
    Offline
  </div>
</template>
