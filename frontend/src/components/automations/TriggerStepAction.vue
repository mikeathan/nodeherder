<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";
import DataInput from "../input/DataInput.vue"
import Selector from "../input/Selector.vue"
import { OperationType, resolveObjectOperations } from "../../contracts/operations"
import { clearAction, setDeviceId, setProperty } from "../../contracts/automations"
import { getDeviceFeaturesByType } from "../../contracts/device";
import { store } from "../../store/index";
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


const getFeatureNames = computed(() => {
    var device = store.getters["devices/find"](action.id) as Device;
    if (device == undefined) {
        return []
    }

    return getDeviceFeaturesByType(device, ExposeTypes.Numeric);
})

</script>

<template>
    <div class="row">
        <div class="col-xl-3 col-md-4">

            we ned a action selector dialog
            has action type , device to use and property to update

        </div>
        <!-- <span v-if="action.id != ''" class="fa fa-trash-alt fa-sm" @click="(v) => clear()
            ">
        </span> -->
    </div>
</template>
