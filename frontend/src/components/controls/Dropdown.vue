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
    type: String as PropType<'small' | 'large' | undefined>,
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
    :model="convert()"
    :disabled="props.disabled"
    text 
    :icon="props.icon" :size="props.size"/>
  
  <!-- <button
    :id="`dropdownControl`"
    type="button"
    :class="`btn ${props.className}`"
    data-bs-toggle="dropdown"
    aria-expanded="false"
    :disabled="props.disabled">
    <slot></slot>
  </button>
  <ul class="dropdown-menu">
    <li v-for="item in props.items" :key="item.value">
      <a
        @click="item.click(item.value)"
        class="dropdown-item"
        data-toggle="dropdown">
        {{ item.name }}</a
      >
    </li>
  </ul> -->
</template>
