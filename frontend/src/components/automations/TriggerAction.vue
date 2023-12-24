<script setup>
import { useStore } from "vuex";
import { computed, ref, watchEffect, watch } from "vue";
import DataInput from "../input/DataInput.vue"
import RadioGroup from "../input/RadioGroup.vue"
import Selector from "../input/Selector.vue"

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
    var value = event;
    if (value == "" || device.value == undefined) {
        data.value = "" // reset data

        return;
    }
    property.value = value
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

function getFeatureNames() {
    // todo:
    var list = {}
    for (const [key, feature] of Object.entries(features.value)) {
        list[feature.name] = feature.name
    }
    return list
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

select.form-control:focus,
:active,
:hover {
    outline: none;
    box-shadow: none;
}

/* select.form-control {
    background-color: transparent;
    background-clip: padding-box;
    border: 1;
    outline: 1;
    appearance: none;
} */
</style>
<template>
    <div class="col-xl-3 col-md-2">

        <Selector placeholder="Select property" :items="getFeatureNames()" :value="property" alignment="center"
            @update:data="propertySelectionChanged" :disabled="props.property != null">
        </Selector>

    </div>
    <div class="col-xl-2 col-md-2">
        <DataInput :type="feature.type" :placeholder="getPlaceholder(feature.type)" :items="getItems()" :data="data"
            :disabled="property == ''" @update:data="dataUpdated">
        </DataInput>
    </div>
    <div class="col-xl-2 col-md-2">
        <DataInput placeholder="Delay" type="numeric" :data="delay" :disabled="property == ''" @update:data="delayUpdated">
        </DataInput>
    </div>

    <div class="row" v-if="feature.type == 'numeric'">
        <label>Steps</label>
        <div class="col-xl-6 col-md-6">
            <RadioGroup :items="Steps" :value="step" @update:data="stepUpdated"></RadioGroup>
        </div>
    </div>

    <!-- buttons -->
    <div class="row pt-2">
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
</template>
