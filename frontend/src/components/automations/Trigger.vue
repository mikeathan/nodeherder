<script setup>
import { useStore } from "vuex";
import { computed, watch, ref } from "vue";

import { Operators } from "../../models/automation"
import TriggerCondition from "./TriggerCondition"
import TriggerActionNew from "./TriggerActionNew.vue";

const props = defineProps({
    id: String,
    trigger: Object,
});

const store = useStore();
const selectedAction = ref(null)
const selectedExpose = ref("")
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

const emit = defineEmits(['addAction', 'removeAction', 'addCondition', 'removeCondition'])

function getExposes() {
    return Object.keys(device.value.exposes)
}

watch(
    () => props.trigger,
    (newTrigger) => {
        selectedAction.value = null
        selectedExpose.value = props.trigger.name
        // select action optionsto current action if not null
        if (props.trigger.action != null) {
            selectedAction.value = props.trigger.action.id
        }
    }, { immediate: true }
)


const featureDevices = computed(() => {
    var devices = store.getters["devices/items"];
    // find exposes with properties
    // let all = items.filter(item=> item.age==='18')
    //     return devices;
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

function addAction(event) {
    emit('addAction', event)
}

function addCondition(event) {
    emit('addCondition', event)
}

function removeCondition(event) {
    emit('removeCondition', event)
}
function removeAction(event) {
    emit('removeAction', event)
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
                <select id="featureDeviceSelector" style="text-align:center;" class="form-control" v-model="selectedAction"
                    @change="featureSelectionChanged" :disabled="selectedExpose == ''">
                    <option :value="null">Select device</option>
                    <option v-for="device in featureDevices" :value="device.id" :key="device.id">
                        {{ device.friendly_name }}
                    </option>
                </select>
            </div>

            <!-- existing action -->
            <div v-if="trigger.action != null" class="row">
                <TriggerActionNew :id="trigger.action.id" :property="trigger.action.property" :data="trigger.action.data"
                    :delay="trigger.action.delay" @add="removeAction"
                    @update:data="newValue => trigger.action.data = newValue"
                    @update:delay="newValue => trigger.action.delay = newValue">
                </TriggerActionNew>
            </div>
            <!-- new action -->
            <div v-else-if="selectedAction != null" class="row">
                <TriggerActionNew :id="selectedAction" @add="addAction"></TriggerActionNew>
            </div>

            <!-- Conditions -->
            <div class="accordion accordion-flush mt-3" id="triggerSelections">
                <div class="accordion-item">
                    <h2 class="accordion-header" id="`header`">

                        <div class="row ">
                            <div class="col accordion-button collapsed " data-bs-toggle="collapse"
                                :data-bs-target="`#collapseOne`" aria-expanded="false" :aria-controls="`collapseOne`">

                                <div class="col-xs-3 col-md-6">
                                    Conditions
                                </div>
                                <div class="col-xs-10 col-md-2">
                                </div>
                                <!-- @click="onDeleteTriggerClick($event, automation.id, index)" -->
                                <div class="col pe-3 text-end ">
                                    <span class="fa fa-trash-alt fa-lg" data-bs-toggle="collapse" data-bs-target>
                                    </span>

                                </div>
                            </div>
                        </div>
                    </h2>

                    <div id="collapseOne" class="accordion-collapse collapse" aria-labelledby="headingOne"
                        data-bs-parent="#triggerSelections">
                        <div class="accordion-body">
                            <div class="row">
                                <TriggerCondition :id="0" :exposes="getExposes()" :name="selectedExpose"
                                    :operator="Operators[0].value" :data="''" @add="addCondition">
                                </TriggerCondition>
                            </div>

                            <div v-for="condition in trigger.conditions">
                                <div class="row">
                                    <TriggerCondition :id="condition.idx" :name="condition.name"
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