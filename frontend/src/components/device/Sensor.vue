<script setup>
import {
    getSensorValue,
    getSensorIcon,
    getSensorName,
} from "../../modules/sensors/sensor-formatter";

const props = defineProps({
    name: String,
    value: Number | Boolean,
});
</script>
<template>
    <div class="me-1">
        <i :class="`fa fa-fw ${getSensorIcon(name)}`"></i>
    </div>
    <div class="flex-shrink-1 flex-grow-1">
        {{ getSensorName(name) }}
    </div>
    <div class="flex-shrink-1">
        {{ getSensorValue(name, value) }}
    </div>
</template>
