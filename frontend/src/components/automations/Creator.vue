<script setup lang="ts">
import { computed, ref } from "vue";
import DeviceAutomation from "./DeviceAutomation.vue"
import Selector from "../input/Selector.vue"
import { store } from "../../store/index";
import { Devices } from "../../types/device";

const selectedDevice = ref("")
const devices = computed<Devices>(() => {
    return store.getters["devices/listAll"] as Devices;
})

function cancel() {
    selectedDevice.value = ""
}


function deviceList() {

    const filtered = devices.value.map((d) => ({ [d.friendly_name]: d.id }))
    return filtered
}

</script>
<template>
    <div class="container-fluid p-0 h-100">

        <h3>Create new Automation</h3>
        <div class="col-xl-5 col-md-3" v-if="selectedDevice == ''">
            <Selector placeholder="Select device" :items="deviceList()" :value="selectedDevice" alignment="left"
                @update:data="e => selectedDevice = e" :disabled="selectedDevice != ''"></Selector>
        </div>
        <div v-else>
            <DeviceAutomation :id="selectedDevice" @cancel="cancel"></DeviceAutomation>
        </div>
    </div>
</template>
