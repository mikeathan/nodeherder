<script setup>
import { useStore } from "vuex";
import { computed, ref } from "vue";
import TriggerActionNew from "./TriggerActionNew";
import TriggerCondition from "./TriggerCondition"

const props = defineProps({
    id: String,
});

const auto = ref(null)


const store = useStore();

const name = computed(() => {
    return store.getters["automations/name"](props.id);;
});

const description = computed(() => {
    return store.getters["automations/description"](props.id);;
});

const enabled = computed(() => {
    return store.getters["automations/enabled"](props.id);;
});
const trigger = computed(() => {
    return store.getters["automations/trigger"](props.id);;
});

const conditions = computed(() => {
    return store.getters["automations/conditions"](props.id);;
});
const automation2 = computed(() => {
    //this.oldForm = Object.assign({}, this.form);
    var result = store.getters["automations/find"](props.id);

    return result
});
function onActionChanged(event, properties) {
    if (event.target.value == "") {
        return;
    }
    var property = properties[event.target.value]
    console.log("onActionChanged:", event.target.value, " type: ", property.type, " attributes:", property.attributes);

    // build div = actionDataDiv
}


function addCondition(event) {
    // emit('addCondition', event)
}

function removeCondition(event) {
    //emit('removeCondition', event)
}

function removeAction(event, trigger) {
    trigger.action = null
}

function update() {
    store.dispatch('automations/save', props.id);
}

function onDeleteTriggerClick(event, automationId, triggerId) {
    // disable accordion from expanding
    event.stopImmediatePropagation();
    event.preventDefault();

    // emit delete trigger event
    store.dispatch('ws/emit', {
        event: "deleteAutomationTrigger", message: {
            automationId: automationId,
            triggerId: triggerId
        }
    });
    console.log("deleteTrigger: automationId", automationId, ",triggerId=", triggerId);
}

// ui example
// https://www.home-assistant.io/docs/automation/editor/
</script>

<template>
    <div v-if="automation">
        <div class="container-fluid p-0 h-100">

            <div class="card">
                <div class="card-header">
                    <div class="form-group">
                        <label for="inputId">Id</label>
                        <input type="input" class="form-control" id="inputName" v-model="props.id" disabled />
                    </div>
                    <div class="form-group">
                        <label for="inputFriendlyName">Friendly Name</label>
                        <input type="input" class="form-control" id="inputName" v-model="automation.friendlyName"
                            disabled />
                    </div>
                    <div class="form-group">
                        <label for="inputDescription">Description</label>
                        <input type="input" class="form-control" id="inputDescription" v-model="automation.description" />
                    </div>
                    <div class="form-check">
                        <input type="checkbox" class="form-check-input" id="checkEnabled" v-model="automation.enabled">
                        <label class="form-check-label" for="checkEnabled">Enabled</label>
                    </div>
                </div>
                <div class="card-body">
                    <div class="col-xl-6 col-md-6 col-sm-3">
                        TODO<br>
                        keep a copy of automation so we dont updat store automtically. only on update click <br>
                        dont allow deleteing all actions, as we cant have automation without it <br>
                        fix layout<br>
                        <div v-for="(trigger, index) in  triggers ">
                            <!-- <div v-for="condition in trigger.conditions">
                                <div class="row">
                                    <TriggerCondition :id="props.id" :index="condition.idx" :name="condition.name"
                                        :operator="condition.equality" :key="condition.idx" :data="condition.value"
                                        @remove="removeCondition($event)"
                                        @update:value="newValue => condition.value = newValue"
                                        @update:operator="newValue => condition.operator = newValue">
                                    </TriggerCondition>
                                </div>
                            </div> -->

                            <div class="row" v-if="trigger.action != null">
                                <TriggerActionNew :id="trigger.action.id" :property="trigger.action.property"
                                    :data="trigger.action.data" :delay="trigger.action.delay"
                                    @add="removeAction($event, trigger)"
                                    @update:data="newValue => trigger.action.data = newValue"
                                    @update:delay="newValue => trigger.action.delay = newValue">
                                </TriggerActionNew>
                            </div>
                        </div>
                        <!-- <div class="accordion accordion-flush" id="triggersList">
                            <div v-for="(trigger, index) in  automation.triggers ">

                                <div class="accordion-item">

                                    <h2 class="accordion-header" id="`header${index}`">

                                        <div class="row ">

                                            <div class="col accordion-button collapsed " data-bs-toggle="collapse"
                                                :data-bs-target="`#collapse${index}`" aria-expanded="false"
                                                :aria-controls="`collapse${index}`">

                                                <div class="col-2 col-md-6">
                                                    Trigger #{{ index + 1 }}
                                                </div>
                                                <div class="col-8 col-md-2">
                                                </div>

                                                <div class="col pe-3 text-end ">
                                                    <span class="fa fa-trash-alt fa-lg"
                                                        @click="onDeleteTriggerClick($event, automation.id, index)"
                                                        data-bs-toggle="collapse" data-bs-target>
                                                    </span>

                                                </div>
                                            </div>
                                        </div>
                                    </h2>


                                    <div :id="`collapse${index}`" class="accordion-collapse collapse"
                                        :aria-labelledby="`header${index}`" data-bs-parent="#triggersList">
                                        <div class="accordion-body">
                                            <label>Condition</label>


                                            <div v-for="condition in trigger.conditions">
                                                <div class="row">
                                                    <TriggerCondition :id="props.id" :index="condition.idx"
                                                        :name="condition.name" :operator="condition.equality"
                                                        :key="condition.idx" :data="condition.value"
                                                        @remove="removeCondition($event)"
                                                        @update:value="newValue => condition.value = newValue"
                                                        @update:operator="newValue => condition.operator = newValue">
                                                    </TriggerCondition>
                                                </div>
                                            </div>

                                            <div class="row">
                                                <TriggerActionNew :id="trigger.action.id"
                                                    :property="trigger.action.property" :data="trigger.action.data"
                                                    :delay="trigger.action.delay" @add="removeAction"
                                                    @update:data="newValue => trigger.action.data = newValue"
                                                    @update:delay="newValue => trigger.action.delay = newValue">
                                                </TriggerActionNew>
                                            </div>
                                        </div>
                                    </div>
                                </div>
                            </div>
                        </div> -->
                    </div>
                </div>
            </div>
            <div>
                <button type="button" class="btn btn-primary mt-3" @click="update()">Update</button>
            </div>
        </div>
    </div>
</template>
