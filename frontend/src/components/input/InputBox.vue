<script setup lang="ts">
import { PropType, computed, ref, watch } from "vue";
import { store } from "@/store/index";
import { ExposeTypes } from "@/types/device.type";
import { Expose, ExposeType } from "@/types/device";
import { DataInputType, DataInputTypes } from "@/types/controls.type";

const props = defineProps({
    dataType: {
        type: Object as PropType<DataInputType>,
        default: DataInputTypes.Text,
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
    (e: "updated", value: any): void;
}>();

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

</script>< template><div v -if="dataType == ExposeTypes.Numeric"><div v -if="props.label != ''" class="form-floating col-sm-3"><input type="text" class="form-control" id="dataInput" v - model="inputValue"@input="inputChanged"

@keypress="isNumber" : disabled="props.disabled" /><label for="dataInput"> {
        {
        props.label
    }
}

</label>< /div>< div v -else class="col-sm-3"><input type="text" class="form-control" id="dataInput"@input="inputChanged" : disabled="props.disabled" /></div>< /div>< div v -if="showPresets" class="form-floating col-sm-5"><select required id="presetsSelector" class="form-select form-select-sm" v - model="selectedPreset"

@change="presetSelected"><option value="">Select < /option>< option v -for="(value, key) in exposePresets" : value="value" : key="key"> {
        {
        key
    }
}

</option>< /select>< label for="presetsSelector" class="form-label">Expose presets < /label>< /div>< /template>
