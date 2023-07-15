<script setup>
import { ref } from 'vue';

const props = defineProps({
  status: String,
  showAnimation: ref(true)
});
</script>
<template>
  <div key="Availability" v-if="status === 'online'">
    <i class="fa fa-circle fa-xs" style="color: green"></i>
  </div>
  <div v-else>
    <i class="fa fa-circle fa-fade fa-xs" style="--fa-animation-duration: 2s; --fa-fade-opacity: 0.2; color: red"></i>
  </div>
</template>
