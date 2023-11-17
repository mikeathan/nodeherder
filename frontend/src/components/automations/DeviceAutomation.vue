<script setup>
import { useStore } from "vuex";
import { computed, watch, ref } from "vue";

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
const selectedAction = ref(null)

let exposeTriggers = {};
const enabled = ref(false)
const description = ref("")
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});


const emit = defineEmits(['cancel', 'create'])

function isSaveEnabled() {
    return Object.keys(exposeTriggers).length > 0 && description.value.length > 0
}

function create() {

    var deviceTrigger = new DeviceTrigger()
    deviceTrigger.friendlyName = device.value.friendly_name;
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

        <Trigger :id="props.id"></Trigger>
        <!-- <br>
        <h5>Triggers</h5>

        <div class="col-3 mb-3">
            <select id="exposeSelector" style="text-align:center;" class="form-control " @change="exposeSelectionChanged"
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

        <br>
        <h5>Actions</h5>
        the action is linked to the selected trigger !!!!!
        <div class="col-3 mb-3">
            <select id="featureDeviceSelector" style="text-align:center;" class="form-control" v-model="selectedAction">
                <option :value="null">Select device</option>
                <option v-for="device in featureDevices" :value="device.id" :key="device.friendly_name">
                    {{ device.friendly_name }}
                </option>
            </select>
        </div>
        <div>
            <div class="row w-50" v-if="selectedAction != null">
                <TriggerActionNew :id="selectedAction" :action="null" @add="addAction"></TriggerActionNew>
            </div>

            <div class="row w-50" v-if="selectedExpose != null">
                {{ exposeTriggers[selectedExpose] }}
          
            </div>
        </div> -->

    </div>
</template>
