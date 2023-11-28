<script setup>
import { useStore } from "vuex";
import { computed, ref } from "vue";
import DeviceAutomation from "./DeviceAutomation"
import { useRouter } from 'vue-router'
const deviceTriggers = ref([])
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
    deviceTriggers.value.push(event)
    store.dispatch('automations/save', event);
    router.push("/viewer")
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
            </div>
        </div>
    </div>
</template>
