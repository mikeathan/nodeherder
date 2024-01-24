<script setup lang="ts">
import { computed, ref, watchEffect, watch } from "vue";
import DataInput from "../input/DataInput.vue"
import Selector from "../input/Selector.vue"
import { OperationType, resolveObjectOperations } from "../../contracts/mappers/action-operation.resolver"
import { store } from "../../store/index";
import { Device, Devices, Expose } from "@/types/device";
import { ExposeTriggerWrapper } from "@/contracts/automations";

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


const emit = defineEmits<{
    (e: 'update:id', id: string, name: string): void,
    (e: 'update:property', property: string): void,
    (e: 'update:data', data: any): void,
    (e: 'update:delay', data: number): void,
    (e: 'update:operation', data: number): void,
}>()

const id = ref<string>("")
const property = ref<string>("");
const data = ref<any>("");
const delay = ref<number | null>(null);
const operation = ref<number>(0)

const showPresets = computed<boolean>(() => {
    const device = store.getters["devices/find"](id.value) as Device;
    if (device === undefined) {
        return false
    }

    if (property.value === '') {
        return false
    }

    const feature = device.exposes[property.value];
    switch (+operation.value) {
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
    var device = store.getters["devices/find"](id.value) as Device;
    if (device == undefined) {
        return []
    }

    return Object.values(device.exposes).filter(f => f.properties != undefined) as Array<Expose>
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
    if (feature.value.type == 'binary' || feature.value.type == "enum") {
        data.value = ""
    } else {
        data.value = 0
    }

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
    return device.exposes[property.value];

});

// NOTE
// its messy but we need it for now as device is not a defined class
// keep it for now until refactoring 

function operationUpdated(op: any) {
    operation.value = parseInt(op)
    emit('update:operation', operation.value)
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
    var device = store.getters["devices/find"](id.value) as Device;
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
    var devices = store.getters["devices/listAll"]() as Devices;

    return Object.entries(devices)
        .filter(([key, value]) => value.properties != null)
        .map((k) => k[1])
});

const deviceList = computed(() => {
    return Object.assign({}, ...featureDevices.value.map(f => ({ [f.friendly_name]: f.id })))

})

function getPlaceholder(type: string): string {
    if (type == 'binary' || type == 'enum') {
        return 'Select'
    }

    return 'Value'
}

const getFeatureNames = computed(() => {
    return Object.assign({}, ...features.value
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
                    <Selector :items="resolveObjectOperations(feature)" :value="operation" @update:data="operationUpdated">
                    </Selector>
                </div>
                <div v-if="showPresets" class="col-xl-3">
                    <Selector placeholder="Presets" :items="getPresets" value="" @update:data="presetUpdated">
                    </Selector>
                </div>
                <div class="col-xl-3">
                    <DataInput placeholder="Delay (min)" type="numeric" :data="delay" @update:data="delayUpdated">
                    </DataInput>
                </div>

            </div>
        </div>
    </div>
</template>
