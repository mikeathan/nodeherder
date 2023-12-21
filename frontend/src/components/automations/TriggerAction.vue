<script setup>
import { useStore } from "vuex";
import { computed, ref, watchEffect, watch } from "vue";
import DataInput from "../input/DataInput.vue"
import RadioGroup from "../input/RadioGroup.vue"

import { ActionTrigger, Steps } from "../../models/automation"

const props = defineProps({
    id: {
        type: String,
        required: true,
    },
    property: String,
    data: null,
    delay: null,
    step: null,
    allowRemove: {
        type: Boolean,
        default: true
    },
});
const property = ref("");
const data = ref("");
const delay = ref(null);
const step = ref(null)
const store = useStore();

const emit = defineEmits(['add', 'remove', 'update:data', 'update:delay', , 'update:step'])

watchEffect(() => data.value = props.data);
watchEffect(() => delay.value = props.delay);
watchEffect(() => step.value = props.step);

watch(
    () => props.property,
    () => {
        property.value = props.property
        if (property.value == null) {
            property.value = "" // select first option
        }
    },
    { immediate: true }
);

const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

const features = computed(() => {

    var device = store.getters["devices/find"](props.id);
    if (device == undefined) {
        return []
    }
    var list = []
    for (const [key, expose] of Object.entries(device.exposes)) {
        if (expose.properties != undefined) {
            list.push(expose)
        }
    }

    return list;
});

function propertySelectionChanged(event) {
    var value = event.target.value;
    if (value == "" || device.value == undefined) {
        data.value = "" // reset data

        return;
    }
    delay.value = null;
    step.value = 0;

    // reset data
    for (const [key, feature] of Object.entries(features.value)) {
        if (feature.name == value) {
            if (feature.type == 'binary' || feature.type == "enum") { // TODO: refactor/cleanup
                data.value = ""
            } else {
                data.value = 0
            }
        }
    }
}

const feature = computed(() => {
    if (property.value == null) {
        return []
    }

    var device = store.getters["devices/find"](props.id);
    if (device.exposes[property.value] == undefined) {

        return []
    }

    return device.exposes[property.value];
});

function dataUpdated(event) {
    data.value = event
    emit('update:data', event)
}

function stepUpdated(event) {
    var value = parseInt(event)
    step.value = value
    emit('update:step', value)
}

function delayUpdated(event) {
    delay.value = event
    emit('update:delay', event)
}

function add() {
    var newAction = new ActionTrigger()
    newAction.delay = delay.value
    newAction.step = step.value
    newAction.data = data.value
    newAction.friendlyname = device.value.friendly_name
    newAction.id = device.value.id
    newAction.property = property.value
    newAction.type = device.value.exposes[property.value].type
    emit("add", newAction)
}

function remove(event) {
    emit("add", props.id)
}

function getPlaceholder(type) {
    if (type == 'binary' || type == 'enum') {
        return 'Select'
    }

    return 'Value'
}

function getItems() {

    switch (feature.value.type) {
        case "binary":
        case "enum":
            return Object.values(feature.value.properties)
        default:
            return null
    }
}

</script>

<style scoped>
.plain-select {
    border: 1;
    outline: 1;
    appearance: none;
    border: none;
    background: none;
    background-color: transparent;
    font-family: inherit;
    outline: none;
}

.plain-select:focus,
:active,
:hover {
    outline: none;
    box-shadow: none;
}

select.form-control {
    background-color: transparent;
    background-clip: padding-box;
}

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
    <div class="col-1">
        <div v-if="props.property == null">
            <button type="button" class="btn btn-default btn-number" @click="add($event)" :disabled="data == ''">
                <span class="fa fa-plus"></span>
            </button>
        </div>
        <div v-else-if="props.allowRemove">
            <button type="button" class="btn btn-default btn-number" @click="remove($event)">
                <span class="fa fa-minus"></span>
                <!-- <span class="fa fa-trash"></span> -->

            </button>
        </div>
    </div>

    <div class="row">
        <div class="col-4">
            <label>Property:</label>
        </div>
        <div class="col-6">
            <select id="featurePropertySelector" style="text-align:left;" class="form-control plain-select"
                v-model="property" @change="propertySelectionChanged" :disabled="props.property != null">
                <option value="">Select</option>
                <option v-for="feature in features" :value="feature.name" :key="feature.name">
                    {{ feature.name }}
                </option>
            </select>
        </div>
    </div>

    <div class="row">
        <div class="col-4">
            <label>Value:</label>
        </div>
        <div class="col-6">
            <DataInput :type="feature.type" :placeholder="getPlaceholder(feature.type)" :items="getItems()" :data="data"
                :disabled="property == ''" @update:data="dataUpdated" alignment="left">
            </DataInput>
        </div>
    </div>
    <div class="row">
        <div class="col-4">
            <label>Delay:</label>
        </div>
        <div class="col-6">
            <DataInput placeholder="Delay" type="numeric" :data="delay" :disabled="property == ''"
                @update:data="delayUpdated" alignment="left">
            </DataInput>
        </div>
    </div>




    <div v-if="feature.type == 'numeric'">
        <label class="pt-3"> Steps</label>
        <RadioGroup :items="Steps" :value="step" @update:data="stepUpdated"></RadioGroup>
    </div>
</template>
