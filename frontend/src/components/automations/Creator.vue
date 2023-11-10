<script setup>
import { useStore } from "vuex";
import { computed, ref } from "vue";
import DeviceAutomation from "./DeviceAutomation"

import { DeviceTrigger } from "../../models/automation"

const deviceTriggers = ref({})

const store = useStore();
const selectedDevice = ref("")

const devices = computed(() => {
    var devices = store.getters["devices/items"];
    return devices;
})
function deviceSelectionChanged(event) {

    var value = event.target.value;

    if (value == "") {
        selectedDevice.value = null; // maybe refactor ????
        return;
    }
    selectedDevice.value = value
    if (deviceTriggers.value[value] != null) {
        console.log(value, " already init");
        return;
    }

    deviceTriggers.value[value] = new DeviceTrigger(value);
    console.log("device selected: ", value, "-", selectedDevice.value);
}

function cancel() {
    selectedDevice.value = ""
}

function exposeSelectionChanged(event) {

    //     if (event.target.value == null ||
    //         //exposes.value[event.target.value] != null) {
    //         return;
    // }

    // exposes.value[event.target.value] = new DeviceExpose(event.target.value)
    //console.log("expose selected: ", event.target.value, " : ", exposes.value[event.target.value])
}

</script>
<template>
    <div class="container-fluid p-0 h-100">
        <div class="card">
            <div class="card-body">
                <h3>Create new Automation</h3>
                <div class="col-3">
                    <div class="d-flex">
                        <select id="deviceSelector" style="text-align:center;" class="form-control"
                            :disabled="selectedDevice != ''" v-model="selectedDevice">
                            <option value="">Select device</option>
                            <option v-for="device in devices" :value="device.id" :key="device.id">
                                {{ device.friendly_name }}
                            </option>
                        </select>
                    </div>
                </div>
                <div>
                    <DeviceAutomation :id="selectedDevice" @cancel="cancel"></DeviceAutomation>
                    <!-- <div class="col-3">
                        <select id="exposeSelector" style="text-align:center;" class="form-control"
                            @click="exposeSelectionChanged($event)">

                            <option :value="null">Select expose</option>
                            <option v-for="expose in selectedDevice.exposes" :value="expose.name" :key="expose.name">
                                {{ expose.name }}
                            </option>
                        </select>
                    </div>
                    <br> -->
                </div>
                <!-- <div v-if="selectedDevice != null">
                    <div class="col-3">
                        <select id="exposeSelector" style="text-align:center;" class="form-control"
                            @click="exposeSelectionChanged($event)" v-model="selectedExpose">

                            <option :value="null">Select expose</option>
                            <option v-for="expose in selectedDevice.exposes">
                                {{ expose.name }}
                            </option>
                        </select>
                    </div>
                    <br>

                    <div v-if="selectedExpose != null">
                        <ExposeTrigger :expose="exposes[selectedExpose]"></ExposeTrigger>
                    </div>
                </div> -->

            </div>
        </div>
    </div>
</template>
