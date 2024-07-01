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

const props = defineProps({
    id: { type: String, required: true }
});

const deviceSettings = computed(() => {
    if (!store.getters["appconfig/initialized"]() as Boolean) {
        store.dispatch('ws/emit', { event: "loadAppConfig" });
    }
    const settings = store.getters["appconfig/findDeviceSettings"](props.id) as DeviceSettings;
    if (!settings) {
        const newDeviceSettings = createDeviceSettings(props.id)
        store.dispatch('appconfig/saveDeviceSettings', newDeviceSettings)
        return newDeviceSettings
    }

    return settings;
});



</script>
<template>
    device settings: {{ deviceSettings }} for {{ props.id }}
</template>
