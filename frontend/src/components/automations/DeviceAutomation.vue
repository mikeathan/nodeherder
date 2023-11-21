<script setup>
import { useStore } from "vuex";
import { computed, watch, ref, watchEffect } from "vue";

import { ExposeTrigger, DeviceTrigger, Operators, ActionTrigger } from "../../models/automation"
import TriggerCondition from "./TriggerCondition"
import TriggerActionNew from "./TriggerActionNew.vue";
import Trigger from "./Trigger"

const props = defineProps({
    id: String,
});

const store = useStore();
const conditions = ref([]);
const selectedExpose = ref("")
const currentTrigger = ref(null)
const exposeTriggers = ref({});
const enabled = ref(false)
const description = ref("")
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});


function exposeSelectionChanged(event) {
    var value = event.target.value;

    if (exposeTriggers.value[value] != null) {
        currentTrigger.value = exposeTriggers.value[value]
        return;
    }

    var exposeTrigger = new ExposeTrigger(value);
    exposeTriggers.value[value] = exposeTrigger;
    currentTrigger.value = exposeTriggers.value[value]
}

watch(
    () => props.id,
    (newId) => {
        selectedExpose.value = ""
        var exposeTrigger = new ExposeTrigger(selectedExpose.value);
        currentTrigger.value = exposeTrigger;
    },
    { immediate: true }
);

/* watch(
    () => selectedExpose,
    () => {
        if (selectedExpose.value == "" || exposeTriggers.value[selectedExpose.value] != null) {

            console.log("####WATCH exposeSelectionChanged return; ", currentTrigger.value);
            return;
        }

        var exposeTrigger = new ExposeTrigger(selectedExpose.value);
        currentTrigger.value = exposeTrigger;
        exposeTriggers.value[value] = exposeTrigger;
        console.log("####WATCH selectedexpose:", selectedExpose.value, " = ", currentTrigger.value);
    },
    { immediate: true }
); */

const emit = defineEmits(['cancel', 'create'])

function isSaveEnabled() {
    return Object.keys(exposeTriggers.value).length > 0 && description.value.length > 0
}

function create() {

    var deviceTrigger = new DeviceTrigger()
    deviceTrigger.friendlyName = device.value.friendly_name;
    deviceTrigger.id = device.value.id
    deviceTrigger.enabled = enabled.value
    deviceTrigger.description = description.value

    for (const [key, item] of Object.entries(exposeTriggers.value)) {
        deviceTrigger.triggers.push(item)
    }

    reset();
    emit("create", deviceTrigger)
}


function reset() {

    conditions.value = []
    exposeTriggers.value = {}
    description.value = ""
    enabled.value = false
    emit("cancel")
}

function addAction(event) {

    var exposeTrigger = exposeTriggers.value[selectedExpose.value]
    exposeTrigger.action = event;
}

function removeAction(event) {
    var exposeTrigger = exposeTriggers.value[selectedExpose.value]
    exposeTrigger.action = null;
}

function addCondition(event) {
    conditions.value.push(event);

    var exposeTrigger = exposeTriggers.value[selectedExpose.value]
    exposeTrigger.conditions.push(event)
}

function removeCondition(event) {
    var index = conditions.value.findIndex(item => item.idx === event);

    if (index != -1) {

        conditions.value.splice(index, 1);

        // remove from our device trigger cache
        for (const [key, item] of Object.entries(exposeTriggers.value)) {
            var index = item.conditions.findIndex(item => item.idx === event);
            if (index == -1) {
                continue
            }

            item.conditions.splice(index, 1);
        }
    }
}
// design
//https://www.home-assistant.io/getting-started/automation/
</script>

<style scoped>
.inputName {
    border: 0;
    outline: 0;
    background: transparent;
    border-bottom: 1px solid #e5e5e5;
    border-radius: 0
}

.custom-control-input {
    transform: scale(1.4);
}
</style>
<template>
    <div class="container-fluid p-0 h-100" v-if="device != null"> <!-- to fix condition-->

        <!-- <div class="col-3">
            <input type="text" class="form-control" name="name" id="name" v-model="description" placeholder="Enter a name"
                onfocus="this.placeholder = ''" onblur="this.placeholder='Enter a name'">
        </div> -->
        <div class="col-3 pt-3 pb-4">
            <label class="form-check-label" for="name">Name</label>

            <input type="text" class="form-control inputName" name="name" id="name" v-model="description"
                placeholder="New automation" onfocus="this.placeholder = ''" onblur="this.placeholder='New Automation 1'">
        </div>

        <div class="col-3 pb-3">
            <div class=" form-check form-switch ms-2">
                <label class="form-check-label ms-3">Enable</label>

                <input class="form-check-input custom-control-input" type="checkbox" role="switch"
                    id="flexSwitchCheckDefault" v-model="enabled">
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

        <h5>Triggers</h5>

        <div class="col-3 mb-3">
            <select id="exposeSelector" style="text-align:center;" class="form-control " @change="exposeSelectionChanged"
                v-model="selectedExpose">
                <option value="">Select trigger</option>
                <option v-for="expose in device.exposes" :value="expose.name" :key="expose.name">
                    {{ expose.name }}
                </option>
            </select>
        </div>

        <div>
            <Trigger :id="props.id" :trigger="currentTrigger" @addAction="addAction" @removeAction="removeAction"
                @addCondition="addCondition" @removeCondition="removeCondition">
            </Trigger>
        </div>

    </div>
</template>
