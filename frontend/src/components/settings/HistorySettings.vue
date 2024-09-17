<script setup lang="ts">
import { PropType, ref } from "vue";
import { store } from "../../store/index";
import { HistorySettingsType, HistorySettingsPropsType } from "@/types/settings";
import InputBox from '../input/InputBox.vue';


const props = defineProps({
    settings: {
        type: Object as PropType<HistorySettingsType>,
        default: {},
        required: true,
    },
})

const historySettings = ref<HistorySettingsType>(props.settings)

function inputLostFocus(propName: any, propValue: any) {
    save(propName, propValue);
}

function save(propName: HistorySettingsPropsType, propValue: any) {
    if (historySettings.value[propName].value != propValue) {
        historySettings.value[propName].value = propValue
        store.dispatch('appconfig/saveHistorySettings', historySettings.value);
    }
}

</script>

<template>
    <h2>History</h2>
    <div class="row border-bottom py-1 w-100 align-items-center" v-for="(interval, key) in historySettings" :key="key">
        <dl class="col-12 col-md-3">
            <dt>
                <strong> {{ key }}</strong>
            </dt>
        </dl>
        <div class="col-md-4">
            <InputBox :label="interval.unit" :value="interval.value" :is-numeric="true"
                @lost-focus="(f) => inputLostFocus(key, f)">
            </InputBox>
        </div>
    </div>
</template>
