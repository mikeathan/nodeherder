<script setup>
import { useStore } from "vuex";
import { computed, ref } from "vue";
import DeviceAutomation from "./DeviceAutomation"
import Selector from "../input/Selector.vue"

import { useRouter } from 'vue-router'
const router = useRouter()
const store = useStore();
const selectedDevice = ref("")

const devices = computed(() => {
    var devices = store.getters["devices/items"];
    return devices;
})

function cancel() {
    selectedDevice.value = ""
}

function create(event) {
    //deviceTriggers.value.push(event) ?? we dont needthat 
    store.dispatch('automations/save', event);
    router.push("/viewer")
}

function deviceList() {


    // todo:
    //var result = Object.keys(obj).map((key) => [key, obj[key]]);
    var list = {}
    for (const [key, device] of Object.entries(devices.value)) {
        list[device.friendly_name] = device.id
    }
    return list
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
