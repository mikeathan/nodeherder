<script setup>
import { useStore } from "vuex";
import { computed, watch, ref } from "vue";

import { ExposeTrigger, DeviceTrigger, Operators } from "../../models/automation"
import TriggerCondition from "./TriggerCondition"

const props = defineProps({
    id: String,
    trigger: Object
});

const store = useStore();
const conditions = ref([]);
const selectedExpose = ref("")
let exposeTriggers = {};
const enabled = ref(false)
const description = ref("")
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

const emit = defineEmits(['cancel', 'create'])

function exposeSelectionChanged(event) {

    var value = event.target.value;
    if (value == "" || exposeTriggers[value] != null) {
        return;
    }
    var exposeTrigger = new ExposeTrigger(value);
    exposeTriggers[value] = exposeTrigger;
}

watch(
    () => props.id,
    (newId) => {
        selectedExpose.value = null
    },
    { immediate: true }
);

function isSaveEnabled() {
    return Object.keys(exposeTriggers).length > 0 && description.value.length > 0
}
function getExposes() {
    return Object.keys(device.value.exposes)
}

function create() {

    var deviceTrigger = new DeviceTrigger()
    deviceTrigger.friendly_name = device.value.friendly_name;
    deviceTrigger.id = device.value.id
    deviceTrigger.enabled = enabled.value
    deviceTrigger.description = description.value

    for (const [key, item] of Object.entries(exposeTriggers)) {
        deviceTrigger.triggers.push(item)
    }

    reset();
    emit("create", deviceTrigger)
}

function reset() {

    conditions.value = []
    exposeTriggers = {}
    description.value = ""
    enabled.value = false
    emit("cancel")
}

function addCondition(event) {

    conditions.value.push(event);

    var exposeTrigger = exposeTriggers[selectedExpose.value];
    exposeTrigger.conditions.push(event)
}

function removeCondition(event) {
    var index = conditions.value.findIndex(item => item.idx === event);

    if (index != -1) {

        conditions.value.splice(index, 1);

        // remove from our device trigger cache
        for (const [key, item] of Object.entries(exposeTriggers)) {
            var index = item.conditions.findIndex(item => item.idx === event);
            if (index == -1) {
                continue
            }

            item.conditions.splice(index, 1);
            if (item.conditions.length == 0) {
                // delete the trigger
                delete exposeTriggers[key]
            }
        }
    }
}
// design 
//https://www.home-assistant.io/getting-started/automation/
</script>
<template>
    <div class="container-fluid p-0 h-100" v-if="device != null"> <!-- to fix condition-->

        <div class="col-3">
            <input type="text" class="form-control" name="name" id="name" v-model="description" placeholder="Enter a name"
                onfocus="this.placeholder = ''" onblur="this.placeholder='Enter a name'">
        </div>

        <div class="col-3">
            <div class="form-check form-switch">
                <input class="form-check-input" type="checkbox" role="switch" id="flexSwitchCheckDefault" v-model="enabled">
                <label class="form-check-label" for="flexSwitchCheckDefault">Enable</label>
            </div>
        </div>
        <div class="col-50 mt-3">
            <div class="btn-group">
                <button type="button" class="btn btn-light" :disabled="isSaveEnabled() == false" @click="create">
                    Save
                </button>

                <button type="button" class="btn btn-light" @click="reset">
                    Cancel
                </button>
            </div>
        </div>
        <!-- TODO: accordion here for exposes -->
        <!-- conditon value needs to be store as the expected type -->

        <br>
        <h5>Triggers</h5>

        <div class="col-3 mb-3">
            <select id="exposeSelector" style="text-align:center;" class="form-control" @change="exposeSelectionChanged"
                v-model="selectedExpose">
                <option :value="null">Select trigger</option>
                <option v-for="expose in device.exposes" :value="expose.name" :key="expose.name">
                    {{ expose.name }}
                </option>
            </select>
        </div>

        <div>
            <div class="row w-50" v-if="selectedExpose != null">
                <TriggerCondition :id="0" :exposes="getExposes()" :name="selectedExpose" :operator="Operators[0].value"
                    :data="''" @add="addCondition($event)">
                </TriggerCondition>
            </div>

            <div v-for="condition in conditions">
                <div class="row w-50">
                    <TriggerCondition :id="condition.idx" :name="condition.name" :operator="condition.equality"
                        :key="condition.idx" :data="condition.value" @remove="removeCondition($event)"></TriggerCondition>
                </div>
            </div>

        </div>
    </div>
</template>
