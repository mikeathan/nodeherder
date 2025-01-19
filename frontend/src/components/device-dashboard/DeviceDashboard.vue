<script setup lang="ts">
  import { computed } from 'vue';
  import DeviceCard from './DeviceCard.vue';
  import { Devices } from '@/types/device';
  import { store } from '../../store/index';

  const devices = computed(() => store.getters['hub/listAllDevices']() as Devices);
</script>
<style scoped>
  /* .dashboard {
  background-color: var(--p-card-background);
  height: 100vh;
  overflow: auto;
} */
</style>

<template>
  <div class="grid">
    <div class="col-12 md:col-6 lg:col-3 xg:col-2" v-for="device in devices" :key="device.id">
      <DeviceCard :device="device" :key="device.id"></DeviceCard>
    </div>
  </div>
</template>
