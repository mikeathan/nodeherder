<script setup lang="ts">
  import { ref, watch } from 'vue';

  import { InputNumberBlurEvent, InputNumberInputEvent } from 'primevue';

  const props = defineProps({
    value: null,
    isNumeric: {
      type: Boolean,
      default: false,
      required: false,
    },
    label: {
      type: String,
      default: '',
      required: false,
    },

    disabled: {
      type: Boolean,
      default: false,
      required: false,
    },
    class: {
      type: String,
      default: '',
      required: false,
    },
  });

  const emit = defineEmits<{
    (e: 'updated', value: any): void;
    (e: 'lostFocus', value: any): void;
  }>();

  const inputValue = ref<any>(props.value);
  watch(
    () => props.value,
    () => {
      if (props.isNumeric && (typeof props.value !== 'number' || isNaN(props.value))) {
        inputValue.value = 0;
      } else {
        inputValue.value = props.value;
      }
    },
    { immediate: true }
  );

  function inputNumberEvent(event: InputNumberInputEvent): void {
    const newValue = event.value;
    if (newValue == null || typeof newValue === 'string' || isNaN(newValue)) {
      inputValue.value = 0;
    } else {
      inputValue.value = newValue;
    }
    emit('updated', inputValue.value);
  }

  function inputTextEvent(event: Event): void {
    let value: any = (event.target as HTMLInputElement).value;
    inputValue.value = value;
    emit('updated', inputValue.value);
  }

  function onLostFocus(event: InputNumberBlurEvent): void {
    emit('lostFocus', inputValue.value);
  }
</script>

<template>
  <div v-if="props.isNumeric">
    <FloatLabel variant="in">
      <InputNumber
        v-if="props.isNumeric"
        v-model="inputValue"
        :disabled="props.disabled"
        inputId="integeronly"
        @input="inputNumberEvent"
        @blur="onLostFocus"
        :min="0"
        class="w-full" />
      <label v-if="props.label != ''">{{ props.label }}</label>
    </FloatLabel>
  </div>
  <div v-else>
    <FloatLabel variant="in">
      <InputText
        v-model="inputValue"
        variant="outlined"
        :disabled="props.disabled"
        @input="inputTextEvent"
        :class="props.class"
        @blur="onLostFocus" />
      <label v-if="props.label != ''">{{ props.label }}</label>
    </FloatLabel>
  </div>
</template>
