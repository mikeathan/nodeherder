<script setup>
import { useStore } from "vuex";
import { computed } from "vue";



const props = defineProps({
    name: String,
});

const store = useStore();
const automation = computed(() => {
    const item = store.getters["automations/find"](props.name);

    return item
});

function update() {
    store.dispatch('automations/save', props.name);
}

// ui example
// https://www.home-assistant.io/docs/automation/editor/
</script>

<template>
    <div v-if="automation">
        <div class="card">
            <div class="card-header">
                <div class="form-group">
                    <label for="inputName">Name</label>
                    <input type="input" class="form-control" id="inputName" v-model="automation.name" />
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
                <div class="accordion accordion-flush" id="triggersList">
                    <div v-for="(sensorTriggers, name) in automation.sensor_triggers">
                        <div v-for="(trigger, index) in sensorTriggers">

                            <div class="accordion-item">
                                <h2 class="accordion-header" :id="`header${index}`">
                                    <button class="accordion-button" type="button" data-bs-toggle="collapse"
                                        :data-bs-target="`#collapse${index}`" aria-expanded="true"
                                        :aria-controls="`collapse${index}`">
                                        Trigger #{{ index + 1 }}
                                    </button>
                                </h2>
                                <div :id="`collapse${index}`" class="accordion-collapse collapse"
                                    :aria-labelledby="`header${index}`" data-bs-parent="#triggersList">
                                    <div class="accordion-body">
                                        <label>Condition</label>
                                        <div class="row w-25" v-for="(condition) in trigger.conditions">
                                            <div class="col">
                                                <input type="text" class="form-control" v-model="condition.name"
                                                    placeholder="Condition name">
                                            </div>
                                            <div class="col">
                                                <input type="text" style="text-align:center;" class="form-control"
                                                    v-model="condition.equalityoperator" placeholder="Equality operator">
                                            </div>
                                            <div class="col">
                                                <input type="text" style="text-align:center;" class="form-control"
                                                    v-model="condition.value" placeholder="Condition value"
                                                    tooltip="Condition value">
                                            </div>
                                        </div>
                                        <p></p>
                                        <label>Action</label>

                                        <div class="row w-50">
                                            <div class="col">
                                                <input type="text" class="form-control"
                                                    v-model="trigger.action.friendlyname"
                                                    placeholder="Action friendly name">
                                            </div>
                                            <div class="col">
                                                <input type="text" style="text-align:center;" class="form-control"
                                                    v-model="trigger.action.name" placeholder="Action name">
                                            </div>
                                            <div class="col">
                                                <input type="text" style="text-align:center;" class="form-control"
                                                    v-model="trigger.action.value" placeholder="Action value">
                                            </div>
                                            <div class="col">
                                                <input type="text" style="text-align:center;" class="form-control"
                                                    v-model="trigger.action.delay" placeholder="Action delay">
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

        <div style="margin-top: 1%;">
            <button type="button" class="btn btn-primary" @click="update()">Update</button>
        </div>
    </div>
</template>
