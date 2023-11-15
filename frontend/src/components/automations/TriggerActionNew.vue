<script setup>
import { useStore } from "vuex";
import { computed } from "vue";

const props = defineProps({
    id: {
        type: String,
        required: true,
    },
    property: String,
    value: {
        type: Object,
    },
    delay: {
        type: String,
        default: null,
    }
});

const store = useStore();
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

const feature = computed(() => {

    if (props.property == null) {
        return []
    }

    var device = store.getters["devices/find"](props.id);
    if (device == undefined) {
        console.log("ERROR: device undefined")
        return []
    }


    if (device.exposes[props.property] == undefined) {
        console.log("ERROR: property undefined")

        return []
    }

    console.log("DEBUG", props.property)
    console.log("DEBUG", device.exposes[props.property])
    return device.exposes[props.property];
});

function getActionBinaryValue(properties) {
    if (properties.on == props.trigger.action.data) {
        return true
    }

    return false
}

function onActionChanged(event) {

    // use v-model ideally to set value
    // props.trigger.action.data = event.target.value
}

function onActionStateChanged(event, properties) {
    // if (event.target.checked) {
    //     props.trigger.action.data = properties.on
    // } else {
    //     props.trigger.action.data = properties.off
    // }
}


</script>
<template>
    <div class="row w-50">

        <!-- we need  to list state -->
        <div class="col">
            <input type="text" class="form-control" v-model="device.friendly_name" placeholder="Action friendly name"
                disabled>
        </div>
        <div class="col">
            <div v-if="feature.type == 'binary'">
                <select id="propertySelect" style="text-align:center;" class="form-control" v-model="props.value"
                    @change="onActionChanged($event)">

                    <option v-for="(value, key) in feature.properties" :value="value" :key="key">
                        {{ value }}
                    </option>
                </select>

            </div>
            <div v-else>
                <div class="col" id="actionDataDiv">
                    <input type="text" style="text-align:center;" class="form-control" v-model="props.value"
                        placeholder="Action data">
                </div>
            </div>

        </div>

        <div class="col">
            <input type="text" style="text-align:center;" class="form-control" v-model="props.delay"
                placeholder="Action delay">
        </div>
    </div>
</template>
