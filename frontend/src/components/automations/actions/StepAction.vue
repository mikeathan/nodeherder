<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";
import Selector from "../../input/Selector.vue"
import { OperationType, resolveObjectOperations } from "../../../contracts/operations"
import { clearAction, setDeviceId, setProperty } from "@/contracts/automations"
import { getDeviceFeaturesByType, getFeatureDevices } from "@/contracts/device";
import { store } from "../../../store/index";
import { Device, Devices, ExposeType } from "@/types/device";
import { KeyyValuePair } from "@/types/types";
import { AutomationTriggerAction } from "@/types/automation";
import { toMillisecs, toMinutes } from '@/modules/formatters/time.formatter'
import { ExposeTypes } from "@/types/device.type";


// New step action control is needed
// step action = brightness increase by value // brightness decrease by value
// brightness increase by step  action * by value/ brightness decrease by step  action * by value/
// in action.data we store the value 

// select expose property is required as is the value we modify
// then we add steps 
// eg property name to use the value. rquirement is only numerica properties can be used
// operator to use - + /

// brighness  = brightness + value
// if we have multi steps 
// brightness = brightness + (action_time * value



// enum /preset rotation can happen in default action control ? 

const props = defineProps({
    action: {
        type: Object as PropType<AutomationTriggerAction>,
        default: {} as AutomationTriggerAction,
        required: true
    },
});

const emit = defineEmits<{
    (e: 'update', action: AutomationTriggerAction): void,
}>()

const action = reactive({ ...props.action })
const device = computed(() => {
    return store.getters["devices/find"](action.id) as Device;
});

const getFeatureDeviceList = computed(() => {
    var devices = store.getters["devices/listAll"]() as Devices;
    if (device == undefined) {
        return {}
    }
    return getFeatureDevices(devices)
})

const getFeatureNames = computed(() => {
    var device = store.getters["devices/find"](action.id) as Device;
    if (device == undefined) {
        return {}
    }

    return getDeviceFeaturesByType(device, ExposeTypes.Numeric);
})

function deviceSelected(id: string) {
    const device = store.getters["devices/find"](id) as Device;
    action.id = device.id;
    action.friendlyname = device.friendly_name;
}
</script>

<template>
    <div class="row">
        <div class="col-xl-4 col-md-3">
            <Selector placeholder="Select device" :items="getFeatureDeviceList" :value="action.id" alignment="center"
                @update:data="deviceSelected" :disabled="action.id != ''">
            </Selector>
        </div>
        <div class="col-xl-3 col-md-4">
            <Selector placeholder="Select property" :items="getFeatureNames" :value="action.property" alignment="center"
                @update:data="" :disabled="action.id == ''">
            </Selector>
        </div>

        TODO:Add steps
    </div>
</template>
