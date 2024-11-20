<script setup lang="ts">
import { computed } from "vue";
import DeviceCard from "./DeviceCard.vue";
import { Devices } from "@/types/device";
import { store } from "../../store/index";

const devices = computed(() => store.getters["devices/listAll"]() as Devices);

</script>

<template>
  <v-row>
    <DeviceCard v-for="item in devices" :device="item" :key="item.id"></DeviceCard>
  </v-row>
  <!-- <div className="row">
    <DeviceCard v-for="item in devices" :device="item" :key="item.id"></DeviceCard>
  </div> -->
</template>
