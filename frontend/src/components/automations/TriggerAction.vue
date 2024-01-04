<script setup lang="ts">
import { useStore } from "vuex";
import { computed, ref, watchEffect, watch } from "vue";
import DataInput from "../input/DataInput.vue"
import Selector from "../input/Selector.vue"
import { getFeatureExposes, getFeatureDevices, createMapFromObject } from "../../modules/convert"
import { Steps } from "../../models/automation"

const props = defineProps({
    id: {
        type: String,
        default: ''
    },
    property: String,
    data: null,
    delay: null,
    step: null,
});
const id = ref<string>("")
const property = ref<string>("");
const data = ref<any>("");
const delay = ref<number | null>(null);
const step = ref<number>(0)
const store = useStore();

const emit = defineEmits<{
    (e: 'update:id', id: string, name: string): void,
    (e: 'update:property', property: string): void,
    (e: 'update:data', data: any): void,
    (e: 'update:delay', data: number): void,
    (e: 'update:step', data: number): void,

}>()

watchEffect(() => id.value = props.id);
watchEffect(() => data.value = props.data);

watch(
    () => props.step,
    () => {
        step.value = props.step
        if (step.value == null) {
            step.value = Steps[0].value
        }
    },
    { immediate: true }
);
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
    step.value = 0;

    // reset data
    for (const [key, feature] of Object.entries(features.value)) {
        if (feature.name == value) {
            if (feature.type == 'binary' || feature.type == "enum") { // TODO: refactor/cleanup
                data.value = ""
            } else {
                data.value = 0
            }
        }
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
    if (device.exposes[property.value] == undefined) {

        return []
    }

    return device.exposes[property.value];
});

function dataUpdated(event: any) {
    data.value = event
    emit('update:data', event)
}

function stepUpdated(event: any) {
    var value = parseInt(event)
    step.value = value
    emit('update:step', value)
}

function delayUpdated(event: number) {
    delay.value = event
    emit('update:delay', event * 60000)// convert to minutes
}

function idUpdated(event: string) {
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

const getSteps = computed(() => {
    return createMapFromObject(Steps, "name", "value");
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
                @update:data="idUpdated" :disabled="id != ''">
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
                <div v-if="feature.presets != null" class="col-xl-3">
                    <Selector placeholder="Presets" :items="getPresets" value="" @update:data="presetUpdated">
                    </Selector>
                </div>
                <div class="col-xl-3">
                    <DataInput placeholder="Delay (min)" type="numeric" :data="delay" @update:data="delayUpdated">
                    </DataInput>
                </div>
                <div v-if="feature.type == 'numeric'" class="col-xl-3 ">
                    <Selector :items="getSteps" :value="step" @update:data="stepUpdated">
                    </Selector>
                </div>
            </div>
        </div>
    </div>
</template>
