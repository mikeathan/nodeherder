<script setup lang="ts">
import { computed } from 'vue';
import DeviceCard from './DeviceCard.vue';
import { Devices } from '@/types/device';
import { store } from '../../store/index';
import Card from 'primevue/card';

const devices = computed(
  () => store.getters['devices/listAll']() as Devices,
);
</script>
<!-- sm="6"
md="4"
lg="3"
xl="2" -->
<template>
  test
  <div class="dashboard">
    <div v-for="device in devices" :key="device.id">
      <Card title="Device {{ device.friendly_name }}">
        <template #header>
          <img
            alt="user header"
            src="/images/usercard.png" />
        </template>
        <template #title>{{
          device.friendly_name
        }}</template>
        <template #subtitle>{{ device.id }}</template>
        <template #content>
          <p class="m-0">
            Lorem ipsum dolor sit amet, consectetur
            adipisicing elit. Inventore sed consequuntur
            error repudiandae numquam deserunt quisquam
            repellat libero asperiores earum nam nobis,
            culpa ratione quam perferendis esse, cupiditate
            neque quas!
          </p>
        </template>
        <template #footer>
          <div class="flex gap-4 mt-1">
            <Button
              label="Cancel"
              severity="secondary"
              outlined
              class="w-full" />
            <Button label="Save" class="w-full" />
          </div>
        </template>
      </Card>
    </div>
  </div>
  <!-- <v-row align="stretch" justify="start" dense>
    <v-col
      cols="3"
      v-for="(device, index) in devices"
      :key="index"
      class="d-flex">
      <DeviceCard
        :device="device"
        :key="device.id"></DeviceCard>
    </v-col>
  </v-row> -->
</template>
