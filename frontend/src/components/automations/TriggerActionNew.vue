<script setup>
import { useStore } from "vuex";
import { computed, ref, watchEffect, watch } from "vue";

import { ActionTrigger } from "../../models/automation"

const props = defineProps({
    id: {
        type: String,
        required: true,
    },
    property: String,
    data: String,
    delay: String,
});

const property = ref("");
const data = ref(null);
const delay = ref(null);

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


function delayInputChange(event) {
    delay.value = event.target.value.replace(/[^0-9.]/g, '');
    emit("update:delay", event.target.value);
}

function dataInputChange(event) {
    data.value = event.target.value.replace(/[^0-9.]/g, '');
    emit("update:data", event.target.value);
}

function dataSelectionChanged(event) {
    if (event.target.value == null) {
        return;
    }
    emit("update:data", event.target.value);
}

function propertySelectionChanged(event) {
    var value = event.target.value;

    if (value == null || device.value == undefined) {
        return;
    }

    // TODO:
    // set default value if type is binary - NEEDS REFACTORING
    for (const [key, expose] of Object.entries(device.value.exposes)) {
        if (expose.name != value) {
            continue;
        }
        var keys = Object.keys(expose.properties)
        data.value = expose.properties[keys[0]]
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

function add() {
    var newAction = new ActionTrigger()
    newAction.delay = delay.value
    newAction.data = data.value
    newAction.friendlyName = device.value.friendly_name
    newAction.id = device.value.id
    newAction.property = property.value
    newAction.type = device.value.exposes[property.value].type
    emit("add", newAction)
}

function remove(event) {
    emit("add", props.id)
}

</script>
<template>
    <!-- <div class="col" v-if="device != null">
        <input type="text" style="text-align:center;" class="form-control" placeholder="Friendly name"
            onfocus="this.placeholder = ''" onblur="this.placeholder='Friendly name'" v-model="device.friendly_name"
            disabled />
    </div> -->
    <div class="col-xl-4">
        <select id="featurePropertySelector" style="text-align:center;" class="form-control" v-model="property"
            @change="propertySelectionChanged" :disabled="props.property != null">
            <option value="">Select property</option>
            <option v-for="feature in features" :value="feature.name" :key="feature.name">
                {{ feature.name }}
            </option>
        </select>
    </div>

    <div class="col-xl-4" v-if="feature.type == 'binary'">
        <select id="propertySelect" style="text-align:center;" class="form-control" v-model="data"
            @change="dataSelectionChanged">
            <option v-for="(value, key) in feature.properties" :value="value" :key="key">
                {{ value }}
            </option>
        </select>

    </div>
    <div class="col-xl-4" v-else>
        <input type="text" style="text-align:center;" class="form-control" placeholder="Value"
            onfocus="this.placeholder = ''" onblur="this.placeholder='Value'" v-model="data" @input="dataInputChange"
            :disabled="property == null" />
    </div>
    <div class="col">
        <input type="text" style="text-align:center;" class="form-control" placeholder="Delay"
            onfocus="this.placeholder = ''" onblur="this.placeholder='Delay'" v-model="delay" :disabled="property == null"
            @input="delayInputChange" />
    </div>

    <div class="col-3">
        <div class="btn-group">
            <div v-if="props.property == null">
                <button type="button" class="btn btn-default btn-number" @click="add($event)" :disabled="property == ''">
                    <span class="fa fa-plus"></span>
                </button>
            </div>
            <div v-else>
                <button type="button" class="btn btn-default btn-number" @click="remove($event)">
                    <span class="fa fa-minus"></span>
                </button>
            </div>
        </div>
    </div>
</template>
