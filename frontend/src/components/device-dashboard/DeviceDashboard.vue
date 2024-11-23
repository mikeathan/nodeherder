<script setup lang="ts">
import { computed } from 'vue';
import DeviceCard from './DeviceCard.vue';
import { Devices } from '@/types/device';
import { store } from '../../store/index';
import Card from 'primevue/card';

const devices = computed(
  () => store.getters['devices/listAll']() as Devices,
);
function getConnectivityIcon(status: string) {
  if (status === "Connected") return "pi pi-wifi";
  if (status === "Weak Signal") return "pi pi-wifi-slash";
  return "pi pi-times-circle";
}

</script>
<style scoped></style>

<template>

  <div class="p-grid p-nogutter">
    <div class="grid ">
      <div class="col-12 md:col-6 lg:col-3 xg:col-2" v-for="device in devices" :key="device.id">
        <DeviceCard :device="device" :key="device.id"></DeviceCard>
      </div>
    </div>
  </div>
</template>
