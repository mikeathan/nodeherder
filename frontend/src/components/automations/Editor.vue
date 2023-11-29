<script setup>
import { useStore } from "vuex";
import { ref, watch } from "vue";
import TriggerAction from "./TriggerAction";
import TriggerCondition from "./TriggerCondition"
import { DeviceTrigger } from "../../models/automation"
import DataInput from "../input/DataInput.vue"
import { useRouter } from 'vue-router'

const props = defineProps({
    id: String,
});

const store = useStore();
const router = useRouter()
const automation = ref(new DeviceTrigger())
watch(
    () => props.id,
    () => {
        var sourceAutomation = store.getters["automations/find"](props.id);
        if (sourceAutomation != undefined) {
            // make a deep copy to make it on reactive
            automation.value = JSON.parse(JSON.stringify(sourceAutomation))
        }
    }, { immediate: true }
)

function hasTriggers() {
    return automation.value.triggers.length > 0
}


function addCondition(event) {
    // emit('addCondition', event)
}

function removeCondition(event, trigger) {
    var index = trigger.conditions.findIndex(item => item.idx === event);
    if (index != -1) {
        trigger.conditions.splice(index, 1);
    }
}

function removeAction(event, trigger) {

    trigger.action = null
}

function save() {
    store.dispatch('automations/save', automation.value);
    router.push("/viewer")
}

function onDeleteTriggerClick(event, triggerId) {
    // disable accordion from expanding
    event.stopImmediatePropagation();
    event.preventDefault();
    automation.value.triggers.splice(triggerId, 1);
}

// ui example
// https://www.home-assistant.io/docs/automation/editor/
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
                        <label class="form-check-label">Id</label>
                        <DataInput :data="props.id" alignment="left" :disabled="true">
                        </DataInput>
                    </div>
                    <div class="pt-3">
                        <label class="form-check-label">Friendly Name</label>
                        <DataInput :data="automation.friendlyname" alignment="left" type="string" :disabled="true">
                        </DataInput>
                    </div>
                    <div class="pt-3  pb-4">
                        <label class="form-check-label">Description</label>
                        <DataInput :data="automation.description" @update:data="(value) => automation.description = value"
                            type="string" alignment="left">
                        </DataInput>
                    </div>

                    <div class="pb-3">
                        <div class=" form-check form-switch ms-2">
                            <label class="form-check-label ms-3">Enabled</label>

                            <input class="form-check-input custom-control-input" type="checkbox" role="switch"
                                id="flexSwitchCheckDefault" v-model="automation.enabled">
                        </div>
                    </div>

                    <div class="btn-group">
                        <button type="button" class="btn btn-light" @click="save" :disabled="hasTriggers() == false">
                            Save
                        </button>
                        <router-link :to="`/viewer`" tag="span">
                            <button type="button" class="btn btn-light">
                                Cancel
                            </button> </router-link>
                    </div>
                </div>
                <div class="card-body ">
                    <div class="accordion accordion-flush" id="triggersList">
                        <div v-for="(trigger, index) in  automation.triggers ">

                            <div class="accordion-item">

                                <h2 class="accordion-header" id="`header${index}`">

                                    <div class="col accordion-button collapsed " data-bs-toggle="collapse"
                                        :data-bs-target="`#collapse${index}`" aria-expanded="false"
                                        :aria-controls="`collapse${index}`">

                                        <div class="col">
                                            Trigger #{{ index + 1 }} - {{ trigger.name }}
                                        </div>

                                        <div class="col pe-3 text-end ">
                                            <span class="fa fa-trash-alt fa-lg" @click="onDeleteTriggerClick($event, index)"
                                                data-bs-toggle="collapse" data-bs-target>
                                            </span>

                                        </div>
                                    </div>
                                </h2>


                                <div :id="`collapse${index}`" class="accordion-collapse collapse"
                                    :aria-labelledby="`header${index}`" data-bs-parent="#triggersList">
                                    <div class="accordion-body">

                                        <label class="pb-4" v-if="trigger.conditions.length > 0">Condition</label>
                                        <div v-for="(condition, idx) in trigger.conditions">
                                            <div class="row">
                                                <TriggerCondition :id="props.id" :index="condition.idx = idx + 1"
                                                    :name="condition.name" :operator="condition.equality"
                                                    :key="condition.idx" :data="condition.value"
                                                    @remove="removeCondition($event, trigger)"
                                                    @update:value="newValue => condition.value = newValue"
                                                    @update:operator="newValue => condition.operator = newValue">
                                                </TriggerCondition>
                                            </div>
                                        </div>

                                        <div class="row">

                                            <div class=" pb-2 pt-2">
                                                <label class="form-check-label">Action</label>
                                                <DataInput type="string" :data="trigger.action.friendlyname"
                                                    alignment="left" :disabled="true">
                                                </DataInput>
                                            </div>

                                            <TriggerAction :id="trigger.action.id" :property="trigger.action.property"
                                                :data="trigger.action.data" :delay="trigger.action.delay"
                                                :allowRemove="false"
                                                @update:data="newValue => trigger.action.data = newValue"
                                                @update:delay="newValue => trigger.action.delay = newValue">
                                            </TriggerAction>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>

        </div>
    </div>
</template>
