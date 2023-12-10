<script setup>
import { useStore } from "vuex";
import { computed, ref, watchEffect, watch } from "vue";
import DataInput from "../input/DataInput.vue"

import { ActionTrigger } from "../../models/automation"

const props = defineProps({
    id: {
        type: String,
        required: true,
    },
    property: String,
    data: null,
    delay: null,
    allowRemove: {
        type: Boolean,
        default: true
    },
});
const steps = ["increase", "decrease"]
const property = ref("");
const data = ref("");
const delay = ref(null);
const step = ref(null)
const store = useStore();
const emit = defineEmits(['add', 'remove', 'update:data', 'update:delay'])

watchEffect(() => data.value = props.data);
watchEffect(() => delay.value = props.delay);
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
    // reset data
    for (const [key, feature] of Object.entries(features.value)) {
        if (feature.name == value) {
            if (feature.type == 'binary') {
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

function delayUpdated(event) {
    delay.value = event
    emit('update:delay', event)
}

function add() {
    var newAction = new ActionTrigger()
    newAction.delay = delay.value
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

function getPlaceholder() {
    if (feature.value.type == 'binary' || feature.value.type == 'enum') {
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
    <div class="col">
        <select id="featurePropertySelector" style="text-align:center;" class="form-control inputName" v-model="property"
            @change="propertySelectionChanged" :disabled="props.property != null">
            <option value="">Select property</option>
            <option v-for="feature in features" :value="feature.name" :key="feature.name">
                {{ feature.name }}
            </option>
        </select>
    </div>

    <div class="col">
        <DataInput :type="feature.type" :placeholder="getPlaceholder()" :items="getItems()" :data="data"
            :disabled="property == ''" @update:data="dataUpdated">
        </DataInput>
    </div>
    <div class="col">
        <DataInput placeholder="Delay" type="numeric" :data="delay" :disabled="property == ''" @update:data="delayUpdated">
        </DataInput>
    </div>

    <div class="col">
        <div class="btn-group">
            <div v-if="props.property == null">
                <button type="button" class="btn btn-default btn-number" @click="add($event)" :disabled="data == ''">
                    <span class="fa fa-plus"></span>
                </button>
            </div>
            <div v-else-if="props.allowRemove">
                <button type="button" class="btn btn-default btn-number" @click="remove($event)">
                    <span class="fa fa-minus"></span>
                </button>
            </div>
        </div>
    </div>

    <div v-if="feature.type == 'numeric'" class="row">
        create drop down for selection of increase/decrease - optional
        <DataInput :type="feature.type" placeholder="Step" :items="steps" :data="step" :disabled="property == ''">
            @update:data="dataUpdated"
        </DataInput>
    </div>
</template>
