<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";
import DataInput from "../../input/DataInput.vue"
import Selector from "../../input/Selector.vue"
import { OperationType, resolveObjectOperations } from "@/contracts/operations"
import { clearAction, setDeviceId, setProperty } from "@/contracts/automations"
import { store } from "@/store/index";
import { Device, Devices, ExposeType } from "@/types/device";
import { getFeatureDevices, getDeviceFeatures } from "@/contracts/device";
import { KeyyValuePair } from "@/types/types";
import { AutomationTriggerAction } from "@/types/automation";
import { toMillisecs, toMinutes } from '@/modules/formatters/time.formatter'
import { ExposeTypes } from "@/types/device.type";
import ButtonPanel from "@/components/controls/ButtonPanel.vue";
import { createSaveDeleteButtonItems } from "../../../configs/automation/trigger-dropdown.config";
import DeviceSelector from "@/components/controls/DeviceSelector.vue";
import { featureDevicesFilter } from "@/configs/automation/device.config";
const props = defineProps({
    action: {
        type: Object as PropType<AutomationTriggerAction>,
        default: {} as AutomationTriggerAction,
        required: true
    },
});

const emit = defineEmits<{
    (e: 'update', action: AutomationTriggerAction): void,
    (e: 'save', action: AutomationTriggerAction): void,
    (e: 'delete', action: AutomationTriggerAction): void,
}>()


const action = reactive({ ...props.action })
const buttonPanelItems = computed(() => {
    const isActionValid = action.data && action.property && action.id;

    return createSaveDeleteButtonItems(
        () => saveAction(),
        () => removeAction(),
        !isActionValid,
        !isActionValid)
});

const deviceFeatureList = computed(() => {
    var devices = store.getters["devices/listAll"]() as Devices;
    if (devices == undefined) {
        return {}
    }
    return getFeatureDevices(devices)
})
function deviceSelected(id: string, friendlyName: string) {
    console.log("deviceSelected", id, friendlyName)
    action.id = id;
    action.friendlyname = friendlyName;

    // reset
    action.property = '';
    action.data = null
    action.delay = null;
    action.steps = [];
}

function dataInputChange(event: Event) {
    let value = (event.target as HTMLInputElement).value;
    if (feature.value.type == 'numeric') {
        value = value.replace(/[^\d+$]/, '');
        action.data = parseInt(value)
    }
}

function delayInputChange(event: Event) {
    const value = (event.target as HTMLInputElement).value;
    action.delay = parseInt(value);
}

function getPropertyList(id: string) {
    const device = store.getters["devices/find"](id) as Device;
    if (device == undefined) {
        return {}
    }

    return getDeviceFeatures(device);
}


function propertySelected(event: Event) {
    action.property = (event.target as HTMLInputElement).value

    // reset
    action.data = null
    action.delay = null;
}

function presetSelected(event: Event) {
    const value = (event.target as HTMLInputElement).value
    action.data = parseInt(value);
}

const feature = computed(() => {
    if (action.property == '') {
        return []
    }
    var device = store.getters["devices/find"](action.id);
    if (device == undefined) {
        return []
    }
    return device.exposes[action.property];
});


const getPresets = computed(() => {
    if (feature.value.presets == undefined) {
        return []
    }

    return feature.value.presets;
})

const showPresets = computed<boolean>(() => {
    const device = store.getters["devices/find"](action.id) as Device;
    if (device === undefined) {
        return false
    }

    if (action.property === '') {
        return false
    }

    const feature = device.exposes[action.property];
    switch (+action.operation) {
        case OperationType.StepIncreaseOperation:
        case OperationType.StepDecreaseOperation:
        case OperationType.RotationOperation:
            return false

        case OperationType.NoOperation:
            if (feature.presets != undefined) {
                return true;
            }

            break
    }

    return false
})


function saveAction() {

    // TODO:
    // action.delay = toMillisecs(num)
}
function removeAction() {


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
    <div class="row pb-3">
        <ButtonPanel :buttons="buttonPanelItems"></ButtonPanel>
    </div>


    <div class="row pb-2">
        <DeviceSelector @updated="deviceSelected" :filter="featureDevicesFilter()"></DeviceSelector>
        <!-- <div class="form-floating col-sm-5">
            <select required id="deviceSelector" class="form-select form-select-solid" v-model="action.id"
                @change="deviceSelected">
                <option value=""> Select </option>
                <option v-for="(value, key) in deviceFeatureList" :value="value" :key="value">
                    {{ key }}
                </option>
            </select>
            <label for="deviceSelector" class="form-label">Device to trigger</label>
        </div> -->
    </div>

    <div class="row pb-2">
        <div class="form-floating col-sm-5">
            <select required id="propertySelector" class="form-select form-select-sm" v-model="action.property"
                @change="propertySelected" :disabled="action.id == ''">
                <option value=""> Select </option>
                <option v-for="property in getPropertyList(action.id)" :value="property" :key="property">
                    {{ property }}
                </option>
            </select>
            <label for="propertySelector" class="form-label">Expose</label>
        </div>

        <div v-if="showPresets" class="form-floating col-sm-5">
            <select required id="presetsSelector" class="form-select form-select-sm" @change="presetSelected">
                <option value=""> Select </option>
                <option v-for="(value, key) in getPresets" :value="value" :key="key">
                    {{ key }}
                </option>
            </select>
            <label for="presetsSelector" class="form-label">Expose presets</label>
        </div>
    </div>


    <div class="row">
        <div class="form-floating col-sm-3">
            <input type="text" class="form-control" id="dataInput" v-model="action.data" @input="dataInputChange"
                :disabled="action.property == ''">
            <label for="dataInput">Set value</label>
        </div>

        <!-- add it in a dropdown -->
        <div class="form-floating col-sm-2">
            <input type="text" class="form-control" id="delayInput" v-model="action.delay" @input="delayInputChange"
                :disabled="action.property == ''">
            <label for="delayInput">Delay in minutes</label>
        </div>
    </div>


</template>
