<script setup lang="ts">
import { computed, PropType } from "vue";
import { getDevices } from "@/contracts/device";
import { store } from "@/store/index";
import { Device, Devices, DeviceFilter } from "@/types/device";

const props = defineProps({
    filter: {
        type: Object as PropType<DeviceFilter>,
        default: () => true,
        required: false
    },
    disabled: {
        type: Boolean,
        default: false,
        required: false
    }
});

const emit = defineEmits<{
    (e: 'updated', id: string, friendlyName: string): void,
}>()

const deviceList = computed(() => {
    var devices = store.getters["devices/listAll"]() as Devices;
    if (devices == undefined) {
        return {}
    }
    return getDevices(devices, props.filter)
})

function deviceSelected(event: Event) {
    const id = (event.target as HTMLInputElement).value;
    const device = store.getters["devices/find"](id) as Device;
    if (device == undefined) {
        emit('updated', '', '')
        return;
    }

    emit('updated', device.id, device.friendly_name)
}

</script>
<style scoped>
select.form-select,
input.form-control {
    border: 0;
    outline: 0;
    border-radius: 0%;
    border-bottom: 1px solid white;
    text-align: left;
    background-image: none;
}

.form-floating>.form-control~label::after {
    background-color: transparent;

}

.form-floating>.form-select~label::after {
    background-color: transparent;
}



input.form-control:focus,
select.form-select:focus,
:active {
    box-shadow: none;
}

select.form-select:hover:not([disabled]) {
    background-image: url("data:image/svg+xml;charset=utf-8,%3Csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 16 16%27%3E%3Cpath fill=%27none%27 stroke=%27%23d4d6d9%27 stroke-linecap=%27round%27 stroke-linejoin=%27round%27 stroke-width=%272%27 d=%27m2 5 6 6 6-6%27/%3E%3C/svg%3E");
    box-shadow: none;
}

select.form-select:first-of-type {
    border-bottom: 0px solid white;
}

select.form-select:required:invalid {
    color: gray;
    border-bottom: 1px solid white;
}

.form-floating>.form-control:focus~label,
.form-floating>.form-control:not(:placeholder-shown)~label,
.form-floating>.form-control~label,
.form-floating>.form-select~label {
    opacity: .6;
    transform: scale(.85) translateY(-.7rem) translateX(.15rem);
}

select.form-select,
input.form-select:disabled {
    color: gray;
    background-color: transparent;
}
</style>
<template>
    <div class="form-floating col-sm-5">
        <select required id="deviceSelector" class="form-select form-select-solid" @change="deviceSelected"
            :disabled="props.disabled">
            <option value=""> Select </option>
            <option v-for="(value, key) in deviceList" :value="value" :key="value">
                {{ key }}
            </option>
        </select>
        <label for="deviceSelector" class="form-label">Device to trigger</label>
    </div>
</template>