<script setup lang="ts">
import { computed, PropType, ref } from "vue";
import { getExposes } from "@/contracts/device";
import { store } from "@/store/index";
import { Device, DeviceFilter } from "@/types/device";

const props = defineProps({
  id: {
    type: String,
    default: "",
    required: true,
  },
  filter: {
    type: Function as PropType<DeviceFilter>,
    default: (expose: any) => true,
    required: false,
  },
  label: {
    type: String,
    default: "",
    required: false,
  },
  disabled: {
    type: Boolean,
    default: false,
    required: false,
  },
});

const emit = defineEmits<{
  (e: "updated", valueid: string): void;
}>();

const selectedExpose = ref<string>('');
const exposeList = computed(() => {
  const device = store.getters["devices/find"](props.id) as Device;
  if (device == undefined) {
    return {};
  }
  return getExposes(device, props.filter);
});

function exposeSelected(event: Event) {
  selectedExpose.value = (event.target as HTMLInputElement).value;
  emit("updated", selectedExpose.value);
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
  opacity: 0.6;
  transform: scale(0.85) translateY(-0.7rem) translateX(0.15rem);
}

select.form-select,
input.form-select:disabled {
  color: gray;
  background-color: transparent;
}
</style>
<template>
  <div v-if="props.label != ''" class="form-floating col-sm-5">
    <select required id="exposeSelector" class="form-select form-select-sm" @change="exposeSelected"
      :disabled="props.disabled">
      <option value="">Select</option>
      <option v-for="value in exposeList" :value="value" :key="value">
        {{ value }}
      </option>
    </select>
    <label for="exposeSelector" class="form-label">{{ props.label }}</label>
  </div>
  <div v-else>
    <select required class="form-select form-select-sm" @change="exposeSelected" :disabled="props.disabled">
      <option value="">Select</option>
      <option v-for="value in exposeList" :value="value" :key="value">
        {{ value }}
      </option>
    </select>
  </div>
</template>
