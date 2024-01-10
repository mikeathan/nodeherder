<script setup lang="ts">
import { useStore } from "vuex";
import { computed, ref, watchEffect, watch } from "vue";
import DataInput from "../input/DataInput.vue"
import Selector from "../input/Selector.vue"
import RadioGroup from "../input/RadioGroup.vue"
import Toggle from "../input/Toggle.vue"


import { getFeatureExposes, getFeatureDevices, createMapFromObject } from "../../modules/convert"
import { ExposeTrigger, OperationType, Operation, OperationContent, OperationTypeKeys } from "../../models/automations"

const props = defineProps({
    id: {
        type: String,
        default: ''
    },
    property: {
        type: String,
        default: ''
    },
    data: null,
    delay: null,
    operation: {
        type: Number,
        default: 0
    }
});

const id = ref<string>("")
const property = ref<string>("");
const data = ref<any>("");
const delay = ref<number | null>(null);
const operation = ref<number>(0)
const showStepsSelection = ref<boolean>(false);
const availableOperations = ref<Array<{ [key: string]: number; }>>()

const store = useStore();

const emit = defineEmits<{
    (e: 'update:id', id: string, name: string): void,
    (e: 'update:property', property: string): void,
    (e: 'update:data', data: any): void,
    (e: 'update:delay', data: number): void,
    (e: 'update:operation', data: number): void,
}>()

watchEffect(() => id.value = props.id);
watchEffect(() => data.value = props.data);
watchEffect(() => operation.value = props.operation);


watch(
    () => props.delay,
    () => {
        delay.value = props.delay
        if (delay.value != null && delay.value >= 1000) {
            delay.value /= 60000 // convert to minutes
        }
    },
    { immediate: true }
);
watch(
    () => props.property,
    () => {
        property.value = props.property == null ? "" : props.property
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

    return getFeatureExposes(device)
});

function propertyUpdated(event: any) {
    var value = event;
    if (value == "" || device.value == undefined) {
        property.value = ""
        data.value = "" // reset data
        return;
    }

    property.value = value
    delay.value = null;
    operation.value = 0;
    availableOperations.value = tempOperationsBuilder()
    if (feature.value.type == 'binary' || feature.value.type == "enum") {
        data.value = ""
    } else {
        data.value = 0
    }

    // // reset data
    // for (const [key, feature] of Object.entries(features.value)) {
    //     if (feature.name == value) {
    //         if (feature.type == 'binary' || feature.type == "enum") { // TODO: refactor/cleanup
    //             data.value = ""
    //         } else {
    //             data.value = 0
    //         }
    //     }
    // }

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
// NOTE
// its messy but we need it for now as device is not a defined class
// keep it for now until refactoring 
function tempOperationsBuilder(): any {

    const availableOperations = Array<Operation>(OperationContent[OperationType.NoOperation])
    if (feature.value.type === 'numeric') {
        availableOperations.push(OperationContent[OperationType.StepOperation])
    }
    if (feature.value.presets !== undefined) {
        availableOperations.push(OperationContent[OperationType.RotationOperation])
    }

    let dictionary = Object.assign({}, ...availableOperations.map(({
        name,
        value
    }) => ({
        [name]: value
    })));

    return dictionary
}

function operationUpdated(operation: number) {
    switch (+operation) {
        case OperationType.StepOperation:
            showStepsSelection.value = true;
            return
        case OperationType.NoOperation:
        case OperationType.RotationOperation:
            showStepsSelection.value = false;
            break;
    }
    emit('update:operation', operation)
}

function getOperationContent() {

    // [OperationType.StepIncreaseOperation]: { name: "Increase", value: 1 },
    // [OperationType.StepDecreaseOperation]: { name: "Decrease", value: 2 },
    return Object.values(OperationContent)
}

function dataUpdated(event: any) {
    data.value = event
    emit('update:data', event)
}

function delayUpdated(event: number) {
    delay.value = event
    emit('update:delay', event * 60000)// convert to minutes
}

function deviceIdUpdated(event: string) {
    id.value = event
    var device = store.getters["devices/find"](id.value);
    if (device != undefined) {
        emit('update:id', id.value, device.friendly_name)
    }
}

function presetUpdated(event: string) {
    var value = parseInt(event)
    data.value = value
    emit('update:data', value)
}

const featureDevices = computed(() => {
    var devices = store.getters["devices/items"];
    return getFeatureDevices(devices)

    // return createMapFromObject(features, "friendly_name", "id")
});

const deviceList = computed(() => {
    return createMapFromObject(featureDevices.value, "friendly_name", "id")
})

function getPlaceholder(type: string): string {
    if (type == 'binary' || type == 'enum') {
        return 'Select'
    }

    return 'Value'
}

const getFeatureNames = computed(() => {
    return createMapFromObject(features.value, "name", "name")
})

const getItems = computed(() => {

    switch (feature.value.type) {
        case "binary":
        case "enum":
            return Object.values(feature.value.properties)
        default:
            return null
    }
})

const getPresets = computed(() => {
    if (feature.value.presets == undefined) {
        return {}
    }

    return feature.value.presets;
})

</script>

<template>
    <div class="row">
        <div v-if="getPresets" class="col-xl-3 col-md-4">
            <Selector placeholder=" Select device" :items="deviceList" :value="id" alignment="center"
                @update:data="deviceIdUpdated" :disabled="id != ''">
            </Selector>
        </div>
        <div class="col-xl-3 col-md-4">
            <Selector placeholder="Select property" :items="getFeatureNames" :value="property" alignment="center"
                @update:data="propertyUpdated" :disabled="id == ''">
            </Selector>
        </div>
        <div class="col-xl-4 col-md-3">
            <DataInput :type="feature.type" :placeholder="getPlaceholder(feature.type)" :items="getItems" :data="data"
                :disabled="property == ''" @update:data="dataUpdated">
            </DataInput>
        </div>
        <div class="col-xl-2">
            <div class="btn-group">
                <button class="btn btn-default btn-number" type="button" data-bs-toggle="collapse"
                    data-bs-target="#collapseOptions" aria-expanded="false" aria-controls="collapseOptions">
                    <span class="fas fa-angle-double-down"></span>
                </button>
            </div>

        </div>
        <div class="collapse" id="collapseOptions">
            <div class="row pt-2" :disabled="property == ''">
                <div class="col-xl-3 ">
                    <Selector :items="availableOperations" :data="operation" @update:data="operationUpdated">
                    </Selector>
                </div>
                <div v-if="showStepsSelection == true" class="col-xl-3 ">

                    <RadioGroup :items="sth"></RadioGroup>
                </div>
                <!-- <div v-if="feature.presets != null" class="col-xl-3">
                    <Selector placeholder="Presets" :items="getPresets" value="" @update:data="presetUpdated">
                    </Selector>
                </div> -->
                <div class="col-xl-3">
                    <DataInput placeholder="Delay (min)" type="numeric" :data="delay" @update:data="delayUpdated">
                    </DataInput>
                </div>
                <!-- <div v-if="feature.type == 'numeric'" class="col-xl-3 ">
                    <Selector :items="getSteps" :value="step" @update:data="stepUpdated">
                    </Selector>
                </div> -->
            </div>
        </div>
    </div>
</template>
