<script setup lang="ts">
import { useStore } from "vuex";
import { computed, watch, ref, PropType } from "vue";
import { Condition, ExposeTrigger, ExposeTriggerWrapper, DefaultExposeTriggerWrapper } from "../../models/automations"
import { getDeviceExposeNamesMap } from "../../modules/convert"

import TriggerCondition from "./TriggerCondition.vue"
import TriggerAction from "./TriggerAction.vue";
import Selector from "../input/Selector.vue"

const props = defineProps({
    id: { type: String },
    trigger: { type: Object as PropType<ExposeTrigger> },
});

const store = useStore();
const trigger = ref<ExposeTriggerWrapper>(new DefaultExposeTriggerWrapper())
const emit = defineEmits(['save', 'delete'])

watch(
    () => props.trigger,
    () => {

        let obj: ExposeTrigger = JSON.parse(JSON.stringify(props.trigger))
        trigger.value = new ExposeTriggerWrapper(obj)

    }, { immediate: true }
)

const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

const action = computed(() => {
    return trigger.value.action()
});

function addAction(): void {
    trigger.value.createAction();
}

function removeAction(event: Event): void {
    trigger.value.clearAction()
}

function addCondition(): void {
    trigger.value.addCondition(new Condition())
}

function removeCondition(index: number): void {
    trigger.value.removeCondition(index)
}


function save() {
    emit('save', trigger.value.getTrigger())
}

function remove() {
    emit('delete', trigger.value.getIdx())
}

const exposesList = computed(() => {
    return getDeviceExposeNamesMap(device.value)
})

</script>

<template>
    <div class="container-fluid p-0 h-100">
        <!-- TODO:  -->
        <!-- if automation for device exists message user else we overwrite it -->

        <div class="row" v-if="trigger.name() == ''">
            <Selector placeholder="Select trigger" :items="exposesList" :value="trigger.name()" alignment="left"
                :disabled="trigger.name() != ''" @update:data="v => trigger.setName(v)">
            </Selector>
        </div>
        <div class="row" v-else>


            <!-- Conditions -->

            <table class="table">
                <thead>
                    <tr>
                        <th scope="col">
                            Trigger {{ trigger.displayName() }}
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
                <tbody v-for="(condition, index) in  trigger.getConditions() " :item="condition">
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
                            <button v-if="trigger.action() == null" type="button" class="btn btn-default btn-number ms-3"
                                @click="addAction()">
                                <span class=" fa fa-plus"></span>
                            </button>
                        </h5>
                    </th>
                </tr>
                <tbody>

                    <tr>
                        <th scope="w-25">
                            <TriggerAction v-if="action != null" :id="action.id" :property="action.property"
                                :data="action.data" :delay="action.delay" :operation="action.operation"
                                @update:id="(id, name) => trigger.setActionDeviceId(id, name)"
                                @update:property="v => trigger.setActionProperty(v)"
                                @update:data="v => trigger.action().data = v"
                                @update:delay="v => trigger.action().delay = v"
                                @update:operation="v => trigger.action().operation = v">
                            </TriggerAction>
                        </th>
                        <td>
                            <span v-if="action != null" class="fa fa-trash-alt fa-sm" @click="removeAction">
                            </span>
                        </td>
                    </tr>
                </tbody>
            </table>
        </div>


        <div class="row  pt-3">
            <div class="row pt-3">
                <div class="col">
                    <button type="button" class="btn btn-light" :disabled="trigger.isValid() == false" @click="save">
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