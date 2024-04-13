<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { EqualityOperators, } from "../../contracts/automations"
import { AutomationTriggerCondition } from "../../types/automation";
import { store } from "../../store/index";
import { allExposeFilter } from '@/configs/automation/device.config';
import ExposeDataInput from '../controls/ExposeDataInput.vue';
import Selection from "../input/Selection.vue"
import ExposeSelector from "@/components/controls/ExposeSelector.vue";


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

function exposeSelected(value: string): void {
    if (value == '') {
        return
    }

    name.value = value;
    data.value = '';
    emit('update:name', value)
}

function operatorUpdated(event: string): void {
    operator.value = event
    emit('update:operator', event)
}

function dataUpdated(event: any): void {
    data.value = event
    emit('update:value', event)
}


const exposeOperators = computed(() => {
    if (name.value == '') {
        return []
    }

    var device = store.getters["devices/find"](props.id);
    if (device.exposes[name.value] == undefined) {
        return []
    }

    const feature = device.exposes[name.value];
    switch (feature.type) {
        case "binary":
        case "enum":
            return Array<string>(EqualityOperators[0]);
        default:
            return EqualityOperators;
    }
});

</script>

<template>
    <div class="row">
        <div class="col-xl-3 col-md-4">
            <ExposeSelector :id="props.id" :value="name" @updated="exposeSelected" :filter="allExposeFilter()"
                :disabled="name != ''" position="center">
            </ExposeSelector>
        </div>
        <div class="col-xl-3 col-md-3">
            <Selection :value="operator" @updated="operatorUpdated" :items="exposeOperators" position="center"
                :disabled="name == ''">
            </Selection>
        </div>
        <div class="col-xl-3 col-md-3">
            <ExposeDataInput :id="props.id" :name="name" :value="data" @updated="dataUpdated" position="center"
                :disabled="name == ''">
            </ExposeDataInput>
        </div>
    </div>
</template>
