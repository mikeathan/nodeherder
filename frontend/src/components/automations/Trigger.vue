<script setup>
import { useStore } from "vuex";
import { computed, watch, ref, watchEffect } from "vue";

import { ExposeTrigger, DeviceTrigger, Operators, ActionTrigger } from "../../models/automation"
import TriggerCondition from "./TriggerCondition"
import TriggerActionNew from "./TriggerActionNew.vue";

const props = defineProps({
    id: String,
    trigger: Object,
});

const store = useStore();
const conditions = ref([]);
const selectedAction = ref(null)

const exposeName = ref("")
const currentAction = ref()
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

const emit = defineEmits(['addAction', 'removeAction', 'addcondition', 'removecondition'])

function getExposes() {
    return Object.keys(device.value.exposes)
}


function featureSelectionChanged(event) {

    // var value = event.target.value;
    // if (value == "" || exposeTriggers[value] != null) {
    //     return;
    // }
    // var exposeTrigger = new ExposeTrigger(value);
    // exposeTriggers[value] = exposeTrigger;
    // currentAction.value = exposeTrigger.action
    // console.log("exposeSelectionChanged", exposeTriggers[value]);

    // console.log("featureSelectionChanged changed ", props.trigger.action)
    // currentAction.value = props.trigger.action
}


watch(
    () => props.trigger,
    (newTrigger) => {
        selectedAction.value = null
        currentAction.value = props.trigger.action
        exposeName.value = props.trigger.name

        not working here
        console.log("trigger:", currentAction.value, newTrigger)
        //todo if we have action not null, select it but we dont have friendle name

        if (currentAction.value != null) {
            var device = store.getters["devices/find"](currentAction.value.id)
            console.log("update:", currentAction.value.id, device.friendly_name)

            if (device |= null) {
                selectedAction.value = device.friendly_name
            }
        }
        console.log("trigger:", props.trigger, " : ", exposeName.value)
    }, { immediate: true }
)


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
    currentAction.value = event
    emit('addAction', event)
}

function removeAction(event) {
    currentAction.value = null

    emit('removeAction', event)
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
        <div>
            ------ Accordion HERE --------------
            <h5>Conditions</h5>
            dont delete
            <!-- <div class="row w-50" v-if="exposeName != null">

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
                @change="featureSelectionChanged" :disabled="exposeName == ''">
                <option :value="null">Select device</option>
                <option v-for="device in featureDevices" :value="device.id" :key="device.friendly_name">
                    {{ device.friendly_name }}
                </option>
            </select>
        </div>
        <div>

            <div v-if="currentAction == null">
                <div class="row w-50" v-if="selectedAction != null">
                    <TriggerActionNew :id="selectedAction" @add="addAction"></TriggerActionNew>
                </div>
            </div>
            <div v-else>
                <div class="row w-50">
                    <TriggerActionNew :id="currentAction.id" :property="currentAction.property" :data="currentAction.data"
                        :delay="currentAction.delay" @add="removeAction">
                    </TriggerActionNew>
                </div>
            </div>
        </div>
    </div>
</template>