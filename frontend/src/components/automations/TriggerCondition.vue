<script setup lang="ts">

import { ref, computed, watch } from 'vue'
import { useStore } from "vuex";
import DataInput from "../input/DataInput.vue"
import { getDeviceExposeNames } from "../../modules/convert"
import { EqualityOperators } from "../../contracts/automations"

const props = defineProps({
    id: {
        type: String,
    },
    name: {
        type: String,
        default: "",
    },
    operator: {
        type: String,
        default: "",
    },
    data: null,
    index: {
        type: Number,
        required: true,
        default: -1,
    }
});

const store = useStore()
const data = ref<any | null>(null)
const operator = ref<string>('')
const name = ref<string>('')

const emit = defineEmits<{
    (e: 'update:name', name: string): void,
    (e: 'update:value', property: any): void,
    (e: 'update:operator', data: string): void,
}>()

const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

const exposes = computed(() => {
    return getDeviceExposeNames(device.value);
})

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

function exposeSelectionChanged(event: string): void {

    if (event == '') {
        return
    }

    data.value = '';
    emit('update:name', event)
}

function operatorUpdated(event: string): void {
    operator.value = event
    emit('update:operator', event)
}

function dataUpdated(event: any): void {
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
            return EqualityOperators[0]
        default:
            return EqualityOperators
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
    <div class="row">
        <div v-if="name == ''" class="col-xl-3 col-md-4">
            <DataInput placeholder="Select trigger" :items="exposes" :data="name" alignment="left" :disabled="name != ''"
                @update:data="exposeSelectionChanged">
            </DataInput>
        </div>
        <div v-else class="col-xl-3 col-md-4">
            <DataInput type="string" placeholder="Name" :data="name" :disabled="true" alignment="left">
            </DataInput>
        </div>
        <div class="col-xl-3 col-md-3">
            <DataInput type="enum" :items="getOperators()" :data="operator" :disabled="name == ''" alignment="center"
                @update:data="operatorUpdated">
            </DataInput>
        </div>
        <div class="col-md-4">
            <DataInput :type="feature.type" :placeholder="getPlaceholder()" :data="data" :items="getItems()"
                alignment="center" :disabled="name == ''" @update:data="dataUpdated">
            </DataInput>
        </div>
    </div>
</template>
