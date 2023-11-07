<script setup>
import { useStore } from "vuex";
import { computed, onUpdated, ref } from "vue";
import Expose from "./Expose"

import { ExposeTrigger } from "../../models/automation"

const props = defineProps({
    trigger: Object,
});

const store = useStore();
const exposes = ref([props.trigger.triggers])
const selectedExpose = ref(null)

const device = computed(() => {
    return store.getters["devices/find"](props.trigger.id);
});


function exposeSelectionChanged(event) {

    var value = event.target.value;
    if (value == "") {
        selectedExpose.value = null; // maybe refactor ????
        return;
    }

    selectedExpose.value = event.target.value;
    if (exposes.value[value] != null) {
        console.log(value, " exists with conditions: ", exposes.value[value].Conditions.length)
        return
    }
    exposes.value[value] = new ExposeTrigger(value)
    console.log("expose added: ", value, " : ", exposes.value[value], "conditions: ", exposes.value[value].Conditions.length)
}

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

        <div v-if="selectedExpose != null">
            <Expose :expose="exposes[selectedExpose]"></Expose>
        </div>
    </div>
    <div v-else>
    </div>
</template>
