<script setup lang="ts">
import { computed } from 'vue';
import { store } from '../../../store/index';
import { HistorySettingsType, HistorySettingsPropsType } from '@/types/settings.type';
import InputBox from '../../input/InputBox.vue';

const historySettings = computed(() => {
  return store.getters['hub/history']() as HistorySettingsType;
});


function inputLostFocus(propName: any, propValue: any) {
  save(propName, propValue);
}

function save(propName: HistorySettingsPropsType, propValue: any) {
  if (historySettings.value[propName].value != propValue) {
    historySettings.value[propName].value = propValue;
    store.dispatch('hub/saveHistorySettings', historySettings.value);
  }
}
</script>
<template>
  <div class="grid grid-nogutter" v-for="(interval, key) in historySettings" :key="key">
    <dl class="col-12 md:col-3 text-secondary">
      <dt>
        <strong>{{ key }}</strong>
      </dt>
    </dl>

    <div class="col-12 md:col-3">
      <InputBox :label="interval.unit" :value="interval.value" :is-numeric="true"
        @lost-focus="(f) => inputLostFocus(key, f)" />
    </div>
  </div>
</template>
