<script setup lang="ts">
import { store } from "../../store/index";
import { computed, ref } from "vue";
import { useRouter } from "vue-router";
import LastSeen from "../device/LastSeen.vue";
import PowerSource from "../device/PowerSource.vue";
import ConnectionType from "../device/ConnectionType.vue";
import RenameDeviceDialog from "../dialogs/RenameDeviceDialog.vue";
import { Device } from "@/types/device";

const props = defineProps({
    id: String,
});

const previousPage = computed(() => {
    const back = useRouter().options.history.state.back;
    if (back != undefined) {
        return back
    }

    return useRouter().push("/");
});

const showDialog = ref(false)
const device = computed(() => {
    return store.getters["devices/find"](props.id);
});

function renameDevice(value: string) {
    store.dispatch("devices/rename", { name: device.value.friendly_name, newName: value });
}

const displayProps = computed(() => {
    const device = store.getters["devices/find"](props.id) as Device;
    if (device == undefined) {
        return [];
    }

    return [

        {
            key: "Id:",
            value: device.id,
        },
        {
            key: "Friendly name:",
            value: device.friendly_name,
        },
        {
            key: "Description:",
            value: device.description,
        },
        {
            key: "Availability:",
            value: device.properties.availability,
        },
        {
            key: "Last seen:",
            type: LastSeen,
            props: {
                timestamp: device.properties.last_seen,
            },
        },
        {
            key: "Power source:",
            type: PowerSource,
            props: {
                power_source: device.power_source,
                value: device.properties.battery,
            },
        },
        {
            key: "Connection Type:",
            type: ConnectionType,
            props: {
                type: device.connection_type,
            },
        },
    ];
});

</script>
<template>
    <div v-if="device != null" class="panel">
        <div className="panel-header d-flex flex-row">
            <div class="align-self-center me-3">
                <RouterLink :to="`${previousPage}`">
                    <i class="fa fa-arrow-left fa-xl" aria-hidden="true"></i>
                </RouterLink>
            </div>
            <div class="h1 align-self-center">
                {{ device.friendly_name }}
            </div>
        </div>

        <div>
            <dl className="row align-self-center" v-for="(prop, idx) in displayProps">
                <dt className="col-12 col-md-5">{{ prop.key }}</dt>
                <dd className="col-12 col-md-7 " v-if="prop.type == undefined">
                    <div title="last update" className="col text-truncate">
                        {{ prop.value }}
                    </div>
                </dd>
                <dd className="col-12 col-md-7" v-else>
                    <component :is="prop.type" v-bind="prop.props"></component>
                </dd>
            </dl>

            <div class="btn-group btn-group-sm" role="group">
                <button class="btn btn-default btn-number" title="Remove device" @click="showDialog = true">
                    <i class="far fa-edit"></i>
                </button>
            </div>
        </div>

        <RenameDeviceDialog :friendlyName="device.friendly_name" :show="showDialog" @update:name="renameDevice"
            @close="e => showDialog = e">
        </RenameDeviceDialog>
    </div>
</template>
