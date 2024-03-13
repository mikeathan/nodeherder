<script setup lang="ts">

import { computed, watch, ref, PropType, toRef, inject } from "vue";
import TriggerCondition from "./TriggerCondition.vue"
import Action from "./actions/Action.vue";
import Selector from "../input/Selector.vue"
import { store } from "../../store/index";
import { Device } from "@/types/device";
import { AutomationTrigger, AutomationTriggerAction, AutomationTriggerCondition, AutomationTriggerConditions } from "@/types/automation";
import { isValid, EditableTriggerCondition, EditableActionTrigger, AutomationActionTypes, ActionType } from "../../contracts/automations"
import { capitalizeText } from "../../modules/formatters/text.formatter";
import { Emitter } from 'mitt'
import { EventActions, Events, OpenPanelEvent } from "@/types/events.type";

const emitter = inject('emitter') as Emitter<Events>;
const props = defineProps({
    id: { type: String },
    trigger: { type: Object as PropType<AutomationTrigger>, default: {} as AutomationTrigger },
});

// TODO: can be refactor to some automation context
const conditions = ref<AutomationTriggerConditions>({} as AutomationTriggerConditions)
const actions = ref<AutomationTriggerAction[]>([]); // have a list of actions with only 1 item capacity 

const trigger = ref<AutomationTrigger>(props.trigger)

watch(
    () => props.trigger,
    () => {
        if (props.trigger.action.id != "") {
            actions.value.push(props.trigger.action)
        }
        conditions.value = JSON.parse(JSON.stringify(props.trigger.conditions)) as AutomationTriggerConditions;
    }, { immediate: true }
)

const emit = defineEmits(['save', 'delete'])
function save() {
    trigger.value.conditions = conditions.value
    trigger.value.action = actions.value[0];

    emit('save', trigger.value)
}

function remove() {
    emit('delete', trigger.value)
}

function addCondition() {
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


function deleteAction(index: number) {
    // console.log("Trigger - deleteAction", index, " actions before: ", actions.value.length)
    actions.value.splice(index, 1);
    //console.log("Trigger - deleteAction actions after: ", actions.value.length)
}

function SaveAction(index: number, action: AutomationTriggerAction) {
    //console.log("trigger save action ", action, " index ", index);
    actions.value[index] = action;
    trigger.value.action = actions.value[0];
}

function addAction(actionType: ActionType) {
    // actions.value?.push(new EditableActionTrigger(actionType));
    let index = actions.value.length - 1;
    if (index < 0) {
        index = 0
    }
    emitter.emit('openPanel', createActionOpenPanelEvent(new EditableActionTrigger(actionType), index));
}

function createActionOpenPanelEvent(action: AutomationTriggerAction, index: number): OpenPanelEvent {
    const events: EventActions = {
        'delete': (e) => {
            console.log("TRIGGER createactionpanel delete event");
            deleteAction(index)
        },
        'save': (a) => {
            console.log("TRIGGER createactionpanel save event");
            SaveAction(index, a)
        },
    };

    return { name: 'Action', args: { item: action }, events: events }
}
</script>

<template>
    <div class="container-fluid p-0 h-100">
        <!-- TODO:  -->
        <!-- if automation for device exists message user else we overwrite it -->

        <div class="row" v-if="trigger.name == ''">
            <Selector placeholder="Select trigger" :items="exposesList" :value="trigger.name" alignment="left"
                :disabled="trigger.name != ''" @update:data="v => trigger.name = v">
            </Selector>
        </div>
        <div class="row" v-else>

            <!-- Conditions -->
            <table class="table">
                <thead>
                    <tr>
                        <th scope="col">
                            Trigger {{ capitalizeText(trigger.name) }}
                        </th>
                        <th scope="col">#</th>
                    </tr>
                </thead>
                <tr>
                    <th scope="col">
                        <h5>Conditions
                            <button type="button" class="btn btn-default btn-number" @click="(e) => addCondition()">
                                <span class="fa fa-plus"></span>
                            </button>
                        </h5>
                    </th>
                </tr>
                <tbody v-for="( condition, index ) in   conditions  " :item="condition">
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
                    <!-- Actions Header -->
                    <th scope="col">
                        <h5>Actions
                            <span v-if="actions.length == 0">
                                <button type="button" class="btn btn-default btn-number  ms-3" data-bs-toggle="dropdown"
                                    aria-expanded="false">
                                    <span class="fa fa-plus"></span>
                                </button>
                                <ul class="dropdown-menu">
                                    <li v-for="actionType in AutomationActionTypes">
                                        <a @click="addAction(actionType as ActionType)" class="dropdown-item"
                                            data-toggle="dropdown">New {{ actionType }}</a>
                                    </li>
                                </ul>
                            </span>
                        </h5>
                    </th>
                </tr>

                <!-- Actions -->
                <tbody v-for="(action, index) in actions" :item="action">
                    <tr>
                        <th>
                            <Action :item="action">
                                <!-- @delete="deleteAction(index)" @save="a => SaveAction(index, a)" -->
                            </Action>
                        </th>

                    </tr>
                </tbody>
            </table>
        </div>
        <div class="row  pt-3">
            <div class="row pt-3">
                <div class="col">
                    <button type="button" class="btn btn-light" @click="save">
                        Save
                    </button>
                    <button type="button" class="btn btn-light" :disabled="isValid(trigger) == false" @click="remove">
                        Delete
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>