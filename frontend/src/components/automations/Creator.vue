<script setup lang="ts">
import { computed, ref } from "vue";
import DeviceAutomation from "./DeviceAutomation.vue"
import Selection from "../input/Selection.vue"
import { store } from "../../store/index";
import { Devices } from "../../types/device";

const selectedDevice = ref("")
const deviceList = computed(() => {
    const devices = store.getters["devices/listAll"]() as Devices;
    return Object.assign({}, ...devices.map((d) => ({ [d.friendly_name]: d.id })))
})

function cancel(): void {
    selectedDevice.value = ""
}
</script>
<template>
    <div class="container-fluid p-0 h-100">

        <h3>Create new Automation</h3>
        <div class="col-xl-5 col-md-3" v-if="selectedDevice == ''">

            <Selection :value="selectedDevice" text="Select device" :disabled="selectedDevice != ''" size="normal"
                @updated="v => selectedDevice = v" :items="deviceList">
            </Selection>
        </div>
        <div v-else>
            <DeviceAutomation :id="selectedDevice" @cancel="cancel"></DeviceAutomation>
        </div>
    </div>
</template>
