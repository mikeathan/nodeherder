<script setup lang="ts">
import { PropType, ref } from "vue";
import { store } from "../../../store/index";
import { LoggerSettingsType, LoggerSettingsTypePropsType } from "@/types/settings";
import InputBox from '../../input/InputBox.vue';
import Toggle from "@/components/input/Toggle.vue";


const props = defineProps({
    settings: {
        type: Object as PropType<LoggerSettingsType>,
        default: {},
        required: true,
    },
})

const loggerSettings = ref<LoggerSettingsType>(props.settings)

function inputLostFocus(propName: any, propValue: any) {
    save(propName, propValue);
}

function save(propName: LoggerSettingsTypePropsType, propValue: any) {
    if (loggerSettings.value[propName] != propValue) {
        loggerSettings.value[propName] = propValue
        store.dispatch('appconfig/saveLoggerSettings', loggerSettings.value);
    }
}



function enableLogging(enabled: boolean) {
    if (enabled == loggerSettings.value.enableRemoteLogger) {
        return;
    }
    loggerSettings.value.enableRemoteLogger = enabled;
    store.dispatch('appconfig/saveLoggerSettings', loggerSettings.value);
}
</script>

<template>
    <h2>Logger</h2>
    <div class="row border-bottom py-1 w-100 align-items-center" v-for="(interval, key) in loggerSettings" :key="key">
        <dl class="col-12 col-md-3">
            <dt>
                <strong> {{ key }}</strong>
            </dt>
        </dl>
        <div class="col-md-4">
            <Toggle :minimal="false" :value="loggerSettings.enableRemoteLogger" :valueOn="true" :valueoff="false"
                @update="(v: boolean) => enableLogging(v)">
            </Toggle>

        </div>
    </div>
</template>
