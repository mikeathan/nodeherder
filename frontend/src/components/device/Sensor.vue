<script setup>
import {
    getSensorValue,
    getSensorIcon,
    getSensorName,
} from "../../modules/sensors/sensor-formatter";

const props = defineProps({
    name: String,
    value: Number | Boolean,
    unit: String,
    expose: Object
});

// todo:
// if has properties
// check if expose.type 
// if numeric show slider
// if binary show toggle
// for enums dont do anyting for now 

function properties() {
    return expose.properties
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

        <div v-if="props.expose.type == 'numeric'">{{ props.expose }}</div>
    </div>
    <div v-else>NA</div>
</template>
