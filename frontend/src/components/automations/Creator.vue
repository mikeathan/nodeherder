<script setup>
import { useStore } from "vuex";
import { computed, ref } from "vue";
import DeviceAutomation from "./DeviceAutomation"

const deviceTriggers = ref([])

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
    var json = event.toJson()
    console.log("DEBUG create", json);
    deviceTriggers.value.push(event)
    store.commit('automations/add', event);
    store.dispatch('automations/save', event.id);
}

</script>
<template>
    <div class="container-fluid p-0 h-100">
        <div class="card">
            <div class="card-body">
                <h3>Create new Automation</h3>
                <div class="col-xl-3 col-md-6 col-sm-3">
                    <select id="deviceSelector" style="text-align:center;" class="form-control "
                        :disabled="selectedDevice != ''" v-model="selectedDevice">
                        <option value="">Select device</option>
                        <option v-for="device in devices" :value="device.id" :key="device.id">
                            {{ device.friendly_name }}
                        </option>
                    </select>
                </div>
                <div class="col-xl-3 col-md-6 col-sm-3">
                    <DeviceAutomation :id="selectedDevice" @cancel="cancel" @create="create"></DeviceAutomation>
                </div>

                <!-- CREATOR DEBUG ----------------------<br>
                <div v-for="deviceTrigger in deviceTriggers">
                    <b>Device id:</b> {{ deviceTrigger.id }} <br>
                    <b>friendly_name:</b> {{ device.friendly_name }} <br>
                    <b>decription:</b> {{ deviceTrigger.description }}<br>
                    <b>enabled:</b> {{ deviceTrigger.enabled }}<br>
                    <div v-for="trigger in deviceTrigger.triggers">

                        <b>Trigger id: </b>{{ trigger.name }} <br>
                        <b>Conditions:</b>

                        <div v-for="condition in trigger.conditions">
                            {{ condition.name }} {{ condition.equality }} {{ condition.value }}
                        </div>
                    </div>
                    <hr>
                </div> -->
            </div>
        </div>
    </div>
</template>
