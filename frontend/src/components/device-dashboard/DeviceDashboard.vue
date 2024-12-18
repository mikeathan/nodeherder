<script setup lang="ts">
import { computed } from 'vue';
import DeviceCard from './DeviceCard.vue';
import { Devices } from '@/types/device';
import { store } from '../../store/index';

const devices = computed(
  () => store.getters['devices/listAll']() as Devices,
);

</script>
<style scoped>

.dashboard {
  background-color: var(--surface-card); /* Use the same background color as PrimeVue cards */
    height: 100vh; /* Optional: Full height */
    overflow: auto; /* Optional: Scroll if content exceeds */}</style>

<template>

    <div class="grid dashboard" >
      <div class="col-12 md:col-6 lg:col-3 xg:col-2" v-for="device in devices" :key="device.id">
        <DeviceCard :device="device" :key="device.id"></DeviceCard>
      </div>
    </div>
</template>
