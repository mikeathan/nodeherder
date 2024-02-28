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

const emit = defineEmits(['update:name', 'close']);


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

onMounted(() => {
    if (modalRef.value) {
        modal = new Modal(modalRef.value)
    }
})

// https://shzhangji.com/blog/2022/06/11/use-bootstrap-v5-in-vue3-project/

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

function deviceSelected() {
    // action.id = event
    // var device = store.getters["devices/find"](id) as Device;
    // if (device != undefined) {

    //     // setDeviceId(action, device.id, device.friendly_name)
    // }
}
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
    // emit('update:name', friendlyName.value);
    close()
}

function clear(): void {
    actionType.value = "";
    deviceId.value = "";
    deviceProperty.value = "";
}

function close(): void {
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

                    <div class="container-fluid p-0 h-100">
                        <div class="card col-xl-5 col-md-6 col-sm-3">
                            <div class="card-header ">
                                <div class="pt-3 ">
                                    <label class="form-check-label">Action Type</label>
                                    <Selector placeholder="Select" :items="actionTypes" :value="actionType"
                                        @update:data="(v) => actionType = v">
                                    </Selector>
                                </div>
                                <div v-if="actionType != ''">
                                    <div class="pt-3">
                                        <label class="form-check-label">Device</label>
                                        <Selector placeholder="Select" :items="deviceFeatureList" :value="deviceId"
                                            @update:data="(v) => deviceId = v">
                                        </Selector>
                                    </div>
                                    <div class="pt-3 ">
                                        <label class="form-check-label">Device property</label>
                                        <Selector placeholder="Select" :items="getDeviceFeatureNames"
                                            :value="deviceProperty" @update:data="(v) => deviceProperty = v"
                                            :disabled="deviceId == ''">
                                        </Selector>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
                buttons for add, close , clear
                <div class="modal-footer">
                    <button type="button" class="btn btn-primary" :disabled="isValid() == false" @click="add">Add</button>
                    <button type="button" class="btn btn-secondary" @click="clear">Clear</button>
                    <button type="button" class="btn btn-secondary" data-bs-dismiss="modal" @click="close">Close</button>
                </div>
            </div>
        </div>
    </div>
</template>


<!-- <tbody>
    <tr v-for="(automation, index) in automations" :item="automation">
        <th scope="row">{{ index + 1 }}</th>
        <td>
            <RouterLink :to="`/editor/${automation.id}`">{{
                automation.friendlyname
            }}</RouterLink>
        </td>
        <td>
            {{ automation.description }}
        </td> -->