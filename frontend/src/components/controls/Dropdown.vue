<script setup lang="ts">
  import { PropType } from 'vue';
  import { DropDownItemType } from '../../types/controls.type';

  const props = defineProps({
    items: {
      type: Object as PropType<Array<DropDownItemType>>,
      default: [],
      required: true,
    },
    size: {
      type: String as PropType<
        'small' | 'large' | undefined
      >,
      default: 'small',
    },
    icon: {
      type: String,
      default: '',
    },
    className: {
      type: String,
      default: '',
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    label: {
      type: String,
      default: '',
    },
  });

  function convert() {
    return props.items.map((item) => {
      return {
        label: item.name,
        command: () => item.click(item.value),
      };
    });
  }
</script>
<template>
  <SplitButton
    :label="props.label"
    :model="convert()"
    :disabled="props.disabled"
    text
    :icon="props.icon"
    :size="props.size" />
</template>
