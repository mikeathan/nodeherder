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

function propertySelectionChanged(event) {
    var value = event.target.value;

    if (value == null || device.value == undefined) {
        return;
    }

    // TODO: we cant do thta becasue it resets the value
    // set default value if type is binary - NEEDS REFACTORING
    // for (const [key, expose] of Object.entries(device.value.exposes)) {
    //     if (expose.name != value) {
    //         continue;
    //     }
    //     var keys = Object.keys(expose.properties)
    //     data.value = expose.properties[keys[0]]
    // }
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
    <!-- <div class="col" v-if="device != null">
        <input type="text" style="text-align:center;" class="form-control" placeholder="Friendly name"
            onfocus="this.placeholder = ''" onblur="this.placeholder='Friendly name'" v-model="device.friendly_name"
            disabled />
    </div> -->
    <div class="col-xl-4">

        <select id="featurePropertySelector" style="text-align:center;" class="form-control inputName" v-model="property"
            @change="propertySelectionChanged" :disabled="props.property != null">
            <option value="">Select property</option>
            <option v-for="feature in features" :value="feature.name" :key="feature.name">
                {{ feature.name }}
            </option>
        </select>
    </div>
    <!-- 
    TODO use conditional logic between data and feature.properties fix order of rendering
    fix blur name -->
    <div class="col-xl-3" v-if="feature.type == 'binary'">
        <DataInput :type="feature.type" name="Value" :data="feature.type == 'binary' ? feature.properties : data"
            @update:data="newValue => data = newValue">
        </DataInput>
    </div>

    <div class="col-xl-3" v-else>
        <DataInput :type="feature.type" name="Value" :data="data" :disabled="property == ''"
            @update:data="newValue => data = newValue"></DataInput>
    </div>
    <div class="col-xl-3">
        <DataInput name="Delay" :data="delay" :disabled="property == ''" @update:data="newValue => delay = newValue">
        </DataInput>
    </div>

    <div class="col-xl-2">
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
