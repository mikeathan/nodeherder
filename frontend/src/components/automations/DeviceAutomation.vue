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
const currentTrigger = ref(null)
const exposeTriggers = ref({});
const enabled = ref(false)
const description = ref("")
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});


function exposeSelectionChanged(value) {
    selectedExpose.value = value;

}


watch(
    () => props.id,
    () => {
        selectedExpose.value = ""
        var exposeTrigger = new ExposeTrigger(selectedExpose.value);
        currentTrigger.value = exposeTrigger;
    },
    { immediate: true }
);

const emit = defineEmits(['cancel', 'create'])
function isAddTriggerEnabled() {
    var value = selectedExpose.value;
    if (exposeTriggers.value[value] != null) {
        return false
    }
    return value != '';
}

function isSaveEnabled() {
    // find entries with actions only
    var values = Object.values(exposeTriggers.value).filter(k => k.action != null);
    return values.length > 0 && description.value.length > 0
}

function addTrigger() {
    var value = selectedExpose.value;
    var exposeTrigger = new ExposeTrigger(value);
    exposeTriggers.value[value] = exposeTrigger;
    currentTrigger.value = exposeTriggers.value[value]
    selectedExpose.value = ''; //de-select combo

}
function onDeleteTriggerClick(event, triggerName) {
    // disable accordion from expanding
    event.stopImmediatePropagation();
    event.preventDefault();
    delete exposeTriggers.value[triggerName]
    //automation.value.triggers.splice(triggerId, 1);
}

function create() {
    var deviceTrigger = new DeviceTrigger()
    deviceTrigger.friendlyname = device.value.friendly_name;
    deviceTrigger.id = device.value.id
    deviceTrigger.enabled = enabled.value
    deviceTrigger.description = description.value

    for (const [key, item] of Object.entries(exposeTriggers.value)) {
        deviceTrigger.triggers.push(item)
    }

    reset();

    console.log("create", deviceTrigger)

    emit("create", deviceTrigger)
}


function reset() {

    conditions.value = []
    exposeTriggers.value = {}
    description.value = ""
    enabled.value = false
    emit("cancel")
}

function addAction(event, triggerName) {
    var exposeTrigger = exposeTriggers.value[triggerName]
    exposeTrigger.action = event;
}

function removeAction(event, triggerName) {
    var exposeTrigger = exposeTriggers.value[triggerName]
    exposeTrigger.action = null;
}

function addCondition(event, triggerName) {
    conditions.value.push(event);

    var exposeTrigger = exposeTriggers.value[triggerName]
    exposeTrigger.conditions.push(event)
}

function removeCondition(event, triggerName) {
    var index = conditions.value.findIndex(item => item.idx === event);

    if (index != -1) {

        conditions.value.splice(index, 1);

        // remove from our device trigger cache
        for (const [key, item] of Object.entries(triggerName)) {
            var index = item.conditions.findIndex(item => item.idx === event);
            if (index == -1) {
                continue
            }

            item.conditions.splice(index, 1);
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
                <label class="form-check-label ms-3">Enable</label>

                <input class="form-check-input custom-control-input" type="checkbox" role="switch"
                    id="flexSwitchCheckDefault" v-model="enabled">
            </div>
        </div>
        <div class="col-50 mt-3 mb-4">
            <div class="btn-group">
                <button type="button" class="btn btn-light" :disabled="isSaveEnabled() == false" @click="create">
                    Save
                </button>
                <button type="button" class="btn btn-light" @click="reset">
                    Cancel
                </button>
            </div>
        </div>

        <div class="mb-3">
            <Selector placeholder="Select trigger" :items="exposesList()" :value="selectedExpose" alignment="left"
                @update:data="exposeSelectionChanged"></Selector>
        </div>
        <div class="col-50 mt-3 mb-4">
            <div class="btn-group">
                <button type="button" class="btn btn-light" :disabled="isAddTriggerEnabled() == false" @click="addTrigger">
                    Add new trigger
                </button>

            </div>
        </div>

        <div class="accordion accordion-flush" id="triggersList">
            <div v-for="(trigger, key) in  exposeTriggers ">

                <div class="accordion-item">

                    <h2 class="accordion-header" :id="`header${trigger.name}`">

                        <div class="col accordion-button collapsed " data-bs-toggle="collapse"
                            :data-bs-target="`#collapse${trigger.name}`" aria-expanded="false"
                            :aria-controls="`collapse${trigger.name}`">

                            <div class="col">
                                Trigger # {{ trigger.name }}
                            </div>

                            <div class="col pe-3 text-end ">
                                <span class="fa fa-trash-alt fa-lg" @click="onDeleteTriggerClick($event, trigger.name)"
                                    data-bs-toggle="collapse" data-bs-target>
                                </span>

                            </div>
                        </div>
                    </h2>

                    <div :id="`collapse${trigger.name}`" class="accordion-collapse collapse"
                        :aria-labelledby="`header${trigger.name}`" data-bs-parent="#triggersList">
                        <div class="accordion-body">
                            <Trigger :id="props.id" :trigger="trigger" @addAction="addAction($event, trigger.name)"
                                @removeAction="removeAction($event, trigger.name)"
                                @addCondition="addCondition($event, trigger.name)"
                                @removeCondition="removeCondition($event, trigger.name)">
                            </Trigger>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>
