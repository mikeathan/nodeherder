<script setup>
import { useStore } from "vuex";
import { computed, onMounted, reactive, ref } from "vue";

const props = defineProps({
    id: String,
});

const operators = ref([
    { text: '=', value: '=' },
    { text: '<=', value: '<=' },
    { text: '>=', value: '>=' },
    { text: '>', value: '>' },
    { text: '<', value: '<' }
])

const store = useStore();
const automation = computed(() => {
    return store.getters["automations/find"](props.id);
});

const features = computed(() => {
    if (!store.getters["features/isInitialized"]) {
        store.dispatch('ws/emit', { event: "loadBridgeFeatures" });
    }
    return store.getters["features/items"]
});

function onActionChanged(event, properties) {
    if (event.target.value == "") {
        return;
    }
    var property = properties[event.target.value]
    console.log("onActionChanged:", event.target.value, " type: ", property.type, " attributes:", property.attributes);

    // build div = actionDataDiv
}
onMounted(() => {
    // if (!store.getters["features/isInitialized"]) {
    //     store.dispatch('ws/emit', { event: "loadBridgeFeatures" });
    // }
});

function update() {
    store.dispatch('automations/save', props.id);
}

function onButtonClick(event) {
    console.log("button click");
    event.stopImmediatePropagation();
    event.preventDefault();
}

// ui example
// https://www.home-assistant.io/docs/automation/editor/
</script>

<template>
    <div v-if="automation">
        <div class="card">
            <div class="card-header">
                <div class="form-group">
                    <label for="inputId">Id</label>
                    <input type="input" class="form-control" id="inputName" v-model="automation.id" disabled />
                </div>
                <div class="form-group">
                    <label for="inputFriendlyName">Friendly Name</label>
                    <input type="input" class="form-control" id="inputName" v-model="automation.friendlyName" disabled />
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
                    <div v-for="(trigger, index) in  automation.triggers ">

                        <div class="accordion-item">

                            <h2 class="accordion-header" id="`header${index}`">

                                <div class="container border">
                                    <div class="row ">

                                        <div class="col  border accordion-button collapsed " data-bs-toggle="collapse"
                                            :data-bs-target="`#collapse${index}`" aria-expanded="false"
                                            :aria-controls="`collapse${index}`">

                                            <div class="col-2 border">
                                                <span class="fa fa-trash-alt fa-lg" @click="onButtonClick($event)"
                                                    data-bs-toggle="collapse" data-bs-target>
                                                </span>
                                            </div>

                                            <div class="col-7 border">
                                                Trigger #{{ index + 1 }}
                                            </div>
                                        </div>
                                    </div>
                                </div>
                                <!-- <div class="accordion-button collapsed col-sm" data-bs-toggle="collapse"
                                            :data-bs-target="`#collapse${index}`" aria-expanded="false"
                                            :aria-controls="`collapse${index}`">
                                           
                                            <div id="triggerHeader" class="container">
                                                <div class="row justify-content-md-center">
                                                    <div class="col-sm">
                                                    <div>Trigger #{{ index + 1 }}</div>                                                
                                                    </div>
                                                    <div class="col-sm">
                                                    <span class="fa fa-trash-alt fa-lg" @click="onButtonClick($event)" 
                                                        data-bs-toggle="collapse" data-bs-target>
                                                    </span>
                                                </div>
                                            </div>
                                            </div>
                                        </div>
                                        -->


                            </h2>

                            <!-- <h2 class="accordion-header" :id="`header${index}`">
                                <button class="accordion-button collapsed" type="button" data-bs-toggle="collapse"
                                    :data-bs-target="`#collapse${index}`" aria-expanded="false"
                                    :aria-controls="`collapse${index}`">
                                    Trigger #{{ index + 1 }}
                                </button>
                            </h2> -->
                            <div :id="`collapse${index}`" class="accordion-collapse collapse"
                                :aria-labelledby="`header${index}`" data-bs-parent="#triggersList">
                                <div class="accordion-body">
                                    <label>Condition</label>
                                    <div class="row w-25" v-for="(condition) in trigger.conditions">
                                        <div class="col">
                                            <input type="text" class="form-control" v-model="condition.name"
                                                placeholder="Condition name" disabled>
                                        </div>
                                        <div class="col">
                                            <select id="selectOperators" style="text-align:center;" class="form-control"
                                                v-model="condition.equality">
                                                <option v-for="operator in operators" :value="operator.value"
                                                    :key="operator.value">
                                                    {{ operator.text }}
                                                </option>
                                            </select>
                                        </div>
                                        <div class="col">
                                            <input type="text" style="text-align:center;" class="form-control"
                                                v-model="condition.value" placeholder="Condition value"
                                                tooltip="Condition value">
                                        </div>
                                    </div>
                                    <p></p>
                                    <div v-for="feature in features">
                                        <div v-if="feature.id == trigger.action.id">

                                            <label>Action</label>

                                            <div class="row w-50">
                                                <div class="col">
                                                    <input type="text" class="form-control"
                                                        v-model="trigger.action.friendlyname"
                                                        placeholder="Action friendly name" disabled>
                                                </div>
                                                <div class="col">

                                                    <select id="propertySelect" style="text-align:center;"
                                                        class="form-control" :modelValue="trigger.action.property"
                                                        @change="onActionChanged($event, feature.properties)">

                                                        <option v-for="property in feature.properties"
                                                            :value="property.name" :key="property.name">

                                                            {{ property.name }}
                                                        </option>
                                                    </select>
                                                </div>
                                                <div class="col" id="actionDataDiv">
                                                    <input type="text" style="text-align:center;" class="form-control"
                                                        v-model="trigger.action.data" placeholder="Action data">
                                                </div>
                                                <div class="col">
                                                    <input type="text" style="text-align:center;" class="form-control"
                                                        v-model="trigger.action.delay" placeholder="Action delay">
                                                </div>


                                            </div>
                                        </div>
                                    </div>
                                    <div id=deleteTriggerDiv>

                                        <button type="submit" class="btn btn-primary mt-3">Delete</button>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </div>
            </div>
        </div>

        <div>
            <button type="button" class="btn btn-primary mt-3" @click="update()">Update</button>
        </div>
    </div>
</template>
