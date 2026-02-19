<script setup lang="ts">
  import { computed } from 'vue';
  import DeviceCard from './cards/DeviceCard.vue';
  import { Devices } from '@/types/device';
  import { store } from '../../store/index';
  const devices = computed(() => store.getters['hub/listAllDevices']() as Devices);
</script>

<style scoped>
  .dashboard-container {
    overflow-x: hidden;
  }
</style>

<template>
  <div class="dashboard-container">
    <div class="grid">
      <div class="col-12 md:col-6 lg:col-4 xl:col-3" v-for="device in devices" :key="device.id">
        <DeviceCard :device="device" :key="device.id"></DeviceCard>
      </div>
    </div>
  </div>
</template>
