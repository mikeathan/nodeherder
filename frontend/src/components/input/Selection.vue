<script setup lang="ts">
import { KeyValuePair } from '@/types/types';
import { PropType, VNode, computed, h, ref, watch } from 'vue';
import {
  SelectSize,
  SelectFormSize,
  SelectionItems,
  LayoutPosition,
  LayoutPositions,
} from '@/types/controls.type';

const props = defineProps({
  items: {
    type: Object as PropType<SelectionItems>,
    default: [],
    required: true,
  },
  value: null,
  text: {
    type: String,
    default: '',
    required: false,
  },
  label: {
    type: String,
    default: '',
    required: false,
  },
  position: {
    type: String as PropType<LayoutPosition>,
    default: LayoutPositions.left,
    required: false,
  },
  size: {
    type: String as PropType<SelectSize>,
    default: SelectFormSize.small,
    required: false,
  },
  disabled: {
    type: Boolean,
    default: false,
    required: false,
  },
});

const emit = defineEmits<{
  (e: 'updated', value: any): void;
}>();

const selectedValue = ref<any>(props.value);
// watch(
//     () => props.items,
//     () => {

//         //inputValue.value = props.value;
//     }, { immediate: true }
// )
// <option v-for="value in items" :value="value" :key="value">
//                 {{ value }}
//             </option>


// const items = computed(() => {
//   if (Array.isArray(props.items)) {
//     return props.items;
//   }
//   return Object.entries(items).map(([key, value]) => ({
//     key,
//     value
//   }))
// });
function createSelection(): VNode {
  if (Array.isArray(props.items)) {
    return ArraySelection(props.items);
  }
  return KeyValuePairSelection(props.items);
}

const selectClass = (): string => {
  return `form-select ${SelectFormSize[props.size]}`;
};

const selectStyle = (): string => {
  return `text-align:${props.position}`;
};
const defaultText = (): string => {
  return props.text != '' ? props.text : 'Select';
};

function ArraySelection(items: Array<string>): VNode {
  return h(
    'select',
    {
      required: true,
      id: 'selection',
      class: selectClass(),
      style: selectStyle(),
      value: props.value,
      disabled: props.disabled,
      onChange({ target }: Event) {
        const value =
          (target as HTMLInputElement)?.value ?? '';
        selectionChanged(value);
      },
    },
    [
      h('option', { value: '' }, defaultText()),
      items.map((item) => {
        return h(
          'option',
          { key: item, value: item },
          item,
        );
      }),
    ],
  );
}

function KeyValuePairSelection(
  items: KeyValuePair<string>,
): VNode {
  return h(
    'select',
    {
      required: true,
      id: 'selection',
      class: selectClass(),
      style: selectStyle(),
      value: props.value,
      disabled: props.disabled,
      onChange({ target }: Event) {
        const value =
          (target as HTMLInputElement)?.value ?? '';
        selectionChanged(value);
      },
    },
    [
      h('option', { value: '' }, defaultText()),
      Object.entries(items).map(([key, value]) => {
        return h('option', { key: key, value: value }, key);
      }),
    ],
  );
}

function selectionChanged(value: any) {
  selectedValue.value = value;
  emit('updated', selectedValue.value);
}
</script>

<style scoped>
select.form-select {
  border: 0;
  outline: 0;
  border-radius: 0%;
  border-bottom: 1px solid white;
  text-align: left;
  background-color: transparent;
  background-image: none;
}

.form-floating>.form-select~label::after {
  background-color: transparent;
}

select.form-select:focus,
:active {
  box-shadow: none;
}

select.form-select:hover:not([disabled]) {
  background-image: url('data:image/svg+xml;charset=utf-8,%3Csvg xmlns=%27http://www.w3.org/2000/svg%27 viewBox=%270 0 16 16%27%3E%3Cpath fill=%27none%27 stroke=%27%23d4d6d9%27 stroke-linecap=%27round%27 stroke-linejoin=%27round%27 stroke-width=%272%27 d=%27m2 5 6 6 6-6%27/%3E%3C/svg%3E');
  box-shadow: none;
}

select.form-select:first-of-type {
  border-bottom: 0px solid white;
}

select.form-select:required:invalid {
  color: gray;
  border-bottom: 1px solid white;
}

.form-floating>.form-select~label {
  opacity: 0.6;
  transform: scale(0.85) translateY(-0.7rem) translateX(0.15rem);
}

select.form-select:disabled {
  color: gray;
  background-color: transparent;
}
</style>

<template>
  <!-- 
  <v-select label="Select" :items="items" :value="value" :disabled="disabled"
    variant="underlined"></v-select> -->
    <!-- <Select v-model="selectedCity" :options="item" optionLabel="name" placeholder="Select a City" class="w-full md:w-56" /> -->

  <div v-if="props.label != ''" class="form-floating col-sm-5">
    <createSelection></createSelection>
    <label for="selection" class="form-label">{{
      props.label
    }}</label>
  </div>
  <div v-else>
    <createSelection></createSelection>
  </div>
</template>
