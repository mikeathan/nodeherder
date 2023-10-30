<script setup>
import { useStore } from "vuex";
import { computed, onMounted, reactive, ref } from "vue";

const props = defineProps({
    id: String,
});

const store = useStore();
const properties = computed(() => {

    var device = store.getters["devices/find"](props.id);
    if (device.properties.feature == false) {
        return []
    }
    return device.properties;
    // get device properties
    // and check for feature = true
    // "properties": {
    //         "feature": true,
    //         "max": 254,
    //         "min": 0,
    //         "type": "numeric"
    //       }
});

function onActionChanged(event, properties) {
    if (event.target.value == "") {
        return;
    }
    var property = properties[event.target.value]
    console.log("onActionChanged:", event.target.value, " type: ", property.type, " attributes:", property.attributes);

    // build div = actionDataDiv
}

</script>
<template>
    <label>Action</label>

    <div class="row w-50">
        <div class="col">
            <input type="text" class="form-control" v-model="trigger.action.friendlyname" placeholder="Action friendly name"
                disabled>
        </div>
        <div class="col">

            <select id="propertySelect" style="text-align:center;" class="form-control"
                :modelValue="trigger.action.property" @change="onActionChanged($event, feature.properties)">

                <option v-for="property in feature.properties" :value="property.name" :key="property.name">

                    {{ property.name }}
                </option>
            </select>
        </div>
        <div class="col" id="actionDataDiv">
            <input type="text" style="text-align:center;" class="form-control" v-model="trigger.action.data"
                placeholder="Action data">
        </div>
        <div class="col">
            <input type="text" style="text-align:center;" class="form-control" v-model="trigger.action.delay"
                placeholder="Action delay">
        </div>


    </div>
</template>
