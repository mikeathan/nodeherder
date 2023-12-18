<script setup>
import {
    getSensorValue,
    getSensorIcon,
    getSensorName,
} from "../../modules/sensors/sensor-formatter";

import { useStore } from "vuex";
import Slider from "../input/Slider.vue"

const store = useStore();
const props = defineProps({
    name: String,
    value: Number | Boolean,
    unit: String,
    expose: Object
});

function updateValue() {
    var payload = ""
    store.dispatch("device/setValue", payload);
}


// todo:
// if has properties
// check if expose.type 
// if numeric show slider
// if binary show toggle
// for enums dont do anyting for now 

function hasNumericFeatures() {
    return props.expose.properties != null && props.expose.type == "numeric"
}

</script>
<template>
    <div class="me-1">
        <i :class="`fa fa-fw ${getSensorIcon(name)}`"></i>
    </div>

    <div class="flex-shrink-1 flex-grow-1">
        {{ getSensorName(name) }}
    </div>
    <div v-if="value != undefined" class="flex-shrink-1">
        {{ getSensorValue(name, value, unit) }}
        <div v-if="hasNumericFeatures()">
            <Slider value="10" :min="0" :max="10"></Slider>
        </div>
    </div>
    <div v-else>NA</div>
</template>
