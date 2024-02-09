<script setup lang="ts">
import { store } from "../../store/index";
import { computed, ref } from "vue";
import { useRouter } from "vue-router";

import { Device, Expose, ExposeAttributes } from "@/types/device";
import { ExposeTypes } from "@/types/device.type";
import { getExposeAttribute, getExposeBinaryProperty, hasSupportedExposeBinaryProperties, getExposePresets } from "../../contracts/device";

import Slider from "../input/Slider.vue";
import Toggle from "../input/Toggle.vue";
import RadioGroup from "../input/RadioGroup.vue";

const props = defineProps({
    id: { type: String, required: true }
});

const device = computed(() => {
    return store.getters["devices/find"](props.id) as Device;
});

function testUpdated(expose: Expose, value: any) {
    console.log("devicexpose updated:", value, typeof value)
    var msg = {
        id: props.id,
        name: expose.name,
        value: value
    }
    store.dispatch("devices/setValue", msg);
}

</script>
<template>
    <div class="row border-bottom py-1 w-100 align-items-center" v-for="( expose, index ) in  device.exposes "
        :item="expose">
        <dl class="col-12 col-md-3">
            <dt><strong> {{ expose.name }}</strong></dt>
            <dd><small> {{ expose.description }} </small></dd>
        </dl>

        <div v-if="expose.properties == null">
            {{ expose.data }} {{ expose.unit }}
        </div>
        <div v-else-if="expose.type == ExposeTypes.Numeric" class="input-group align-items-center">
            expose: {{ expose.data }}
            <RadioGroup v-if="expose.presets != null" :items="(expose.presets as any)" :value="expose.data"
                @update="v => testUpdated(expose, v)"></RadioGroup>

            <Slider :value="expose.data" :min="getExposeAttribute(expose, 'min')" :max="getExposeAttribute(expose, 'max')"
                @update:value="v => testUpdated(expose, v)">
            </Slider>
            <input class="form-control ms-1" type="number" :value="expose.data" style="max-width: 100px;">
        </div>
        <div v-else-if="expose.type == ExposeTypes.Binary">

            <Toggle :enabled="getExposeBinaryProperty(expose)" :disabled="!hasSupportedExposeBinaryProperties(expose)">
            </Toggle>
        </div>
        <div v-else-if="expose.type == ExposeTypes.Enum">
            <!-- <RadioGroup :items="expose.data" value="test"></RadioGroup> -->

        </div>
    </div>
</template>
