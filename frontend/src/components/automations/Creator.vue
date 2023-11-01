<script setup>
import { useStore } from "vuex";
import { computed, ref } from "vue";


const store = useStore();
const selectedDevice = ref()
const devices = computed(() => {
    var devices = store.getters["devices/items"];
    return devices;
})

function selectDevice() {
    console.log("selectDevice click ", selectedDevice);
    console.log("exposes: ");
    for (var e in selectedDevice.value.exposes) {
        console.log(e);
    }
    // show panel with exposes to choose from

    // TODO:
    // to fix make combo default to first item
    // click should call function and set the selecteditem
    // 
}

</script>
<template>
    <div class="container-fluid p-0 h-100">
        <div class="card">
            <div class="card-body">
                <h4>Create new Automation</h4>
                <div class="col-3">
                    <div class="d-flex">
                        <select id="deviceSelector" style="text-align:center;" class="form-control"
                            v-model="selectedDevice">
                            <option v-for="device in devices" :value="device" :key="device.id">
                                {{ device.friendly_name }}
                            </option>
                        </select>
                        <div class="input-group-btn">
                            <button type="button" class="btn btn-secondary text-nowrap"
                                @click="selectDevice(selectedDevice)">Select
                                device</button>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
