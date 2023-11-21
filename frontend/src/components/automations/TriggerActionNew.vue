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
const emit = defineEmits(['add', 'remove'])

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



//TODO
function delayInputChange(event) {
    console.log("delayInputChange", event.target.value)
    delay.value = event.target.value.replace(/[^0-9.]/g, '');
    //emit("newValue", inputValue.value);
}
function dataInputChange(event) {
    console.log("dataInputChange", event.target.value)
    data.value = event.target.value.replace(/[^0-9.]/g, '');
    //emit("newValue", inputValue.value);
}
function dataOptionChange(event) {
    console.log("dataInputChange", event.target.value)
    data.value = event.target.value.replace(/[^0-9.]/g, '');
    //emit("newValue", inputValue.value);
}
//TODO



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
    <div class="col" v-if="device != null">
        <input type="text" style="text-align:center;" class="form-control" placeholder="Friendly name"
            onfocus="this.placeholder = ''" onblur="this.placeholder='Friendly name'" v-model="device.friendly_name"
            disabled />
    </div>
    <div class="col">
        <select id="featurePropertySelector" style="text-align:center;" class="form-control" v-model="property"
            @change="propertySelectionChanged" :disabled="props.property != null">
            <option value="">Select property</option>
            <option v-for="feature in features" :value="feature.name" :key="feature.name">
                {{ feature.name }}
            </option>
        </select>
    </div>

    <div class="col" v-if="feature.type == 'binary'">
        <select id="propertySelect" style="text-align:center;" class="form-control" v-model="data">
            <option v-for="(value, key) in feature.properties" :value="value" :key="key">
                {{ value }}
            </option>
        </select>

    </div>
    <div class="col" v-else>
        <input type="text" style="text-align:center;" class="form-control" placeholder="Value"
            onfocus="this.placeholder = ''" onblur="this.placeholder='Value'" v-model="data" @keyup="dataInputChange"
            :disabled="property == null" />
    </div>
    <div class="col">
        <input type="text" style="text-align:center;" class="form-control" placeholder="Delay (milliseconds)"
            onfocus="this.placeholder = ''" onblur="this.placeholder='Delay (milliseconds)'" v-model="delay"
            :disabled="property == null" @keyup="delayInputChange" />
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
