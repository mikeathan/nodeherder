<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";
import DataInput from "../input/DataInput.vue"
import Selector from "../input/Selector.vue"
import { OperationType, resolveObjectOperations } from "../../contracts/operations"
import { setDeviceId, setProperty } from "../../contracts/automations"

import { store } from "../../store/index";
import { Device, Devices } from "@/types/device";
import { KeyyValuePair } from "@/types/types";
import { AutomationTriggerAction } from "@/types/automation";
import { toMillisecs, toMinutes } from '@/modules/formatters/time.formatter'
const props = defineProps({
    action: {
        type: Object as PropType<AutomationTriggerAction>,
        default: {} as AutomationTriggerAction,
        required: true
    },
});


const emit = defineEmits<{
    (e: 'update', action: AutomationTriggerAction): void,
}>()

const action = reactive({ ...props.action })
const device = computed(() => {
    return store.getters["devices/find"](action.id) as Device;
});

const deviceFeatureList = computed(() => {
    var devices = store.getters["devices/listAll"]() as Devices;

    let list: KeyyValuePair<string> = {}
    for (const [key, device] of Object.entries(devices)) {
        for (const [key, expose] of Object.entries(device.exposes)) {
            if (expose.properties != undefined) {
                list[device.friendly_name] = device.id
                break;
            }
        }
    }

    return list
    // return Object.assign({}, ...featureDevices.value.map(f => ({ [f.friendly_name]: f.id })))
})

const getFeatureNames = computed(() => {
    var device = store.getters["devices/find"](action.id) as Device;
    if (device == undefined) {
        return []
    }

    return Object.assign({},
        ...Object.values(device.exposes)
            .filter(f => f.properties != undefined)
            .map(f => ({ [f.name]: f.name })))
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
        return []
    }

    return feature.value.presets;
})

const showPresets = computed<boolean>(() => {
    const device = store.getters["devices/find"](action.id) as Device;
    if (device === undefined) {
        return false
    }

    if (action.property === '') {
        return false
    }

    const feature = device.exposes[action.property];
    switch (+action.operation) {
        case OperationType.StepIncreaseOperation:
        case OperationType.StepDecreaseOperation:
        case OperationType.RotationOperation:
            return false

        case OperationType.NoOperation:
            if (feature.presets != undefined) {
                return true;
            }

            break
    }

    return false
})


const feature = computed(() => {
    if (action.property == '') {
        return []
    }
    var device = store.getters["devices/find"](action.id);
    if (device == undefined) {
        return []
    }
    return device.exposes[action.property];
});


function propertyUpdated(event: any) {
    const value = event;
    if (value == "" || device.value == undefined) {
        action.data = ""
        action.property = ""
        emit('update', action)

        return;
    }
    const newFeature = device.value.exposes[value]
    setProperty(action, value, newFeature.type)
    emit('update', action)
}

function operationUpdated(op: any) {
    action.operation = parseInt(op)
    emit('update', action)
}

function dataUpdated(event: any) {
    action.data = event
    emit('update', action)
}

function delayUpdated(event: number) {
    action.delay = toMillisecs(event)
    emit('update', action)
}

function deviceIdUpdated(event: string) {
    action.id = event
    var device = store.getters["devices/find"](action.id) as Device;
    if (device != undefined) {

        setDeviceId(action, device.id, device.friendly_name)
        emit('update', action)
    }
}

function presetUpdated(event: string) {
    var value = parseInt(event)
    action.data = value;
    emit('update', action)
}


function getPlaceholder(type: string): string {
    if (type == 'binary' || type == 'enum') {
        return 'Select'
    }

    return 'Value'
}

</script>

<template>
    <div class="row">
        <div v-if="getPresets" class="col-xl-3 col-md-4">
            <Selector placeholder=" Select device" :items="deviceFeatureList" :value="action.id" alignment="center"
                @update:data="deviceIdUpdated" :disabled="action.id != ''">
            </Selector>
        </div>
        <div class="col-xl-3 col-md-4">
            <Selector placeholder="Select property" :items="getFeatureNames" :value="action.property" alignment="center"
                @update:data="propertyUpdated" :disabled="action.id == ''">
            </Selector>
        </div>
        <div class="col-xl-4 col-md-3">
            <DataInput :type="feature.type" :placeholder="getPlaceholder(feature.type)" :items="getItems"
                :data="action.data" :disabled="action.property == ''" @update:data="dataUpdated">
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
            <div class="row pt-2" :disabled="action.property == ''">
                <div class="col-xl-3 ">
                    <Selector :items="resolveObjectOperations(feature)" :value="action.operation"
                        @update:data="operationUpdated">
                    </Selector>
                </div>
                <div v-if="showPresets" class="col-xl-3">
                    <Selector placeholder="Presets" :items="getPresets" value="" @update:data="presetUpdated">
                    </Selector>
                </div>
                <div class="col-xl-3">
                    <DataInput placeholder="Delay (min)" type="numeric" :data="toMinutes(action.delay)"
                        @update:data="delayUpdated">
                    </DataInput>
                </div>

            </div>
        </div>
    </div>
</template>
