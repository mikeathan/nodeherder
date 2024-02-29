<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";
import Selector from "../../input/Selector.vue"
import DataInput from "../../input/DataInput.vue"

import { OperationType, resolveObjectOperations } from "../../../contracts/operations"
import { AutomationTriggerAction, AutomationActionStep } from "@/types/automation";
import { clearAction, setDeviceId, setProperty, NumericOperators } from "@/contracts/automations"
import { getDeviceFeaturesByType, getFeatureDevices } from "@/contracts/device";
import { store } from "../../../store/index";
import { Device, Devices, ExposeType } from "@/types/device";
import { KeyyValuePair } from "@/types/types";
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
function addStep() {
    const newStep: AutomationActionStep = {
        operator: "+",
        property: ""
    }
    action.steps.push(newStep);
}

function removeStep(step: AutomationActionStep) {
    action.steps = action.steps.filter((c) => c != step);
}

function deviceSelected(id: string) {
    const device = store.getters["devices/find"](id) as Device;
    action.id = device.id;
    action.friendlyname = device.friendly_name;
}

// Temporary 

const placeholder = ref("value")
function blurChanged() {
    placeholder.value = "value"
}

function focusChanged() {
    placeholder.value = ""
}
</script>
<style scoped>
select.form-select,
input.form-control {
    border: 0;
    outline: 0;
    border-radius: 0%;
    border-bottom: 1px solid white;
    text-align: left;
}

.form-floating>.form-control~label::after {
    background-color: transparent;

}

.form-floating>.form-select~label::after {
    background-color: transparent;
}


/* .form-floatingform-select~label {
    color: grey;
} */


input.form-control:focus,
select.form-select:focus,
:active {
    box-shadow: none;
}


/* 
select.form-select:first-of-type {
    border-bottom: 0px solid white;
}

select.form-select:required:invalid {
    color: gray;
    border-bottom: 1px solid white;
} */

.form-floating>.form-control:focus~label,
.form-floating>.form-control:not(:placeholder-shown)~label,
.form-floating>.form-control~label,
.form-floating>.form-select~label {
    opacity: .6;
    transform: scale(.85) translateY(-.7rem) translateX(.15rem);
}

input.form-control:disabled {
    color: gray;
    background-color: transparent;
}
</style>

<template>
    <!-- action controls -->
    <div class="row pb-3">
        <form class="container">
            <button class="btn btn-light btn-sm" type="button">Delete</button>
            <button class="btn btn-light btn-sm" type="button" @click="addStep" :disabled="action.id == ''">Add
                Step</button>
        </form>
    </div>
    <!-- Testing select box  -->
    <div class="row">
        <div class="form-floating col-sm-5">
            <select required id="dataSelect" class="form-select form-select-solid" v-model="action.id">
                <option value=""> Select </option>
                <option v-for="(value, key) in getFeatureDeviceList" :value="value" :key="value">
                    {{ key }}
                </option>
            </select>
            <label for="dataSelect" class="form-label">Device to trigger</label>
        </div>
    </div>

    <!-- Testing input box  -->
    <div class="row">
        <div class="form-floating col-sm-7">
            <input type=" text" class="form-control" id="dataInput" v-model="action.data">
            <label for="dataInput">Set value</label>
        </div>
    </div>

    <!-- <div class="col-xl-2">
        <div class="btn-group">
            <button class="btn btn-default btn-number" type="button" @click="addStep">
                Add step
    </button>
    </div>
    </div> -->

    <!-- Steps -->
    <div v-if="action.steps.length > 0" class="row pt-3">
        <h5>Steps</h5>
        <div class="row" v-for="step in action.steps">


            <div class="form-floating col-sm-4">
                <select required id="dataSelect" class="form-select form-select-solid" v-model="step.property">
                    <option value=""> Select </option>
                    <option v-for="(value, key) in getFeatureNames" :value="value" :key="value">
                        {{ key }}
                    </option>
                </select>
                <label for="dataSelect" class="form-label">Device property</label>
            </div>
            <div class="col-md-2">
                <DataInput type="enum" :items="NumericOperators" :data="step.operator" alignment="center" @update:data="">
                </DataInput>
            </div>
            <div class="col-md-3">
                <span class="fa fa-trash-alt fa-sm" @click="removeStep(step)">
                </span>
            </div>
        </div>
    </div>
</template>
