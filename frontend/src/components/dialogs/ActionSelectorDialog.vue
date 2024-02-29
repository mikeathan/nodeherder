<script setup lang="ts">
import { ref, watchEffect, onMounted, watch, computed } from 'vue';
import DataInput from "../input/DataInput.vue"
import Selector from "../input/Selector.vue"
import { OperationType, resolveObjectOperations } from "../../contracts/operations"
import { clearAction, setDeviceId, setProperty } from "../../contracts/automations"
import { getDeviceFeaturesByType } from "../../contracts/device"
import { store } from "../../store/index";
import { Device, Devices, ExposeType } from "@/types/device";
import { KeyyValuePair } from "@/types/types";
import { AutomationTriggerAction } from "@/types/automation";
import { toMillisecs, toMinutes } from '@/modules/formatters/time.formatter'
import { ExposeTypes } from "@/types/device.type";
import { Modal } from 'bootstrap'

const props = defineProps<{
    show: boolean
}>()

const emit = defineEmits<{
    (e: 'update', actionType: string, deviceId: string, deviceProperty: string): void,
    (e: 'close', exit: boolean): void,
}>()

const actionTypes = {
    "Trigger action": "trigger_action",
    "Step action": "step_action"
}

const actionType = ref<string>("")
const deviceId = ref<string>("")
const deviceProperty = ref<string>("")

const closeRef = ref<HTMLButtonElement | null>(null);
const modalRef = ref<HTMLElement | null>(null)
const showDialog = ref<boolean>(false)
let modal: Modal

watch(
    () => actionType.value,
    () => {
        deviceId.value = "";
        deviceProperty.value = "";
    }, { immediate: true }
)
watch(
    () => deviceId.value,
    () => {
        deviceProperty.value = "";
    }, { immediate: true }
)


watch(
    () => props.show,
    () => {
        showDialog.value = props.show
        if (showDialog.value) {
            modal.show()
        } else {
            modal.hide()
        }
    }
);

onMounted(() => {
    if (modalRef.value) {
        modal = new Modal(modalRef.value)
    }
})

function isValid(): boolean {
    return actionType.value != '' && deviceId.value != '' && deviceProperty.value != '';
}

const getDeviceFeatureNames = computed(() => {

    var device = store.getters["devices/find"](deviceId.value) as Device;
    if (device == undefined) {
        return {}
    }

    return getDeviceFeaturesByType(device, ExposeTypes.Numeric);;
})

const deviceFeatureList = computed(() => {
    var devices = store.getters["devices/listAll"]() as Devices;

    let list: KeyyValuePair<string> = {}
    for (const [key, device] of Object.entries(devices)) {
        for (const [key, expose] of Object.entries(device.exposes)) {
            if (expose.properties != undefined) {
                list[device.friendly_name] = device.id
                break;
            }
        }
    }

    return list
})

function add(event: Event) {
    emit('update', actionType.value, deviceId.value, deviceProperty.value);
    close()
}

function clear(): void {
    actionType.value = "";
    deviceId.value = "";
    deviceProperty.value = "";
}

function close(): void {
    clear();
    emit('close', false)
}

</script>
<template>
    <div class="modal fade" tabindex="-1" aria-hidden="true" ref="modalRef">
        <div class="modal-dialog">
            <div class="modal-content">
                <div class="modal-header">
                    <h5 class="modal-title">Action Selector</h5>
                </div>
                <div class="modal-body">
                    <table class="table responsive ">
                        <thead>
                            <tr>
                                <th style="text-align: center;" scope="col">Action Type</th>
                                <th style="text-align: center;" scope="col">Device</th>
                                <th style="text-align: center;" scope="col">Expose</th>
                            </tr>
                        </thead>
                        <tbody>

                            <tr>
                                <td class="col-xl-3 col-md-4  col-sm-6">
                                    <Selector placeholder="Select" :items="actionTypes" :value="actionType"
                                        @update:data="(v) => actionType = v">
                                    </Selector>
                                </td>
                                <td class="col-xl-3 col-md-4 col-sm-6">
                                    <Selector :disabled="actionType == ''" placeholder="Select" :items="deviceFeatureList"
                                        :value="deviceId" @update:data="(v) => deviceId = v">
                                    </Selector>
                                </td>
                                <td class="col-xl-3 col-md-4 col-sm-6">
                                    <Selector :disabled="deviceId == ''" placeholder="Select" :items="getDeviceFeatureNames"
                                        :value="deviceProperty" @update:data="(v) => deviceProperty = v">
                                    </Selector>
                                </td>
                            </tr>
                        </tbody>
                    </table>
                </div>
                <div class="modal-footer">
                    <button type="button" class="btn btn-primary" :disabled="isValid() == false" @click="add">Add</button>
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal" @click="close">Close</button>
                </div>
            </div>
        </div>
    </div>
</template>