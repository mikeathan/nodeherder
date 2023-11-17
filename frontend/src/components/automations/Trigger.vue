<script setup>
import { useStore } from "vuex";
import { computed, watch, ref } from "vue";

import { ExposeTrigger, DeviceTrigger, Operators, ActionTrigger } from "../../models/automation"
import TriggerCondition from "./TriggerCondition"
import TriggerActionNew from "./TriggerActionNew.vue";

const props = defineProps({
    id: String,
});

const store = useStore();
const conditions = ref([]);
const selectedExpose = ref("")
const selectedAction = ref(null)
let exposeTriggers = {};
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

function getExposes() {
    return Object.keys(device.value.exposes)
}

function exposeSelectionChanged(event) {

    var value = event.target.value;
    if (value == "" || exposeTriggers[value] != null) {
        return;
    }
    var exposeTrigger = new ExposeTrigger(value);
    exposeTriggers[value] = exposeTrigger;

    console.log("exposeSelectionChanged", exposeTriggers[value]);
}

watch(
    () => props.id,
    (newId) => {
        selectedExpose.value = null
        selectedAction.value = null
    },
    { immediate: true }
);

const featureDevices = computed(() => {
    var devices = store.getters["devices/items"];
    // find exposes with properties
    // let all = items.filter(item=> item.age==='18')
    //     return devices;
    // });
    var list = []
    for (const [key, device] of Object.entries(devices)) {
        for (const [key, expose] of Object.entries(device.exposes)) {
            if (expose.properties != undefined) {
                list.push(device)
                break;
            }
        }
    }

    return list;

});

function addAction(event) {
    console.log("addAction ", event)

    var exposeTrigger = exposeTriggers[selectedExpose.value]
    exposeTrigger.action = event;

    console.log("addAction :", exposeTrigger.action)
}

function removection(event) {
    console.log("removection ", event)
}

//const emit = defineEmits(['cancel', 'create'])
</script>

<template>
    <div class="container-fluid p-0 h-100">
        <!-- TODO: accordion here for exposes -->
        <!-- conditon value needs to be store as the expected type -->
        <!-- Save button shouls navigate to automation viewer -->
        <!-- check url design above for styling of creator text input -->
        <!-- if automation for device exists message user else we overwrite it -->

        <br>
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
            ------ Accordion HERE --------------
            <h5>Conditions</h5>
            <!-- <div class="row w-50" v-if="selectedExpose != null">

            <TriggerCondition :id="0" :exposes="getExposes()" :name="selectedExpose" :operator="Operators[0].value"
                :data="''" @add="addCondition($event)">
            </TriggerCondition>
        </div>

        <div v-for="condition in conditions">
            <div class="row w-50">
                <TriggerCondition :id="condition.idx" :name="condition.name" :operator="condition.equality"
                    :key="condition.idx" :data="condition.value" @remove="removeCondition($event)"></TriggerCondition>
            </div>
        </div> -->
        </div>

        <br>
        <h5>Actions</h5>
        <div class="col-3 mb-3">
            <select id="featureDeviceSelector" style="text-align:center;" class="form-control" v-model="selectedAction"
                :disabled="selectedExpose == null">
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

            <div class="row w-50" v-if="selectedExpose != null && exposeTriggers[selectedExpose] != NonNullable">
                <TriggerActionNew :id="exposeTriggers[selectedExpose].id" :action="exposeTriggers[selectedExpose].action"
                    @add="addAction"></TriggerActionNew>
            </div>
        </div>
    </div>
</template>