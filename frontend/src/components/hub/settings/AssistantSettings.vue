<script setup lang="ts">
  import { computed } from 'vue';
  import { AssistantSettingsType, AssistantSettingsPropsType } from '@/types/settings.type';
  import InputBox from '../../input/InputBox.vue';
  import { useAlerts } from '@/composables/useAlerts';
  import { isValidHttpUrl } from '@/utils/http.utils';
  import { store } from '@/store';

  const assistantSettings = computed(() => {
    return store.getters['hub/assistant']() as AssistantSettingsType;
  });

  const { showError } = useAlerts();

  async function save(propName: AssistantSettingsPropsType, propValue: any) {
    var sanitizedValue = propValue?.trim() || '';

    if (propName === 'url' && sanitizedValue !== '') {
      if (!isValidHttpUrl(sanitizedValue)) {
        showError('Invalid URL format');
        return;
      }
    }

    if (assistantSettings.value[propName] != sanitizedValue) {
      assistantSettings.value[propName] = sanitizedValue;
      await store.dispatch('hub/saveAssistantSettings', assistantSettings.value);
    }
  }
  async function inputLostFocus(propName: any, propValue: any) {
    await save(propName, propValue);
  }
</script>

<template>
  <div class="grid grid-nogutter" v-for="(value, key) in assistantSettings" :key="key">
    <dl class="col-12 md:col-3 text-secondary">
      <dt>
        <strong>{{ key }}</strong>
      </dt>
    </dl>

    <div class="col-12 md:col-3">
      <InputBox class="w-full" :value="value" @lost-focus="(f) => inputLostFocus(key, f)" />
    </div>
  </div>
</template>
