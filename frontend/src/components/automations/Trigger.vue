<script setup lang="ts">
import { computed, watch, ref, PropType } from "vue";
import TriggerCondition from "./TriggerCondition.vue"
import Selection from "../input/Selection.vue"
import Dropdown from "../controls/Dropdown.vue"
import { store } from "../../store/index";
import { Device } from "@/types/device";
import { AutomationTrigger, AutomationTriggerAction, AutomationTriggerCondition, AutomationTriggerConditions } from "@/types/automation";
import { isValid, EditableTriggerCondition, EditableActionTrigger, AutomationActionTypes, ActionType, getActionType } from "../../contracts/automations"
import { capitalizeText } from "../../modules/formatters/text.formatter";
import { EventActions, OpenPanelEvent } from "@/types/events.type";
import { emitClosePanel, emitOpenPanel } from "@/mixins/useAutomationsEventBus";
import ActionViewer from "./actions/ActionViewer.vue";
import ButtonPanel from "@/components/controls/ButtonPanel.vue";

import { createButtons, createNewActionDropdownItems, createSaveDeleteButtonItems } from "../../configs/automation/trigger-dropdown.config";
const props = defineProps({
    id: { type: String },
    trigger: { type: Object as PropType<AutomationTrigger>, default: {} as AutomationTrigger },
});

// TODO: can be refactor to some automation context
const conditions = ref<AutomationTriggerConditions>({} as AutomationTriggerConditions)
const actions = ref<AutomationTriggerAction[]>([]); // have a list of actions with only 1 item capacity

const trigger = ref<AutomationTrigger>(props.trigger)

const dropDownitems = computed(() => createNewActionDropdownItems((e: ActionType) => addNewAction(e)));


const buttonPanelItems = computed(() => {
    return createButtons([
        {
            name: "Save",
            click: save,
            disabled: isValid(trigger.value) == false
        },
        {
            name: "Delete",
            click: remove,
            disabled: isValid(trigger.value) == false
        },
        // { // It doesnt work - needs investigation
        //     name: "Add Condition",
        //     click: (e: any) => { console.log("add condition event") },
        //     disabled: isValid(trigger.value) == false
        // },
    ])
});

watch(
    () => props.trigger,
    () => {
        if (props.trigger.action != undefined && props.trigger.action.id != "") {
            const action = JSON.parse(JSON.stringify(props.trigger.action)) as AutomationTriggerAction
            actions.value.push(action);
        }
        conditions.value = JSON.parse(JSON.stringify(props.trigger.conditions)) as AutomationTriggerConditions;
    }, { immediate: true }
)

const emit = defineEmits<{
    (e: 'save', trigger: AutomationTrigger): void,
    (e: 'delete', trigger: AutomationTrigger): void,
}>()


function save() {

    trigger.value.conditions = conditions.value
    if (actions.value.length > 0) {
        trigger.value.action = actions.value[0];
    }

    emit('save', trigger.value)
    emitClosePanel('Trigger')
}

function remove() {
    emit('delete', trigger.value)
    emitClosePanel('Trigger')
}

function addNewCondition() {
    conditions.value.push(new EditableTriggerCondition());
}

function removeTriggerCondition(condition: AutomationTriggerCondition) {
    conditions.value = conditions.value.filter((c) => c != condition);
}

const exposesList = computed(() => {
    const device = store.getters["devices/find"](props.id) as Device;
    return Object.assign({}, ...Object
        .values(device.exposes)
        .map((e) => ({ [e.name]: e.name })))
})

function deleteAction() {
    actions.value = [];
}

function SaveAction(action: AutomationTriggerAction) {
    actions.value[0] = action;
    trigger.value.action = actions.value[0];
}

function addNewAction(actionType: ActionType) {
    emitOpenPanel(createActionOpenPanelEvent(new EditableActionTrigger(actionType), true));
}

const actionEvents = (): EventActions => {
    return {
        'delete': (e) => {
            deleteAction()
        },
        'save': (a) => {
            SaveAction(a)
        }
    }
};

function createActionOpenPanelEvent(action: AutomationTriggerAction, editMode: boolean): OpenPanelEvent {
    return { name: 'ActionEditor', args: { automationId: props.id, item: action, editMode: editMode }, events: actionEvents() }
}

</script>

<template>
    <!-- TODO:  -->
    <!-- if automation for device exists message user else we overwrite it -->
    <div class="row pb-3">
        <div class="col">
            <ButtonPanel :buttons="buttonPanelItems">
                <button :id="`dropdownControl`" type="button" class="btn btn-light" @click="addNewCondition"
                    aria-expanded="false">New condition</button>

                <Dropdown :items="dropDownitems" class-name="btn-light" :disabled="actions.length != 0">
                    New Action
                </Dropdown>
            </ButtonPanel>

        </div>
    </div>
    <div class="row" v-if="trigger.name == ''">
        <Selection :value="trigger.name" text="Select trigger" :disabled="trigger.name != ''" size="normal"
            @updated="v => trigger.name = v" :items="exposesList">
        </Selection>
    </div>
    <div class="row" v-else>

        <h5>Trigger for {{ capitalizeText(trigger.name) }}</h5>

        <!-- Conditions -->
        <div class="card " v-if="conditions.length > 0">

            <div class="card-header">
                <h5>WHEN</h5>
            </div>
            <ul class="list-group list-group-flush">
                <li class="list-group-item" v-for="( condition, index ) in conditions  " :item="condition">
                    <div class="row">
                        <div class="col-sm-10">
                            <TriggerCondition :id="props.id" :name="condition.name" :operator="condition.equality"
                                :data="condition.value" @update:name="newValue => condition.name = newValue"
                                @update:value="newValue => condition.value = newValue"
                                @update:operator="newValue => condition.equality = newValue">
                            </TriggerCondition>
                        </div>
                        <div class="col-sm-2">
                            <span class="fa fa-trash-alt fa-sm" @click="removeTriggerCondition(condition)">
                            </span>
                        </div>
                    </div>
                </li>
            </ul>

        </div>
        <!-- ACTIONS -->
        <div class="card" v-if="actions.length > 0">
            <div class="card-header">
                <h5>THEN</h5>
            </div>
            <ul class="list-group list-group-flush">
                <li class="list-group-item" v-for="( action, index ) in actions  " :item="action">
                    <div class="row">
                        <div class="col-sm-10">
                            <ActionViewer :automation-id="props.id" :item="action" :edit-events="actionEvents()"
                                @delete="deleteAction()">
                            </ActionViewer>
                        </div>
                        <div class="col-sm-2">
                            <span class="fa fa-trash-alt fa-sm" @click="deleteAction()">
                            </span>
                        </div>
                    </div>
                </li>
            </ul>
        </div>

        <!-- DELETE -->
        <!-- Conditions -->
        <!-- table-striped -->
        <!-- <table class="table table-responsive borderless table-sm">
            <thead>
                <tr>
                    <th scope="col">
                        Trigger for {{ capitalizeText(trigger.name) }}
                    </th>
                    <th scope="col">#</th>
                </tr>
            </thead>
            <tr>
                <th scope="col" v-if="conditions.length > 0">
                    <h4>WHEN</h4>
                </th>
            </tr>
            <tbody v-for="( condition, index ) in conditions  " :item="condition">
                <tr>
                    <th scope="w-25">
                        <TriggerCondition :id="props.id" :name="condition.name" :operator="condition.equality"
                            :data="condition.value" @update:name="newValue => condition.name = newValue"
                            @update:value="newValue => condition.value = newValue"
                            @update:operator="newValue => condition.equality = newValue">
                        </TriggerCondition>
                    </th>
                    <td>
                        <span class="fa fa-trash-alt fa-sm" @click="removeTriggerCondition(condition)">
                        </span>
                    </td>
                </tr>
            </tbody>
            <tr>
                <th scope="col" v-if="actions.length > 0">
                    <h4>THEN</h4>
                </th>
            </tr>
            <tbody v-for="(action) in actions" :item="action">
                <tr>
                    <th>
                        <ActionViewer :automation-id="props.id" :item="action" :edit-events="actionEvents()"
                            @delete="deleteAction()">
                        </ActionViewer>
                    </th>
                    <td>
                        <span class="fa fa-trash-alt fa-sm" @click="deleteAction()">
                        </span>
                    </td>
                </tr>
            </tbody>
        </table> -->
    </div>
</template>