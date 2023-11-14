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
const conditions = ref([]);
const selectedExpose = ref("")
const exposeTriggers = {};

const description = ref("")
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
        selectedExpose.value = null
    },
    { immediate: true }
);

function getExposes() {
    return Object.keys(device.value.exposes)
}

function create() {

    var deviceTrigger = new DeviceTrigger()
    deviceTrigger.friendly_name = device.value.friendly_name;
    deviceTrigger.id = device.value.id
    deviceTrigger.description = description.value

    for (const [key, item] of Object.entries(exposeTriggers)) {
        deviceTrigger.triggers.push(item)
    }
    console.log("new device Trigger:", deviceTrigger)
    cancel();
}

function cancel() {

    conditions.value = []
    emit("cancel")
}

function add(event) {

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
            <input type="text" class="form-control" name="name" id="name" v-model="description" placeholder="Name"
                onfocus="this.placeholder = ''" onblur="this.placeholder='Name'">
        </div>

        <div class="col-3">
            <select id="exposeSelector" style="text-align:center;" class="form-control" @change="exposeSelectionChanged"
                v-model="selectedExpose">
                <option :value="null">Select expose</option>
                <option v-for="expose in device.exposes" :value="expose.name" :key="expose.name">
                    {{ expose.name }}
                </option>
            </select>
        </div>
        <br>

        <div>
            <div class="row w-50" v-if="selectedExpose != null">
                <TriggerCondition :id="0" :exposes="getExposes()" :name="selectedExpose" :operator="Operators[0].value"
                    :data="''" @add="add($event)">
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
                    <button type="button" class="btn btn-default"
                        :disabled="exposeTriggers.length == 0 && description.length == 0" @click="create">
                        Save
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>
