<script setup>
import { useStore } from "vuex";
import { computed, onMounted, reactive, ref } from "vue";

const props = defineProps({
    trigger: Object
});

const store = useStore();
const properties = computed(() => {

    var device = store.getters["devices/find"](props.trigger.action.id);
    if (device == undefined) {
        console.log("ERROR: device undefined")

        return []
    }

    var action = props.trigger.action;
    var expose = device.exposes[action.property]
    if (expose == undefined) {
        console.log("ERROR: expose undefined")

        return []
    }
    // console.log("friendlyname:" + device.friendly_name)
    // console.log("id:" + device.id)
    // for (var e in device.exposes) {
    //     console.log("expose:" + e)
    // }

    var properties = expose.properties;
    if (properties.feature == false) {
        console.log("properties.feature false")

        return []
    }
    return properties;
});

function getActionBinaryValue(properties) {
    if (properties.on == props.trigger.action.data) {
        return true
    }

    return false
}

function onActionChanged(event, properties) {

    // use v-model ideally to set value
    props.trigger.action.data = event.target.value
}

function onActionStateChanged(event, properties) {
    if (event.target.checked) {
        props.trigger.action.data = properties.on
    } else {
        props.trigger.action.data = properties.off
    }
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
            <div v-if="properties.type == 'binary'">

                <select id="propertySelect" style="text-align:center;" class="form-control"
                    @change="onActionChanged($event, properties)" :modelValue="trigger.action.data">

                    <option :value="properties.on">
                        {{ properties.on }}
                    </option>
                    <option :value="properties.off">
                        {{ properties.off }}
                    </option>
                </select>

                <!-- 
                    use that in the creator component page
                    <div class="form-check form-switch">

                    <label class="form-check-label" for="flexSwitchCheckDefault">State</label>
                    <input class="form-check-input" type="checkbox" id="flexSwitchCheckDefault"
                        :checked="getActionBinaryValue(properties)" @change="onActionStateChanged($event, properties)">
                </div> -->
            </div>


            <!-- <select id="propertySelect" style="text-align:center;" class="form-control"
                :modelValue="trigger.action.property" @change="onActionChanged($event, properties)">

                <option v-for="(value, key) in properties" :value="value" :key="key">
                    {{ value }}
                </option>
            </select> -->
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
