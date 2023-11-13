<script setup>
import { useStore } from "vuex";
import { computed, watch, ref, onBeforeMount, watchEffect, reactive } from "vue";
import { useRouter } from "vue-router";

import { ExposeTrigger, DeviceTrigger, Operators } from "../../models/automation"
import Conditions from "./Conditions"
import TriggerCondition from "./TriggerCondition"

const props = defineProps({
    id: String,
    trigger: Object
});

const store = useStore();
const deviceTrigger = ref(new DeviceTrigger())
const conditions = ref([]);
const selectedExpose = ref("")
const exposeTriggers = ref({});
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

const previousPage = computed(() => {
    var back = useRouter().options.history.state.back;
    if (back == undefined) {
        cancel();
        back = useRouter().push("/");
    }
    return back;
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
        deviceTrigger.value = new DeviceTrigger()
        deviceTrigger.value.id = newId
        selectedExpose.value = null
    },
    { immediate: true }
);

function addTrigger() {
    deviceTrigger.value.friendly_name = device.value.friendly_name;
    deviceTrigger.value.triggers.push(exposeTriggers[selectedExpose.value])

    // reset state
    conditions.value = []
    exposeTriggers[selectedExpose.value] = null
    selectedExpose.value = null
}

function getExposes() {
    return Object.keys(device.value.exposes)
}

function create() {

    var t = new DeviceTrigger()
    t.friendly_name = device.value.friendly_name;
    t.description = "TODO";// TODO;
    // for (var key in Object.keys(exposeTriggers)) {
    //     var exposeTrigger = exposeTriggers[key]
    //     console.log("adding ", key, ":", exposeTrigger)
    //     t.triggers.push(exposeTrigger)
    // }
    // console.log("new device Trigger:", t)

    // emit("create", deviceTrigger.value)

    for (const [key, item] of Object.entries(exposeTriggers)) {
        console.log("key; ", key, "item ", item)
        t.triggers.push(item)
    }
    console.log("new device Trigger:", t)
    cancel();
}

function cancel() {

    conditions.value = []
    emit("cancel")
}

function add(event) {
    // TODO:
    // to minizze the buttons eg get rid of add trigger button
    // we can add the condition to conditions and add it to a dictionary with key of current expose
    // only on final create we can construct the object

    // TODO: !!!!!!11
    // need to be able to exit from locked exposes options 
    conditions.value.push(event);

    var exposeTrigger = exposeTriggers[selectedExpose.value];
    exposeTrigger.conditions.push(event)
}

function remove(event) {
    var index = conditions.value.findIndex(item => item.id === event);
    if (index != -1) {
        conditions.value.splice(index, 1);
    }
}

</script>
<template>
    <div class="container-fluid p-0 h-100" v-if="device != null"> <!-- to fix condition-->

        <div class="align-self-center me-3">
            <RouterLink :to="`${previousPage}`">
                <i class="fa fa-arrow-left fa-xl" aria-hidden="true"></i>
            </RouterLink>
        </div>
        <div class="col-3">
            <input type="text" class="form-control" name="name" id="name" v-model="deviceTrigger.description"
                placeholder="Name" onfocus="this.placeholder = ''" onblur="this.placeholder='Name'">
        </div>

        <div class="col-3">
            <select id="exposeSelector" style="text-align:center;" class="form-control" @change="exposeSelectionChanged"
                v-model="selectedExpose" :disabled="selectedExpose != null">
                <option :value="null">Select expose</option>
                <option v-for="expose in device.exposes" :value="expose.name" :key="expose.name">
                    {{ expose.name }}
                </option>
            </select>
        </div>
        <br>

        <div>
            <div class="row w-50" v-if="selectedExpose != null">
                <TriggerCondition :id="0" :exposes="getExposes()" :operator="Operators[0].value" :data="''"
                    @add="add($event)">
                </TriggerCondition>
            </div>

            <div v-for="condition in conditions">
                <div class="row w-50">
                    <TriggerCondition :id="condition.id" :name="condition.name" :operator="condition.operator"
                        :key="condition.id" :data="condition.data" @remove="remove($event)"></TriggerCondition>
                </div>
            </div>
            <div class="col-50">
                <div class="btn-group">

                    <!-- <button type="button" class="btn btn-default" :disabled="conditions.length == 0" @click="addTrigger">
                        Add
                    </button> -->

                    <button type="button" class="btn btn-default"
                        :disabled="deviceTrigger.triggers.length == 0 && deviceTrigger.description.length == 0"
                        @click="create">
                        Done
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>
