<script setup lang="ts">
import { useStore } from "vuex";
import { computed, watch, ref, PropType } from "vue";
import { Condition, ActionTrigger, ExposeTrigger } from "../../models/automations"
import { getDeviceExposeNamesMap } from "../../modules/convert"

import TriggerCondition from "./TriggerCondition.vue"
import TriggerAction from "./TriggerAction.vue";
import Selector from "../input/Selector.vue"

const props = defineProps({
    id: { type: String },
    trigger: { type: Object as PropType<ExposeTrigger> },
});

const store = useStore();
const trigger = ref<ExposeTrigger>(new ExposeTrigger(''))
const emit = defineEmits(['save', 'delete'])

watch(
    () => props.trigger,
    () => {

        let obj: ExposeTrigger = JSON.parse(JSON.stringify(props.trigger))
        trigger.value = obj
        trigger.value.conditions.forEach(function callback(condition, index) {
            condition.idx = index + 1
        });

    }, { immediate: true }
)

const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

function addAction(): void {
    todo
    // make so we can set or clear it . so we dont have to do null checks
    trigger.value.action = new ActionTrigger();
}

function removeAction(event: Event): void {
    trigger.value.action = null;
}

function addCondition(): void {
    trigger.value.conditions.push(new Condition())
}

function removeCondition(index: number): void {
    console.log("index to remove", index)
    trigger.value.conditions = trigger.value.conditions.filter(k => k.idx != index);
}

function isSaveEnabled(): boolean {
    return trigger.value.name != "" && trigger.value.action != null;
}

function actionIdUpdated(id: string, friendlyName: string): void {
    if (trigger.value.action) {
        trigger.value.action.id = id
        trigger.value.action.friendlyname = friendlyName
    }
}

function updateActionProperty(value: string): void {
    if (trigger.value.action != null) {
        trigger.value.action.property = value
    }
}
function updateActionData(value: string): void {
    if (trigger.value.action != null) {
        trigger.value.action.data = value
    }
}
function updateActionDelay(value: number | null): void {
    if (trigger.value.action != null) {
        trigger.value.action.delay = value
    }
}
function updateActionStep(value: number): void {
    if (trigger.value.action != null) {
        trigger.value.action.step = value
    }
}

function save() {
    emit('save', trigger.value)
}

function remove() {
    emit('delete', trigger.value?.idx)
}

const exposesList = computed(() => {
    return getDeviceExposeNamesMap(device.value)
})


</script>

<template>
    <div class="container-fluid p-0 h-100">
        <!-- TODO:  -->
        <!-- if automation for device exists message user else we overwrite it -->

        <div class="row" v-if="trigger.name == ''">

            <Selector placeholder="Select trigger" :items="exposesList" :value="trigger.name" alignment="left"
                :disabled="props.trigger?.name != ''" @update:data="v => trigger.name = v">
            </Selector>
        </div>
        <div class="row" v-else>


            <!-- Conditions -->

            <table class="table">
                <thead>
                    <tr>
                        <th scope="col">
                            Trigger {{ capitalize(trigger.name) }}
                        </th>
                        <th scope="col">#</th>
                    </tr>
                </thead>
                <tr>
                    <th scope="col">
                        <h5>Conditions
                            <button type="button" class="btn btn-default btn-number" @click="addCondition()">
                                <span class="fa fa-plus"></span>
                            </button>
                        </h5>
                    </th>
                </tr>
                <tbody v-for="(condition, index) in  trigger.conditions " :item="condition">
                    <tr>
                        <th scope="w-25">
                            <TriggerCondition :id="props.id" :index="condition.idx" :name="condition.name"
                                :operator="condition.equality" :key="condition.idx" :data="condition.value"
                                @update:name="newValue => condition.name = newValue"
                                @update:value="newValue => condition.value = newValue"
                                @update:operator="newValue => condition.equality = newValue">
                            </TriggerCondition>
                        </th>
                        <td>
                            <span class="fa fa-trash-alt fa-sm" @click="removeCondition(condition.idx)">
                            </span>
                        </td>
                    </tr>
                </tbody>
                <tr>
                    <th scope="col">
                        <h5>Actions
                            <button v-if="trigger.action == null" type="button" class="btn btn-default btn-number ms-3"
                                @click="addAction()">
                                <span class=" fa fa-plus"></span>
                            </button>
                        </h5>
                    </th>
                </tr>
                <tbody>

                    <tr>
                        <th scope="w-25">
                            <TriggerAction v-if="trigger.action != null" :id="trigger.action.id"
                                :property="trigger.action.property" :data="trigger.action.data"
                                :delay="trigger.action.delay" :step="trigger.action.step" @update:id="actionIdUpdated"
                                @update:property="updateActionProperty" @update:data="updateActionData"
                                @update:delay="updateActionDelay" @update:step="updateActionStep">
                            </TriggerAction>
                        </th>
                        <td>
                            <span v-if="trigger.action != null" class="fa fa-trash-alt fa-sm" @click="removeAction">
                            </span>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>


        <!-- Actions -->
        <div class="row  pt-3">
            <!-- <table class="table">
                    <thead>
                        <tr>
                            <th scope="col">
                                <h5>Actions
                                    <button v-if="trigger.action == null" type="button" class="btn btn-default btn-number"
                                        @click="addAction()">
                                        <span class=" fa fa-plus"></span>
                                    </button>
                                </h5>
                            </th>
                            <th scope="col">#</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr>
                            <th scope="w-25">
                                <TriggerAction v-if="trigger.action != null" :id="trigger.action.id"
                                    :property="trigger.action.property" :data="trigger.action.data"
                                    :delay="trigger.action.delay" :step="trigger.action.step" @update:id="actionIdUpdated"
                                    @update:property="newValue => trigger.action.property = newValue"
                                    @update:data="newValue => trigger.action.data = newValue"
                                    @update:delay="newValue => trigger.action.delay = newValue"
                                    @update:step="newValue => trigger.action.step = newValue">
                                </TriggerAction>
                            </th>
                            <td>
                                <span v-if="trigger.action != null" class="fa fa-trash-alt fa-sm" @click="removeAction">
                                </span>
                            </td>
                        </tr>
                    </tbody>
                </table> -->


            <div class="row pt-3">
                <div class="col">
                    <button type="button" class="btn btn-light" :disabled="isSaveEnabled() == false" @click="save">
                        Save
                    </button>
                    <button type="button" class="btn btn-light" @click="remove">
                        Delete
                    </button>
                </div>
            </div>
        </div>
    </div>
</template>