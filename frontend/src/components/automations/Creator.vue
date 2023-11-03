<script setup>
import { useStore } from "vuex";
import { computed, ref } from "vue";
import Condition from "./Condition"
import { Expose } from "../../models/automation"

const exposes = ref({})

const store = useStore();
const selectedDevice = ref(null)
const selectedExpose = ref(null)

const devices = computed(() => {
    var devices = store.getters["devices/items"];
    return devices;
})



function exposeSelectionChanged(event) {

    if (event.target.value == null) {
        return;
    }
    if (exposes.value[event.target.value] != null) {
        return;
    }

    exposes.value[event.target.value] = new Expose(event.target.value)
    console.log("expose selected: ", event.target.value, " : ", exposes.value[event.target.value])
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
                            v-model="selectedDevice">
                            <option :value="null">Select device</option>
                            <option v-for="device in devices" :value="device" :key="device.id">
                                {{ device.friendly_name }}
                            </option>
                        </select>
                        <!-- <div class="input-group-btn">
                            <button type="button" class="btn btn-secondary text-nowrap"
                                @click="selectDevice(selectedDevice)">Select
                                device</button>
                        </div> -->

                    </div>
                </div>

                <div v-if="selectedDevice != null">
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
                        <Condition :Expose="exposes[selectedExpose]"></Condition>
                    </div>
                </div>

            </div>
        </div>
    </div>
</template>
