<script setup lang="ts">
  import { computed } from 'vue';
  import DeviceCard from './cards/DeviceCard.vue';
  import { Devices } from '@/types/device';
  import { store } from '../../store/index';
  const baseUrl = import.meta.env.VITE_API_BASE_URL;
  const devices = computed(() => store.getters['hub/listAllDevices']() as Devices);

  const loginWithGoogle = () => {
    window.location.href = `${baseUrl}/api/auth/google/login`;
  };
</script>

<style scoped>
  /* .dashboard {
  background-color: var(--p-card-background);
  height: 100vh;
  overflow: auto;
} */
</style>

<template>
  <button @click="loginWithGoogle">Login with Google</button>

  <div class="grid">
    <div class="col-12 md:col-6 lg:col-3 xg:col-2" v-for="device in devices" :key="device.id">
      <DeviceCard :device="device" :key="device.id"></DeviceCard>
    </div>
  </div>
</template>
