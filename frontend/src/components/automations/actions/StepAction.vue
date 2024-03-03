<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";
import Selector from "../../input/Selector.vue"
import DataInput from "../../input/DataInput.vue"

import { AutomationTriggerAction, AutomationActionStep, NumericOperator } from "@/types/automation";
import { NumericOperators, StepAction } from "@/contracts/automations"
import { getDeviceFeaturesByType, getFeatureDevices, getDevicesFeaturesByType } from "@/contracts/device";
import { store } from "../../../store/index";
import { Device, Devices } from "@/types/device";
import { ExposeTypes } from "@/types/device.type";
import { KeyyValuePair } from "@/types/types";


const props = defineProps({
    action: {
        type: Object as PropType<AutomationTriggerAction>,
        default: {} as AutomationTriggerAction,
        required: true
    },
});

const emit = defineEmits<{
    (e: 'save', action: AutomationTriggerAction): void,
    (e: 'delete', action: AutomationTriggerAction): void,
}>()

const action = reactive({ ...props.action })
function getStepDevicesList(step: AutomationActionStep) {

    var devices = store.getters["devices/listAll"]() as Devices;
    if (action.steps.length == 1) {
        devices = devices.filter(d => d.id == action.id)

        step.id = action.id
    }

    return devices;
}

function getStepPropertyList(step: AutomationActionStep) {
    const device = store.getters["devices/find"](step.id) as Device;
    if (device == undefined) {
        return {}
    }

    return getDeviceFeaturesByType(device, ExposeTypes.Numeric);
}

const getStepPropertySelectionList = computed(() => {

    // TODO: refactor
    let deviceMap: KeyyValuePair<string[]> = {};

    var devices = store.getters["devices/listAll"]() as Devices;
    if (devices == undefined) {
        deviceMap[""] = [];
        return deviceMap;
    }

    if (action.steps.length == 1) {
        devices = devices.filter(d => d.id == action.id)
    }

    for (const [key, device] of Object.entries(devices)) {
        for (const [key, expose] of Object.entries(device.exposes)) {
            if (expose.type == ExposeTypes.Numeric) {
                if (device.friendly_name in deviceMap == false) {
                    deviceMap[device.friendly_name] = [];
                }
                deviceMap[device.friendly_name].push(expose.name);
            }
        }
    }

    return deviceMap;
})

const getFeatureDeviceList = computed(() => {
    var devices = store.getters["devices/listAll"]() as Devices;
    if (devices == undefined) {
        return {}
    }
    return getFeatureDevices(devices)
})
const getFeatureNames = computed(() => {
    const device = store.getters["devices/find"](action.id) as Device;
    if (device == undefined) {
        return {}
    }

    return getDeviceFeaturesByType(device, ExposeTypes.Numeric);
})

function saveAction(): void {
    emit('save', action);
}

function removeAction(): void {
    emit('delete', action);
}

function addStep(numericOperator: NumericOperator) {
    const newStep: AutomationActionStep = {
        operator: numericOperator,
        property: ""
    }
    action.steps.push(newStep);
}

function removeStep(step: AutomationActionStep) {
    action.steps = action.steps.filter((c) => c != step);
}

function deviceSelected(event: Event) {
    const id = (event.target as HTMLInputElement).value;
    const device = store.getters["devices/find"](id) as Device;
    if (device == undefined) {
        // error
        return;
    }
    action.id = device.id;
    action.friendlyname = device.friendly_name;
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
    background-image: none;
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

select.form-select:hover:not([disabled]) {
    background-image: url("data:image/svg+xml;charset=utf-8,%3Csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 16 16%27%3E%3Cpath fill=%27none%27 stroke=%27%23d4d6d9%27 stroke-linecap=%27round%27 stroke-linejoin=%27round%27 stroke-width=%272%27 d=%27m2 5 6 6 6-6%27/%3E%3C/svg%3E");
    box-shadow: none;
}

select.form-select:first-of-type {
    border-bottom: 0px solid white;
}

select.form-select:required:invalid {
    color: gray;
    border-bottom: 1px solid white;
}

.form-floating>.form-control:focus~label,
.form-floating>.form-control:not(:placeholder-shown)~label,
.form-floating>.form-control~label,
.form-floating>.form-select~label {
    opacity: .6;
    transform: scale(.85) translateY(-.7rem) translateX(.15rem);
}

select.form-select,
input.form-select:disabled {
    color: gray;
    background-color: transparent;
}
</style>

<template>
    <!-- Edit mode -->
    <!-- action controls -->
    <div class="row pb-3">
        <form class="container">
            <button class="btn btn-light btn-sm" type="button" @click="saveAction">Save</button>
            <button class="btn btn-light btn-sm" type="button" @click="removeAction">Delete</button>
            <button type="button" class="btn btn-light btn-sm" data-bs-toggle="dropdown" :disabled="action.id == ''">Add
                Operation</button>
            <ul class="dropdown-menu">
                <li v-for="operator in NumericOperators">
                    <a @click="addStep(operator as NumericOperator)" class="dropdown-item" data-toggle="dropdown">
                        {{
                operator
            }}</a>
                </li>
            </ul>
        </form>
    </div>

    <!-- Testing select box  -->
    <div class="row pb-2">
        <div class="form-floating col-sm-5">
            <select required id="dataSelect" class="form-select form-select-solid" v-model="action.id"
                :disabled="action.id != ''" @change="deviceSelected">
                <option value=""> Select </option>
                <option v-for="(value, key) in getFeatureDeviceList" :value="value" :key="value">
                    {{ key }}
                </option>
            </select>
            <label for="dataSelect" class="form-label">Device to trigger</label>
        </div>friendly name but on chnage event
    </div>

    <!-- Testing input box  -->
    <div class="row">
        <div class="form-floating col-xl-7">
            <input type="text" class="form-control" id="dataInput" v-model="action.data">
            <label for="dataInput">Set value</label>
        </div>
    </div>

    <!-- Steps -->
    <div v-if="action.steps.length > 0" class="row pt-3">
        <h5>Operations</h5>
        <div class="row" v-for="step in action.steps">
            <div class=" col-xl-1">
                {{ step.operator }}
            </div>

            we need combo for device that once selected is text. also we load devices and key is device,id and vlaue is
            friendly_name
            <div class="col-xl-4">
                some device
            </div>
            <div class="col col-xl-4">
                <select required class="form-select form-select-sm" v-model="step.id">
                    <option value=""> Select device</option>
                    <option v-for="device in getStepDevicesList(step)" :value="device.id" :key="device.id">
                        {{ device.friendly_name }}
                    </option>
                </select>
            </div>
            <div class="col col-xl-6">
                <select required class="form-select form-select-sm" v-model="step.property">
                    <option value="">Some property</option>
                    <option v-for="property in getStepPropertyList(step)" :value="property" :key="property">
                        {{ property }}
                    </option>
                </select>
            </div>
            <div class="col-xl-1">
                <span class="fa fa-trash-alt fa-sm" @click="removeStep(step)">
                </span>
            </div>
        </div>
    </div>
</template>
