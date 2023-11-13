<script setup>
import { useStore } from "vuex";
import { computed, ref } from "vue";
import DeviceAutomation from "./DeviceAutomation"

import { DeviceTrigger } from "../../models/automation"

const deviceTriggers = ref([])

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

    console.log("device selected: ", value, "-", selectedDevice.value);
}

function cancel() {
    selectedDevice.value = ""
}

function create(event) {
    console.log("create ", event);
    deviceTriggers.value.push(event)
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
                    <DeviceAutomation :id="selectedDevice" @cancel="cancel" @create="create"></DeviceAutomation>

                </div>

                CREATOR DEBUG ----------------------<br>
                <div v-for="deviceTrigger in deviceTriggers">
                    <b>Device id:</b> {{ deviceTrigger.id }} <br>
                    <!-- <b>friendly_name:</b> {{ device.friendly_name }} <br> -->
                    <b>decription:</b> {{ deviceTrigger.description }}<br>

                    <div v-for="trigger in deviceTrigger.triggers">

                        <b>Trigger id: </b>{{ trigger.name }} <br>
                        <b>Conditions:</b>

                        <div v-for="condition in trigger.conditions">
                            {{ condition.name }} {{ condition.operator }} {{ condition.data }}
                        </div>
                    </div>
                    <hr>
                </div>
            </div>
        </div>
    </div>
</template>
