<script setup lang="ts">
import { store } from "../../store/index";
import { computed, ref } from "vue";
import LastSeen from "../device/LastSeen.vue";
import PowerSource from "../device/PowerSource.vue";
import ConnectionType from "../device/ConnectionType.vue";
import RenameDeviceDialog from "../dialogs/RenameDeviceDialog.vue";
import { Device } from "@/types/device";

const props = defineProps({
    id: String,
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
            key: "IEEE Address:",
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
    <div>
        <dl class="grid grid-nogutter" v-for="(prop, idx) in displayProps" :key="idx">
            <dt class="col-12 md:col-5">
                {{ prop.key }}
            </dt>
            <dd class="col-12 md:col-7">
                <div v-if="prop.type === undefined">
                    <span title="last update">{{ prop.value }}</span>
                </div>
                <template v-else>
                    <component :is="prop.type" v-bind="prop.props"></component>
                </template>
            </dd>
        </dl>

        <div class="btn-group btn-group-sm" role="group">
            <button class="btn btn-default btn-number" title="Remove device" @click="showDialog = true">
                <i class="far fa-edit"></i>
            </button>
        </div>
    </div>
    <Button  icon="pi pi-user-edit"  variant="text" v-tooltip="'Rename device'"  @click="showDialog = true" />
    <RenameDeviceDialog :friendlyName="device.friendly_name" :show="showDialog" @update:name="renameDevice" @close="showDialog = false" />
</template>
