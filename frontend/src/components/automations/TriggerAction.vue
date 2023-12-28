<script setup>
import { useStore } from "vuex";
import { computed, ref, watchEffect, watch } from "vue";
import DataInput from "../input/DataInput.vue"
import Selector from "../input/Selector.vue"

import { Steps } from "../../models/automation"

const props = defineProps({
    id: {
        type: String,
        default: ''
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
const id = ref("")
const property = ref("");
const data = ref("");
const delay = ref(null);
const step = ref('')
const store = useStore();

const emit = defineEmits(['remove', 'update:property', 'update:id', 'update:data', 'update:delay', , 'update:step'])
watchEffect(() => id.value = props.id);
watchEffect(() => data.value = props.data);
watchEffect(() => delay.value = props.delay);
watchEffect(() => step.value = props.step);


watch(
    () => props.step,
    () => {
        step.value = props.step
        if (step.value == null) {
            step.value = Steps[0].value
        }
    },
    { immediate: true }
);
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
    return store.getters["devices/find"](id.value);
});

const features = computed(() => {

    var device = store.getters["devices/find"](id.value);
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
        property.value = ""
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

    emit('update:property', property.value)
}

const feature = computed(() => {
    if (property.value == null) {
        return []
    }

    var device = store.getters["devices/find"](id.value);
    if (device == undefined) {
        return []
    }
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

function idUpdated(event) {
    id.value = event
    var device = store.getters["devices/find"](id.value);
    if (device != undefined) {
        emit('update:id', id.value, device.friendly_name)
    }
}

function remove() {
    emit("remove", id.value)
}

const featureDevices = computed(() => {
    var devices = store.getters["devices/items"];

    // find exposes with properties
    // let all = items.filter(item=> item.age==='18')
    //     return deviceimport 
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

function deviceList() {
    // todo:
    //var result = Object.keys(obj).map((key) => [key, obj[key]]);
    var list = {}
    for (const [key, device] of Object.entries(featureDevices.value)) {
        list[device.friendly_name] = device.id
    }
    return list
}

function getPlaceholder(type) {
    if (type == 'binary' || type == 'enum') {
        return 'Select'
    }

    return 'Value'
}

const getFeatureNames = computed(() => {
    // todo:
    var list = {}
    for (const [key, feature] of Object.entries(features.value)) {
        list[feature.name] = feature.name
    }
    return list
})

function getSteps() {
    // todo:
    var list = {}
    for (const [key, step] of Object.entries(Steps)) {
        list[step.name] = step.value
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

<template>
    <div class="col-xl-3">
        <Selector placeholder=" Select device" :items="deviceList()" :value="id" alignment="left" @update:data="idUpdated"
            :disabled="id != ''">
        </Selector>
    </div>
    <div class="col-xl-3">
        <Selector placeholder="Select property" :items="getFeatureNames" :value="property" alignment="center"
            @update:data="propertySelectionChanged" :disabled="id == ''">
        </Selector>
    </div>
    <div class="col-xl-4">
        <DataInput :type="feature.type" :placeholder="getPlaceholder(feature.type)" :items="getItems()" :data="data"
            :disabled="property == ''" @update:data="dataUpdated">
        </DataInput>
    </div>
    <div class="col-xl-2">
        <div class="btn-group">
            <button v-if="feature.type == 'numeric'" class="btn btn-default btn-number" type="button"
                data-bs-toggle="collapse" data-bs-target="#collapseOptions" aria-expanded="false"
                aria-controls="collapseOptions">
                <span class="fas fa-angle-double-down"></span>
            </button>

            <button type="button" class="btn btn-default btn-number" @click="remove">
                <span class="fa fa-minus"></span>
            </button>
        </div>

    </div>
    <div class="collapse " id="collapseOptions">
        <div class="row pt-2">
            <div class="col-xl-2 col-md-2">
                <DataInput placeholder="Delay" type="numeric" :data="delay" :disabled="property == ''"
                    @update:data="delayUpdated">
                </DataInput>
            </div>
            <div class="col-xl-3 col-md-2" v-if="feature.type == 'numeric'">
                <Selector :items="getSteps()" :value="step" :disabled="property == ''" @update:data="stepUpdated">
                </Selector>
            </div>
        </div>
    </div>
</template>
