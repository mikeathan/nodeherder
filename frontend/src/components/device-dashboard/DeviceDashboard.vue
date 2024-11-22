<script setup lang="ts">
import { computed } from 'vue';
import DeviceCard from './DeviceCard.vue';
import { Devices } from '@/types/device';
import { store } from '../../store/index';

const devices = computed(
  () => store.getters['devices/listAll']() as Devices,
);
</script>
<!-- sm="6"
md="4"
lg="3"
xl="2" -->
<template>
  <v-row align="stretch" justify="start" dense>
    <v-col
      cols="3"
      v-for="(device, index) in devices"
      :key="index"
      class="d-flex">
      <DeviceCard
        :device="device"
        :key="device.id"></DeviceCard>
    </v-col>
  </v-row>
</template>
