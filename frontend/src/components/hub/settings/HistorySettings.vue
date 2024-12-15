<script setup lang="ts">
import { PropType, ref } from "vue";
import { store } from "../../../store/index";
import { HistorySettingsType, HistorySettingsPropsType } from "@/types/settings";
import InputBox from '../../input/InputBox.vue';


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
    <div class="grid col-12 align-items-center grid-nogutter" v-for="(interval, key) in historySettings" :key="key">
        <dl class="col md:col-3 sm:col-3">
            <dt>
                <strong> {{ key }}</strong>
            </dt>
        </dl>
        <div class="col md:col-3 sm:col-3">
            <InputBox :label="interval.unit" :value="interval.value" :is-numeric="true"
                @lost-focus="(f) => inputLostFocus(key, f)">
            </InputBox>
        </div>
    </div>
</template>
