<script setup>
import { useStore } from "vuex";
import { computed, ref } from "vue";
import Expose from "./Expose"

import { ExposeTrigger } from "../../models/automation"

const props = defineProps({
    trigger: Object,
});

const store = useStore();
const exposes = ref({})
const selectedExpose = ref(null)

const device = computed(() => {
    return store.getters["devices/find"](props.trigger.id);
});

function exposeSelectionChanged(event) {

    if (event.target.value == null ||
        exposes.value[event.target.value] != null) {
        return;
    }

    exposes.value[event.target.value] = new ExposeTrigger(event.target.value)
    console.log("expose selected: ", event.target.value, " : ", exposes.value[event.target.value])
}

</script>
<template>
    <div v-if="device != null" class="container-fluid p-0 h-100">

        <div class="col-3">
            <select id="exposeSelector" style="text-align:center;" class="form-control"
                @click="exposeSelectionChanged($event)" v-model="selectedExpose">

                <option :value="null">Select expose</option>
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
</template>
