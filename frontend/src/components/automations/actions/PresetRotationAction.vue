<script setup lang="ts">
import { computed, ref, watchEffect, watch, PropType, reactive } from "vue";
import { AutomationTriggerAction, NumericOperator } from "@/types/automation";
import { getEnumDevices, getDevicePropertiesWithPresets, getPropertiesByExposeType } from "@/contracts/device";
import { store } from "../../../store/index";
import { ExposeTypes } from "@/types/device.type";
import Dropdown from "@/components/controls/Dropdown.vue";
import ButtonPanel from "@/components/controls/ButtonPanel.vue";
import { createSaveDeleteButtonItems } from "../../../configs/automation/trigger-dropdown.config";
import { Device, Devices } from "@/types/device";

const props = defineProps({
    action: {
        type: Object as PropType<AutomationTriggerAction>,
        default: {} as AutomationTriggerAction,
        required: true
    },
});

const action = reactive({ ...props.action })
const buttonPanelItems = computed(() => {
    const actionIsValid = (action.property && action.id);
    return createSaveDeleteButtonItems(
        () => saveAction(),
        () => removeAction(),
        !actionIsValid,
        !actionIsValid)
});

const emit = defineEmits<{
    (e: 'save', action: AutomationTriggerAction): void,
    (e: 'delete', action: AutomationTriggerAction): void,
}>()

function deviceSelected(event: Event) {
    const id = (event.target as HTMLInputElement).value;
    const device = store.getters["devices/find"](id) as Device;
    if (device == undefined) {
        // error
        return;
    }

    action.id = device.id;
    action.friendlyname = device.friendly_name;
    action.property = "";
    action.steps = [];
}

const getEnumDeviceList = computed(() => {
    var devices = store.getters["devices/listAll"]() as Devices;
    if (devices == undefined) {
        return {}
    }
    return getEnumDevices(devices)
})

function getPropertyList(id: string) {
    const device = store.getters["devices/find"](id) as Device;
    if (device == undefined) {
        return {}
    }

    return getDevicePropertiesWithPresets(device);
}

function propertySelected(event: Event) {
    action.property = (event.target as HTMLInputElement).value
}

function saveAction(): void {
    emit('save', action);
}

function removeAction(): void {
    emit('delete', action);
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
        <div class="form-floating col-sm-5">
            <select required id="dataSelect" class="form-select form-select-solid" v-model="action.id"
                @change="deviceSelected">
                <option value=""> Select </option>
                <option v-for="(value, key) in getEnumDeviceList" :value="value" :key="value">
                    {{ key }}
                </option>
            </select>
            <label for="dataSelect" class="form-label">Device to trigger</label>
        </div>
    </div>

    <div class="row pb-2">
        <div class="form-floating col-sm-5">

            <select required id="propertySelector" class="form-select form-select-sm" v-model="action.property"
                @change="propertySelected">
                <option value="">Select property</option>
                <option v-for="property in getPropertyList(action.id)" :value="property" :key="property">
                    {{ property }}
                </option>
            </select>
            <label for="propertySelector" class="form-label">Preset for rotation</label>

        </div>
    </div>

</template>