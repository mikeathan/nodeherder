<script setup>
import DeviceFooter from "./DeviceCardFooter.vue";
import Sensor from "../device/Sensor.vue";
import { RouterLink } from "vue-router";

const props = defineProps({
  device: Object,
});

</script>
<style lang="css" scoped>
.v-list-item {
  padding-top: 0px;
  padding-bottom: 0px;

}
</style>
<template>

  <v-card outlined class="d-flex flex-column fill-height"
    :class="device.properties.availability == 'offline' ? 'disabled-card' : ''">
    <v-card-title>
      <RouterLink :to="`/devicepage/${device.id}`">{{
        device.friendly_name
      }}</RouterLink>
    </v-card-title>
    <v-card-text class="flex-grow-1">
      <v-list dense>
        <v-list-item dense v-for="(value, sensor) in device.exposes" :key="sensor">
          <v-list-item-title class="d-flex align-center">
            <Sensor :id="device.id" :expose="value" />
          </v-list-item-title>
        </v-list-item>
      </v-list>
    </v-card-text>

    <v-card-actions class="justify-space-between">
      <div class="d-flex align-center">
        <v-icon color="blue">mdi-wifi</v-icon>
        <span class="ml-1">66%</span>
      </div>
      <div class="d-flex align-center">
        <v-icon color="green">mdi-battery</v-icon>
        <span class="ml-1">100%</span>
      </div>
    </v-card-actions>
    <!-- <DeviceFooter :device="device"></DeviceFooter> -->
  </v-card>
  <!-- <div class="col-xl-3 col-lg-4 col-sm-6 col-12 d-flex">
    <div class="card flex-fill flex-shrink-1"
      :class="device.properties.availability == 'offline' ? 'disabled-card' : ''">
      <div class="card-header pb-0 d-flex justify-content-left">
        <h4>
          <RouterLink :to="`/devicepage/${device.id}`">{{
            device.friendly_name
            }}</RouterLink>
        </h4>
      </div>
      <div class="card-body row align-items-center">
        <div class="d-flex align-items-center" v-for="(value, sensor) in device.exposes">
          <Sensor :id="device.id" :expose="device.exposes[sensor]" />
        </div>
      </div>
      <DeviceFooter :device="device"></DeviceFooter>
    </div>
  </div> -->
</template>
