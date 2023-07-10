<script setup>
import DeviceFooter from "./DeviceCardFooter.vue";
import Availability from "../device/Availability.vue";
import Sensor from "../device/Sensor.vue";

const props = defineProps({
  device: Object,
});
// $('.close-icon').on('click', function () {
//   $(this).closest('.card').fadeOut();
// })
</script>
<template>
  <div class="col-xl-3 col-lg-4 col-sm-6 col-12 d-flex">
    <div class="card flex-fill flex-shrink-1">

      <span class="pull-right clickable close-icon" data-effect="fadeOut"><i class="fa fa-times"></i></span>


      <div class="card-header pb-0 d-flex justify-content-between">
        {{ device.id }}
        <Availability :status="device.stats.availability" />
      </div>

      <div class="card-body row align-items-center">
        <div class="d-flex align-items-center" v-for="(value, sensor) in device.sensors">
          <Sensor :name="sensor" :value="value" />
        </div>
      </div>
      <DeviceFooter :payload="device"></DeviceFooter>
    </div>
  </div>
</template>
