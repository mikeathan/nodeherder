<script setup>
import { useStore } from "vuex";
import { computed, ref } from "vue";

const props = defineProps({
    id: {
        type: String,
        required: true,
    },
    action: {
        type: Object,
        default: null
    },
});

const property = ref(null);
const data = ref(null);
const delay = ref(null);

const store = useStore();
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

const features = computed(() => {

    var device = store.getters["devices/find"](props.id);

    var list = []
    for (const [key, expose] of Object.entries(device.exposes)) {
        if (expose.properties != undefined) {
            list.push(expose)
        }
    }
    if (list.length > 0) {
        property.value = list[0].name
    }
    return list;
});

const feature = computed(() => {

    if (property.value == null) {
        return []
    }

    var device = store.getters["devices/find"](props.id);
    if (device.exposes[property.value] == undefined) {
        console.log("ERROR: property undefined")

        return []
    }

    console.log("DEBUG", property.value)
    console.log("DEBUG", device.exposes[property.value])
    return device.exposes[property.value];
});

function getActionBinaryValue(properties) {
    // if (properties.on == props.trigger.action.data) {
    //     return true
    // }

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
function add() {

}

function remove() {
}

</script>
<template>
    <div class="col" v-if="props.action == null">
        <input type="text" style="text-align:center;" class="form-control" placeholder="Friendly name"
            onfocus="this.placeholder = ''" onblur="this.placeholder='Friendly name'" v-model="device.friendly_name"
            disabled />
    </div>
    <div class="col">
        <select id="featurePropertySelector" style="text-align:center;" class="form-control" v-model="property">
            <option v-for="feature in features" :value="feature.name" :key="feature.name">
                {{ feature.name }}
            </option>
        </select>
    </div>


    <div class="col" v-if="feature.type == 'binary'">
        <select id="propertySelect" style="text-align:center;" class="form-control" @change="onActionChanged($event)">

            <option v-for="(value, key) in feature.properties" :value="value" :key="key">
                {{ value }}
            </option>
        </select>

    </div>
    <div class="col" v-else>
        <input type="text" style="text-align:center;" class="form-control" placeholder="Value"
            onfocus="this.placeholder = ''" onblur="this.placeholder='Value'" v-model="data" />
    </div>




    <div class="col">
        <input type="text" style="text-align:center;" class="form-control" placeholder="Delay (optional)"
            onfocus="this.placeholder = ''" onblur="this.placeholder='Delay (optional)'" v-model="delay" />
    </div>

    <div class="col-3">
        <div class="btn-group">
            <div v-if="props.action == null">
                <button type="button" class="btn btn-default btn-number" @click="add($event)">
                    <span class="fa fa-plus"></span>
                </button>
            </div>
            <div v-else>
                <button type="button" class="btn btn-default btn-number" @click="remove($event)">
                    <span class="fa fa-minus"></span>
                </button>
            </div>
        </div>
    </div>
</template>
