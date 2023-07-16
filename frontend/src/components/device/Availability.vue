<script setup>
import { ref } from 'vue';

const props = defineProps({
  status: String,
  showAnimation: ref(true)
});
</script>
<template>
  <div key="Availability" v-if="status === 'online'">
    <i class="fa fa-circle fa-2xs" style="color: rgb(57, 217, 57)"></i>
  </div>
  <div v-else>
    <i class="fa fa-circle fa-2xs" style="color: rgb(243, 83, 83)"></i>
  </div>
</template>
