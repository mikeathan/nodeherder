<script setup lang="ts">
import { store } from "../../store/index";
import { computed, ref } from "vue";
import { Device, Expose } from "@/types/device";
import { ExposeTypes } from "@/types/device.type";
import { getExposeAttribute, getExposeProperty } from "../../contracts/device";

import Slider from "../input/Slider.vue";
import Toggle from "../input/Toggle.vue";
import RadioGroup from "../input/RadioGroup.vue";
import { getSensorUnit, getSensorValue } from "@/modules/formatters/sensor-formatter";
import { DeviceSettings } from "@/types/settings";
import { createDeviceSettings } from "@/contracts/settings";
import InputBox from "../input/InputBox.vue";

const props = defineProps({
    id: { type: String, required: true }
});

const deviceSettings = computed(() => {
    if (!store.getters["appconfig/initialized"]() as Boolean) {
        store.dispatch('ws/emit', { event: "loadAppConfig" });
    }

    const settings = store.getters["appconfig/findDeviceSetting"](props.id) as DeviceSettings;
    if (!settings) {
        const newDeviceSettings = createDeviceSettings(props.id)
        store.dispatch('appconfig/saveDeviceSettings', newDeviceSettings)
        return newDeviceSettings
    }

    return settings;
});


const settingsElements = computed(() => {

    return Object.entries(deviceSettings.value);
});

function updateValue(propName: any, propValue: any) {
    settingsElements.value[propName] = propValue;

    console.log("uodate value ", propName, settingsElements.value[propName])

    //store.dispatch('appconfig/saveDeviceSettings', deviceSettings.value)
}

function save() {
    console.log("save ", deviceSettings.value)
    store.dispatch('appconfig/saveDeviceSettings', deviceSettings.value)
}


</script>
<template>
    <form>

        <div v-for="([key, value]) in settingsElements">
            <label>{{ key }}</label>

            <div v-if="typeof value === 'boolean'">
                <Toggle :minimal="false" :value="value" valueOn="enable" valueoff="disable"
                    @update="(v) => updateValue(key, v)">
                </Toggle>
            </div>
            <div v-else>
                <InputBox @updated="(v) => updateValue(key, v)" :value="value">
                </InputBox>
            </div>
        </div>
        <button :onclick="save()">Save</button>
    </form>
</template>
