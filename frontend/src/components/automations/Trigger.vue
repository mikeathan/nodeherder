<script setup>
import { useStore } from "vuex";
import { computed, watch, ref } from "vue";

import { Operators } from "../../models/automation"
import TriggerCondition from "./TriggerCondition"
import TriggerAction from "./TriggerAction.vue";
import Selector from "../input/Selector.vue"


const props = defineProps({
    id: String,
    trigger: Object,
});

const store = useStore();
const selectedAction = ref(null)
const selectedExpose = ref("")
const trigger = ref(null)


const emit = defineEmits(['addAction', 'removeAction', 'addCondition', 'removeCondition'])


watch(
    () => props.trigger,
    () => {
        selectedAction.value = ""
        selectedExpose.value = props.trigger.name
        // select action optionsto current action if not null
        if (props.trigger.action != null) {
            selectedAction.value = props.trigger.action.id
        }
        trigger.value = props.trigger
    }, { immediate: true }
)


const featureDevices = computed(() => {
    var devices = store.getters["devices/items"];
    // find exposes with properties
    // let all = items.filter(item=> item.age==='18')
    //     return deviceimport DeviceAutomation from "./DeviceAutomation"

    // });
    var list = []
    for (const [key, device] of Object.entries(devices)) {
        for (const [key, expose] of Object.entries(device.exposes)) {
            if (expose.properties != undefined) {
                list.push(device)
                break;
            }
        }
    }

    return list;
});

function deviceList() {
    // todo:
    //var result = Object.keys(obj).map((key) => [key, obj[key]]);
    var list = {}
    for (const [key, device] of Object.entries(featureDevices.value)) {
        list[device.friendly_name] = device.id
    }
    return list
}

function addAction(event) {
    emit('addAction', event)
}

function removeAction(event) {
    emit('removeAction', event)
}


function addCondition(event) {
    emit('addCondition', event)
}

function removeCondition(event) {
    emit('removeCondition', event)
}

</script>

<template>
    <div class="container-fluid p-0 h-100">
        <!-- TODO:  -->
        <!-- mobile dimensions are wrong -->
        <!-- fix triggeractonnew - check refactoring logic -->
        <!-- fix triggeractonnew - props dont update unless we remove and add again -->
        <!-- Save button shouls navigate to automation viewer -->
        <!-- check url design above for styling of creator text input -->
        <!-- if automation for device exists message user else we overwrite it -->

        <br>
        <div>
            <!-- Actions -->
            <h5>Action</h5>
            <div class="mb-3">

                <Selector placeholder="Select device" :items="deviceList()" :value="selectedAction" key="id"
                    alignment="left" @update:data="e => selectedAction = e" :disabled="selectedExpose == ''"></Selector>

            </div>

            <!-- existing action -->
            <div v-if="trigger.action != null" class="row">
                <TriggerAction :id="trigger.action.id" :property="trigger.action.property" :data="trigger.action.data"
                    :delay="trigger.action.delay" @add="removeAction"
                    @update:data="newValue => trigger.action.data = newValue"
                    @update:delay="newValue => trigger.action.delay = newValue">
                </TriggerAction>
            </div>
            <!-- new action -->
            <div v-else-if="selectedAction != ''" class="row">
                <TriggerAction :id="selectedAction" @add="addAction"></TriggerAction>
            </div>

            <!-- Conditions -->
            <div class="accordion accordion-flush mt-3 " id="triggerSelections">
                <div class="accordion-item">
                    <h2 class="accordion-header" id="`header`">

                        <div class="accordion-button collapsed " data-bs-toggle="collapse" :data-bs-target="`#collapseOne`"
                            aria-expanded="false" :aria-controls="`collapseOne`">

                            Conditions
                        </div>
                    </h2>

                    <div id="collapseOne" class="accordion-collapse collapse" aria-labelledby="headingOne"
                        data-bs-parent="#triggerSelections">
                        <div class="accordion-body">
                            <div class="row">
                                <TriggerCondition :id="props.id" :index="0" :name="selectedExpose"
                                    :operator="Operators[0].value" :data="''" @add="addCondition">
                                </TriggerCondition>
                            </div>

                            <div v-for="condition in trigger.conditions">
                                <div class="row">
                                    <TriggerCondition :id="props.id" :index="condition.idx" :name="condition.name"
                                        :operator="condition.equality" :key="condition.idx" :data="condition.value"
                                        @remove="removeCondition($event)"
                                        @update:value="newValue => condition.value = newValue"
                                        @update:operator="newValue => condition.operator = newValue">
                                    </TriggerCondition>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>
    </div>
</template>