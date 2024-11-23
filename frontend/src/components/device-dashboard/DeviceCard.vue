<script setup>
import DeviceFooter from './DeviceCardFooter.vue';
import Sensor from '../device/Sensor.vue';
import { RouterLink } from 'vue-router';
import Card from 'primevue/card';

const props = defineProps({
  device: Object,
});
</script>
<style lang="css" scoped>

</style>
<template>
<Card style="width: 25rem; overflow: hidden">
    <template #header>
        <img alt="user header" src="/images/usercard.png" />
    </template>
    <template #title>Advanced Card</template>
    <template #subtitle>Card subtitle</template>
    <template #content>
        <p class="m-0">
            Lorem ipsum dolor sit amet, consectetur adipisicing elit. Inventore sed consequuntur error repudiandae numquam deserunt quisquam repellat libero asperiores earum nam nobis, culpa ratione quam perferendis esse, cupiditate neque
            quas!
        </p>
    </template>
    <template #footer>
        <div class="flex gap-4 mt-1">
            <Button label="Cancel" severity="secondary" outlined class="w-full" />
            <Button label="Save" class="w-full" />
        </div>
    </template>
</Card>
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
