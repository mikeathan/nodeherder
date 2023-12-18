<script setup>
import {
    getSensorValue,
    getSensorIcon,
    getSensorName,
    getSensorUnit,
} from "../../modules/sensors/sensor-formatter";

import { useStore } from "vuex";
import Slider from "../input/Slider.vue"
import Toggle from "../input/Toggle.vue"

const store = useStore();
const props = defineProps({
    id: { type: String, require: true },
    expose: Object
});

function updateValue(event) {

    var msg = {
        id: props.id,
        name: props.expose.name,
        value: event
    }

    store.dispatch("devices/setValue", msg);
}

function hasNumericFeatures() {
    return props.expose.properties != null && props.expose.type == "numeric"
}

function hasBinaryFeatures() {
    return props.expose.properties != null && props.expose.type == "binary"
}

function getValue() {
    return getSensorValue(props.expose.name, props.expose.data, props.expose.unit)
}
function getUnit() {
    if (props.expose.unit == undefined) {
        return getSensorUnit(props.expose.name)
    }
    return props.expose.unit
}
</script>
<template>
    <div class="me-1">
        <i :class="`fa fa-fw ${getSensorIcon(props.expose.name)}`"></i>
    </div>

    <div class="flex-shrink-1 flex-grow-1">
        {{ getSensorName(props.expose.name) }}
    </div>
    <div v-if="props.expose.data != undefined" class="flex-shrink-1">
        <div v-if="hasNumericFeatures()">
            <Slider :value="getValue()" :min="props.expose.attributes['min']" :max="props.expose.attributes['max']"
                @update:value="updateValue">
            </Slider>
        </div>
        <div v-else="hasBinaryFeatures()">
            <Toggle :value="getValue()" @update:value="updateValue">
            </Toggle>
        </div>
        <div v-else>
            {{ getValue() }}
            {{ getUnit() }}
        </div>
    </div>
    <div v-else>NA</div>
</template>
