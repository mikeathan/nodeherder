<script setup>
import { useStore } from "vuex";
import { computed, watch, ref, onBeforeMount } from "vue";

import { ExposeTrigger, DeviceTrigger } from "../../models/automation"
import Conditions from "./Conditions"

const props = defineProps({
    id: String
});

const store = useStore();
const exposes = ref({})

const deviceTrigger = ref(new DeviceTrigger(props.id))
const selectedExpose = ref(null)
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

function exposeSelectionChanged(event) {

    var value = event.target.value;
    if (value == "") {
        //selectedExpose.value = null; // maybe refactor ????
        return;
    }
    deviceTrigger.value.triggers
    selectedExpose.value = event.target.value; // maybe refactor ????
    if (exposes.value[value] != null) {
        console.log(value, " exists with conditions: ", exposes.value[value].Conditions.length)
        return
    }
    exposes.value[value] = new ExposeTrigger(value)
    console.log("expose added: ", value, " : ", exposes.value[value], "conditions: ", exposes.value[value].Conditions.length)
}

watch(
    () => props.id,
    (t) => {
        selectedExpose.value = null
    },
    { immediate: true }
);

</script>
<template>
    <div v-if="device != null" class="container-fluid p-0 h-100">

        <div class="col-3">
            <select id="exposeSelector" style="text-align:center;" class="form-control" @change="exposeSelectionChanged">
                <option value="">Select expose</option>
                <option v-for="expose in device.exposes" :value="expose.name" :key="expose.name">
                    {{ expose.name }}
                </option>
            </select>
        </div>
        <br>
        Device: {{ props.id }} - {{ selectedExpose }}

        <div>
            <!-- <Expose :id="props.id" :expose="exposes[selectedExpose]"></Expose> -->
            <h4>Condition</h4>
            <div v-if="selectedExpose != null" class="col">
                <Conditions :expose="exposes[selectedExpose]">
                </Conditions>
            </div>
        </div>
    </div>
    <div v-else>
    </div>
</template>
