<script setup lang="ts">
import { computed, ref, watch } from "vue";
import { useRouter } from 'vue-router'
import InputBox from "../input/InputBox.vue"
import { store } from "../../store/index";
import { Device } from "@/types/device";
import { Automation, AutomationTrigger } from "@/types/automation";
import { EditableAutomationTrigger, DeviceAutomation } from "../../contracts/automations";
import Panel from "../controls/Panel.vue";
import { EventActions, OpenPanelEvent } from "@/types/events.type";
import ButtonPanel from "@/components/controls/ButtonPanel.vue";
import { createSaveDeleteCancelButtonItems } from "../../configs/automation/trigger-dropdown.config";

const emit = defineEmits(['cancel'])

const props = defineProps({
    id: String,
});

const router = useRouter()
const automation = ref<Automation>({} as Automation)
const selectedTrigger = ref<AutomationTrigger>();

const buttonPanelItems = computed(() => {

    const isActionValid = automation.value.triggers.length == 0 &&
        automation.value.triggers.filter(k => k.action != null).length == automation.value.triggers.length;

    return createSaveDeleteCancelButtonItems(
        () => saveAutomation(),
        () => deleteAutomation(),
        () => cancel(),
        isActionValid,
        isActionValid);
});

watch(
    () => props.id,
    () => {
        var sourceAutomation = store.getters["automations/find"](props.id) as Device;
        if (sourceAutomation != undefined) {
            // make a deep copy to make it not reactive
            automation.value = JSON.parse(JSON.stringify(sourceAutomation)) as DeviceAutomation

        } else {

            automation.value = new DeviceAutomation()
            var device = store.getters["devices/find"](props.id) as Device;
            if (device != undefined) {
                automation.value.id = device.id
                automation.value.friendlyname = device.friendly_name
            }
        }
    }, { immediate: true }
)




function createNewTrigger() {
    selectedTrigger.value = EditableAutomationTrigger.create()
}

function cancel() {
    console.log("cancel");
    emit('cancel')
}

function saveAutomation() {
    store.dispatch('automations/save', automation.value as Automation);
    router.push("/viewer")
}

function deleteAutomation() {
    var sourceAutomation = store.getters["automations/find"](props.id);
    if (sourceAutomation != undefined) {
        store.dispatch('automations/delete', automation.value.id);
        // todo; alert message box to ask user
        router.push("/viewer")
    }
}

function deleteTrigger(trigger: AutomationTrigger): void {
    automation.value.triggers = automation.value.triggers.filter((e, i) => e != trigger);
}

function saveTrigger(trigger: AutomationTrigger): void {
    const idx = automation.value.triggers.indexOf(trigger)
    if (idx == -1) {
        automation.value.triggers.push(trigger)
    } else {
        automation.value.triggers[idx] = trigger
    }
}

function getConditionsDescription(trigger: AutomationTrigger): string {
    var conditions = trigger.conditions
    if (conditions.length == 0) {
        return ""
    }
    var condition = conditions[0];
    var description = condition.name + " " + condition.equality + " " + condition.value;
    if (conditions.length > 1) {
        description += "..."
    }

    return description;
}


function getActionDescription(trigger: AutomationTrigger): string {
    if (trigger.action?.id == '') {
        return "<EMPTY>"
    }

    var description = trigger.action.friendlyname + " " +
        trigger.action.property;

    return description;
}

function rowClicked(trigger: AutomationTrigger): void {
    selectedTrigger.value = trigger;
}

function onDeleteTriggerClick(event: Event, trgger: AutomationTrigger): void {
    deleteTrigger(trgger);
}

function resetSelection() {
    selectedTrigger.value = undefined;
}

const panelItem = computed(() => {
    console.log("DEVICEAUTOMATION - createOpenPanelEvent");
    return createOpenPanelEvent()
});


function createOpenPanelEvent(): OpenPanelEvent {

    const events: EventActions = {
        save: (e: AutomationTrigger) => {
            console.log("DEVICEAUTOMATION - SAVE");
            saveTrigger(e)
        },
        delete: (e: AutomationTrigger) => {
            console.log("DEVICEAUTOMATION - DELETE");
            deleteTrigger(e)
        },
    };
    return { name: 'Trigger', args: { id: props.id, trigger: selectedTrigger.value }, events: events }

}
</script>

<style scoped>
.custom-control-input {
    transform: scale(1.4);
}
</style>

<template>
    <div v-if="automation">
        <div class="container-fluid p-0 h-100">
            <div class="card col-xl-5 col-md-6 col-sm-3">
                <div class="card-header ">
                    <div class="pt-3 ">
                        <InputBox label="Id" :disabled="true" :value="automation.id">
                        </InputBox>
                    </div>
                    <!-- <div class="pt-3 ">
                        <label class="form-check-label">Id</label>
                        <DataInput :data="automation.id" alignment="left" :disabled="true">
                        </DataInput>
                    </div> -->
                    <div class="pt-3 ">
                        <InputBox label="Friendly Name" :disabled="true" :value="automation.friendlyname">
                        </InputBox>
                    </div>
                    <div class="pt-3  pb-4">
                        <InputBox label="Description" @updated="(v) => automation.description = v"
                            :value="automation.description">
                        </InputBox>
                    </div>
                    <div class="pb-3">
                        <div class=" form-check form-switch ms-2">
                            <label class="form-check-label ms-3">Enabled</label>
                            TODO create new component
                            <input class="form-check-input custom-control-input" type="checkbox" role="switch"
                                id="flexSwitchCheckDefault" v-model="automation.enabled">
                        </div>
                    </div>
                </div>

                <div class="card-body ">
                    <div v-if="selectedTrigger == null">
                        <ButtonPanel :buttons="buttonPanelItems"></ButtonPanel>

                        <table class="table responsive table-hover ">
                            <thead>
                                <tr>
                                    <th scope="col">#</th>
                                    <th scope="col">Action</th>
                                    <th scope="col">Conditions</th>
                                    <th scope="col">
                                        <button type="button" class="btn btn-default btn-number"
                                            @click="createNewTrigger()">
                                            <span class="fa fa-plus"></span>
                                        </button>
                                    </th>
                                </tr>
                            </thead>
                            <tbody v-for="(trigger, index) in automation.triggers" :item="trigger">
                                <tr>
                                    <th scope="row">
                                        {{ index + 1 }}
                                    </th>
                                    <td @click="rowClicked(trigger)">
                                        {{ getActionDescription(trigger) }}
                                    </td>
                                    <td>
                                        {{ getConditionsDescription(trigger) }}
                                    </td>
                                    <td>
                                        <span class="fa fa-trash-alt fa-sm"
                                            @click="onDeleteTriggerClick($event, trigger)" data-bs-toggle="collapse"
                                            data-bs-target>
                                        </span>
                                    </td>
                                </tr>
                            </tbody>
                        </table>
                    </div>

                    <div v-else class="row">
                        <!-- <button type="button" class="btn-close" aria-label="Close"
                            @click="() => showTriggerCreation = false"></button> -->
                        <!-- <div class="col"> -->

                        <Panel :item="panelItem" @close="resetSelection">

                        </Panel>
                    </div>
                </div>


            </div>
        </div>
    </div>
</template>