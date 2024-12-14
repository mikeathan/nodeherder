<script setup lang="ts">
import {
    getSensorValue,
    getSensorIcon,
    getSensorName,
    getSensorUnit,
} from "../../modules/formatters/sensor-formatter";
import { getExposeProperty, getExposeAttribute } from "../../contracts/device";
import { PropType } from "vue";
import { store } from "../../store/index";
import { Expose } from "@/types/device";
import Range from "../input/Range.vue"
import Toggle from "../input/Toggle.vue"

const props = defineProps({
    id: { type: String, require: true },
    expose: { type: Object as PropType<Expose>, default: {} as Expose },
});

function updateValue(event: any): void {
    var msg = {
        id: props.id,
        name: props.expose.name,
        value: event
    }

    store.dispatch("devices/setValue", msg);
}

function hasNumericFeatures(): Boolean {
    return props.expose.properties != null && props.expose.type == "numeric"
}

function hasBinaryFeatures() {
    return props.expose.properties != null && props.expose.type == "binary"
}

function getValue() {
    return getSensorValue(props.expose.data)
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
        <i :class="`fa fa-fw ${getSensorIcon(props.expose.name, props.expose.data)}`"></i>
    </div>
    <div class="flex-grow-1">
        {{ getSensorName(props.expose.name) }}
    </div>
    <div v-if="hasNumericFeatures()" class='col-7'>
        <Range :value="getValue()" :min="getExposeAttribute(props.expose, 'min')"
            :max="getExposeAttribute(props.expose, 'max')" @update="updateValue">
        </Range>
    </div>
    <div v-else-if="hasBinaryFeatures()">
        <Toggle :value="props.expose.data" :valueOn="getExposeProperty(props.expose, 'on')"
            :valueoff="getExposeProperty(props.expose, 'off')" @update="(v) => updateValue(v)">
        </Toggle>
    </div>
    <div v-else>
        {{ getValue() }}
        {{ getUnit() }}
    </div>
    <!-- <div v-else>NA</div> -->
</template>
