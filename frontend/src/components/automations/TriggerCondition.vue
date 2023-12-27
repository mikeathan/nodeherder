<script setup>
import { OperatorKeys } from "../../models/automation"
import DataInput from "../input/DataInput.vue"
import { useStore } from "vuex";
import Selector from "../input/Selector.vue"

import { ref, computed, watch } from 'vue'
const props = defineProps({
    id: {
        type: String,
        required: true,
    },
    name: String,
    operator: String,
    data: null,
    index: {
        type: Number,
        required: true,
        default: -1,
    }
});

const store = useStore()
const data = ref(null)
const operator = ref('')
const name = ref('')

const emit = defineEmits(['add', 'remove', 'update:name', 'update:value', 'update:operator'])


function remove() {
    emit("remove", props.index)
}

function getExposes() {

    var list = {}
    for (const [key, expose] of Object.entries(device.value.exposes)) {
        list[expose.name] = expose.name
    }
    return list
}

const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

const feature = computed(() => {
    if (name.value == '') {
        return []
    }

    var device = store.getters["devices/find"](props.id);
    if (device.exposes[name.value] == undefined) {

        return []
    }

    return device.exposes[name.value];
});

watch(
    () => props.name,
    () => {
        name.value = props.name
    }, { immediate: true }
)

watch(
    () => props.data,
    () => {
        data.value = props.data
    }, { immediate: true }
)

watch(
    () => props.operator,
    () => {
        operator.value = props.operator
    }, { immediate: true }
)

function reset() {
    data.value = ""
    operator.value = OperatorKeys[0]
}

function exposeSelectionChanged(event) {

    if (event == '') {
        return
    }

    data.value = "";
    emit('update:name', event)
}

function operatorUpdated(event) {
    operator.value = event
    emit('update:operator', event)
}

function dataUpdated(event) {
    data.value = event
    emit('update:value', event)
}

function getPlaceholder() {

    switch (feature.value.type) {
        case "binary":
        case "enum":
            return 'Select'
        default:
            return 'Value'
    }
}

function getOperators() {

    switch (feature.value.type) {
        case "binary":
        case "enum":
            return OperatorKeys[0]
        default:
            return OperatorKeys
    }

}
function getItems() {

    if (feature.value.attributes == undefined) {
        return null
    }
    switch (feature.value.type) {
        case "binary":
        case "enum":

            return Object.values(feature.value.attributes)
        default:
            return null
    }
}

</script>

<template>
    <div v-if="props.name == ''" class="col-xl-3">
        <Selector placeholder="Select trigger" :items="getExposes()" :value="name" @update:data="exposeSelectionChanged">
        </Selector>
    </div>
    <div v-else class="col-xl-3">
        <DataInput type="string" placeholder="Name" :data="name" :disabled="true" alignment="center">
        </DataInput>
    </div>
    <div class="col-xl-2">
        <DataInput type="enum" :items="getOperators()" :data="operator" :disabled="name == ''" alignment="center"
            @update:data="operatorUpdated">
        </DataInput>
    </div>
    <div class="col">
        <DataInput :type="feature.type" :placeholder="getPlaceholder()" :data="data" :items="getItems()" alignment="center"
            :disabled="name == ''" @update:data="dataUpdated">
        </DataInput>
    </div>
    <div class="col-xl-2">
        <div class="btn-group">
            <button type="button" class="btn btn-default btn-number" @click="remove($event)">
                <span class="fa fa-minus"></span>
            </button>
        </div>
    </div>
</template>
