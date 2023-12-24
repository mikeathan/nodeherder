<script setup>
import { useStore } from "vuex";
import { computed, ref } from "vue";
import DeviceAutomation from "./DeviceAutomation"
import Editor from "./Editor"

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
        <div class="card col-xl-6 col-md-5 col-sm-3">
            <div class="card-body">
                <h3>Create new Automation</h3>
                <div>
                    <Selector placeholder="Select device" :items="deviceList()" :value="selectedDevice" key="id"
                        alignment="left" @update:data="e => selectedDevice = e" :disabled="selectedDevice != ''"></Selector>
                </div>
                <div>
                    <Editor :id="selectedDevice"></Editor>
                </div>
            </div>
        </div>
    </div>
</template>
