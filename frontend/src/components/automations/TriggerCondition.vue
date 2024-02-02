<script setup lang="ts">

import { ref, computed, watch, PropType, reactive } from 'vue'
import DataInput from "../input/DataInput.vue"
import { EqualityOperators, } from "../../contracts/automations"
import { AutomationTriggerCondition } from "../../types/automation";
import { store } from "../../store/index";
import { Device } from "@/types/device";

const props = defineProps({
    condition: {
        type: Object as PropType<AutomationTriggerCondition>,
        default: {} as AutomationTriggerCondition,
        required: false
    },
    id: {
        type: String,
        default: '',
        required: true
    },
    name: {
        type: String,
        default: "",
    },
    operator: {
        type: String,
        default: "",
    },
    data: null
});

const data = ref<any | null>(null)
const operator = ref<string>('')
const name = ref<string>('')

const emit = defineEmits<{
    (e: 'update:name', name: string): void,
    (e: 'update:value', property: any): void,
    (e: 'update:operator', data: string): void,
    (e: 'update', condition: AutomationTriggerCondition): void,

}>()
const condition = reactive({ ...props.condition }) as AutomationTriggerCondition
const device = computed(() => {
    return store.getters["devices/find"](props.id) as Device;
});

const exposes = computed(() => {
    return Object.keys(device.value.exposes) // CHECK that is correct and we dont eed to return expose.name instead
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

    //condition.value.name = event
    //emit('update', condition)
}

function operatorUpdated(event: string): void {
    operator.value = event
    emit('update:operator', event)

    //condition.value.equality = event
    //emit('update', condition)
}

function dataUpdated(event: any): void {
    data.value = event
    emit('update:value', event)

    //condition.value.value = event
    //emit('update', condition)
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
