<script setup>
import { useStore } from "vuex";
import { computed, watch, ref } from "vue";
import { ExposeTrigger, DeviceTrigger } from "../../models/automation"
import Trigger from "./Trigger"
import DataInput from "../input/DataInput.vue"
import Selector from "../input/Selector.vue"

const props = defineProps({
    id: String,
});

const store = useStore();
const conditions = ref([]);
const selectedExpose = ref("")
const triggers = ref([])
const showExposeSelection = ref(false)
const enabled = ref(false)
const description = ref("")
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});


watch(
    () => props.id,
    () => {
        selectedExpose.value = ""
    },
    { immediate: true }
);

const emit = defineEmits(['cancel', 'create'])

function isSaveEnabled() {
    var values = triggers.value.filter(k => k.action != null);
    return values.length == triggers.value.length && description.value.length > 0
}

function onDeleteTriggerClick(event, index) {
    // disable accordion from expanding
    event.stopImmediatePropagation();
    event.preventDefault();
    triggers.value.splice(index, 1);
}

function showExposesSelection() {
    // show exposes drop down
    selectedExpose.value = "" // reset 
    showExposeSelection.value = true
}

function create() {
    var deviceTrigger = new DeviceTrigger()
    deviceTrigger.friendlyname = device.value.friendly_name;
    deviceTrigger.id = device.value.id
    deviceTrigger.enabled = enabled.value
    deviceTrigger.description = description.value

    for (const trigger of triggers.value) {
        deviceTrigger.triggers.push(trigger)
    }

    reset();
    emit("create", deviceTrigger)
}

function reset() {

    conditions.value = []
    triggers.value = []
    description.value = ""
    enabled.value = false
    showExposeSelection.value = false
    selectedExpose.value = ""
    emit("cancel")
}

function addTrigger() {
    if (selectedExpose.value == "") {
        return
    }

    var trigger = new ExposeTrigger(selectedExpose.value)
    triggers.value.push(trigger)
    showExposeSelection.value = false // hide selection   
}

function addAction(event, index) {
    var trigger = triggers.value[index]
    trigger.action = event
}

function removeAction(event, index) {
    var trigger = triggers.value[index]
    trigger.action = null
}

function addCondition(event, index) {
    conditions.value.push(event);
    var trigger = triggers.value[index]
    trigger.conditions.push(event)
}

function removeCondition(event, index) {
    var cIdx = conditions.value.findIndex(item => item.idx === event);
    if (cIdx != -1) {

        conditions.value.splice(cIdx, 1);

        // remove from triggers cache
        var trigger = triggers.value[index]
        var idx = trigger.conditions.findIndex(i => i.idx === event);
        if (idx != -1) {
            trigger.conditions.splice(idx, 1);
        }
    }
}
function exposesList() {
    // todo:
    //var result = Object.keys(obj).map((key) => [key, obj[key]]);
    var list = {}
    for (const [key, expose] of Object.entries(device.value.exposes)) {
        list[expose.name] = expose.name
    }
    return list
}

// design
//https://www.home-assistant.io/getting-started/automation/
</script>

<style scoped>
.custom-control-input {
    transform: scale(1.4);
}
</style>
<template>
    <div v-if="device != null"> <!-- to fix condition-->

        <div class="pt-3 pb-4">
            <label class="form-check-label" for="name">Name</label>
            <DataInput placeholder="New automation" :data="description" @update:data="(value) => description = value"
                type="string" alignment="left">
            </DataInput>
        </div>

        <div class="pb-3">
            <div class=" form-check form-switch ms-2">
                <label class="form-check-label">Enable</label>

                <input class="form-check-input custom-control-input" type="checkbox" role="switch"
                    id="flexSwitchCheckDefault" v-model="enabled">
            </div>
        </div>

        <div class="col-50 mt-3 mb-4">
            <div class="btn-group">
                <button type="button" class="btn btn-light" @click="showExposesSelection">
                    Add
                </button>
                <button type="button" class="btn btn-light" :disabled="isSaveEnabled() == false" @click="create">
                    Save
                </button>
                <button type="button" class="btn btn-light" @click="reset">
                    Cancel
                </button>
            </div>
        </div>

        <div class="mb-3" v-if="showExposeSelection">
            <div class="d-flex">
                <Selector placeholder="Select trigger" :items="exposesList()" :value="selectedExpose" alignment="left"
                    @update:data="val => selectedExpose = val" @change="addTrigger"></Selector>
            </div>
        </div>

        <div class="accordion accordion-flush" id="triggersList">
            <div v-for="(trigger, index) in  triggers ">

                <div class="accordion-item">
                    <h2 class="accordion-header" :id="`header${index}`">
                        <div class="col accordion-button collapsed " data-bs-toggle="collapse"
                            :data-bs-target="`#collapse${index}`" aria-expanded="false" :aria-controls="`collapse${index}`">

                            <div class="col">
                                Trigger #{{ index + 1 }} - {{ trigger.name }}
                            </div>

                            <div class="col pe-3 text-end ">
                                <span class="fa fa-trash-alt fa-lg" @click="onDeleteTriggerClick($event, index)"
                                    data-bs-toggle="collapse" data-bs-target>
                                </span>

                            </div>
                        </div>
                    </h2>

                    <div :id="`collapse${index}`" class="accordion-collapse collapse show"
                        :aria-labelledby="`header${index}`" data-bs-parent="#triggersList">
                        <div class="accordion-body">
                            <Trigger :id="props.id" :trigger="trigger" @addAction="addAction($event, index)"
                                @removeAction="removeAction($event, index)" @addCondition="addCondition($event, index)"
                                @removeCondition="removeCondition($event, index)">
                            </Trigger>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
