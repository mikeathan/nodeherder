<script setup lang="ts">

import { ref, computed, watch } from 'vue'
import InputBox from '@/components/input/InputBox.vue';
import DataInput from "../input/DataInput.vue"
import Selection from "../input/Selection.vue"

import { EqualityOperators, } from "../../contracts/automations"
import { AutomationTriggerCondition } from "../../types/automation";
import { store } from "../../store/index";
import { Device } from "@/types/device";
import ExposeSelector from "@/components/controls/ExposeSelector.vue";
import { allExposeFilter } from '@/configs/automation/device.config';
import ExposeDataInput from '../controls/ExposeDataInput.vue';


const props = defineProps({
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

function exposeSelected(event: string): void {

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

const exposeOperators = computed(() => {

    switch (feature.value.type) {
        case "binary":
        case "enum":
            return Array<string>(EqualityOperators[0]);
        default:
            return EqualityOperators;
    }
});


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
        <div class="col-xl-3 col-md-4">
            <ExposeSelector :id="props.id" @updated="exposeSelected" :filter="allExposeFilter()" :disabled="name != ''">
            </ExposeSelector>
        </div>
        <div class="col-xl-3 col-md-3">
            <Selection :value="operator" :disabled="name == ''" @updated="operatorUpdated" :items="exposeOperators">
            </Selection>
        </div>
        <div class="col-md-4">
            {{ feature.id }}
            <ExposeDataInput :id="feature.id" :items="getItems()" label="Set value" @updated="dataUpdated"
                :disabled="name == ''"></ExposeDataInput>
            <!--

            could be selection or input
            <DataInput :type="feature.type" :placeholder="getPlaceholder()" :data="data" :items="getItems()"
                alignment="center" :disabled="name == ''" @update:data="dataUpdated">
            </DataInput> -->
        </div>
    </div>
</template>
