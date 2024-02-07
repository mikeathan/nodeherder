<script setup lang="ts">
import { store } from "../../store/index";
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import LastSeen from "../device/LastSeen.vue";
import PowerSource from "../device/PowerSource.vue";
import ConnectionType from "../device/ConnectionType.vue";
import RenameDeviceDialog from "../dialogs/RenameDeviceDialog.vue";
import { Device, Expose } from "@/types/device";
import { ExposeTypes } from "@/types/device.type";
import Slider from "../input/Slider.vue";

const props = defineProps({
    id: { type: String, required: true }
});

const device = computed(() => {
    return store.getters["devices/find"](props.id) as Device;
});

const Exposes = computed(() => {

    let list: Array<Expose> = []
    Object.values(device.value.exposes).forEach(expose => {

        list.push(expose)
    });


    return list;
});
</script>
<template>
    <div v-if="device != null" class="panel">
        <div className="panel-header d-flex flex-row">
        </div>
        <div class="h1 align-self-center">
            {{ device.friendly_name }}
        </div>
        if it has properties then its a feature <br>
        if its numeric add slider <br>
        if its binary add toggle <br>
        if enum add radio group <br>
        if no properties then just display values <br>
        <div class="row border-bottom py-1 w-100 align-items-center" v-for="expose in Exposes">
            <dl class="col-12 col-md-3">
                <dt><strong> {{ expose.name }}</strong></dt>
                <dd><small> {{ expose.description }} </small></dd>
            </dl>

            <div v-if="expose.properties == null" class="col-12 col-md-9">
                {{ expose.data }} {{ expose.unit }}
            </div>
            <div v-else>
                <div v-if="expose.type == ExposeTypes.Numeric">
                    <Slider :value="expose.data" :min="expose.attributes['min']" :max="expose.attributes['max']"
                        @update:value="updateValue">
                    </Slider>
                </div>
            </div>
        </div>

    </div>
</template>
