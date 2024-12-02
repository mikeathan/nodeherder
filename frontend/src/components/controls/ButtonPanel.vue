<script setup lang="ts">
import { PropType, h, VNode } from 'vue';
import {
  ButtonPanelType,
  ButtonType,
  isDropdown,
  DropDownType,
  ButtonClickEventType,
  DropDownItemType,
} from '@/types/controls.type';

const props = defineProps({
  buttons: {
    type: Object as PropType<Array<ButtonPanelType>>,
    default: [],
    required: true,
  },
});

// function Panel() {
//   return props.buttons.map((item: ButtonPanelType) =>
//     isDropdown(item)
//       ? createDropdown(item as DropDownType)
//       : createButton(item as ButtonType),
//   );
// }

// function createDropdown(dropDown: DropDownType): VNode {
//   return h(
//     'Dropdown',
//     {
//       className: 'btn-light',
//       disabled: dropDown.disabled,
//       items: dropDown.items,
//     },
//     dropDown.name,
//   );
// }

// function createButton(button: ButtonType): VNode {
//   return h(
//     'button',
//     {
//       class: 'btn btn-light',
//       disabled: button.disabled,
//       onClick: (event: any) => {
//         event.preventDefault();
//         try {
//           button.click(event);
//         } catch (error) {
//           console.error(
//             'Error during button click:',
//             error,
//           );
//         }
//       },
//     },
//     button.name,
//   );
// }

function isButtonType(
  item: ButtonPanelType,
): item is ButtonType {
  return (
    item.hasOwnProperty('name') &&
    !item.hasOwnProperty('items')
  );
}
function isDropdownType(
  item: ButtonPanelType,
): item is DropDownType {
  return (
    item.hasOwnProperty('name') &&
    item.hasOwnProperty('items')
  );
}

function createEvent(event: Event, button: ButtonType) {
  event.preventDefault();
  try {
    console.log("[DEBUG] Buttonpanel - createEvent", event, button);

    button.click(event);
  } catch (error) {
    console.error('Error during button click:', error);
  }
}

function createDropEvent(
  event: Event,
  button: DropDownItemType,
) {
  event.preventDefault();
  try {
    button.click(event);
  } catch (error) {
    console.error('Error during button click:', error);
  }
}
</script>

<!-- // if button is dropdown use <SplitButton label="Save" @click="save" :model="items" /> -->
<template>
  <div class="grid grid-cols-4 gap-1">
    <div v-for="item in props.buttons" :key="item.name">
      <template v-if="isButtonType(item)">
        <Button :key="item.name" :label="item.name" severity="secondary" variant="outlined" :disabled="item.disabled"
          @click="(event) => createEvent(event, item)" raised />
      </template>
      <template v-else-if="isDropdownType(item)">
        <SplitButton v-for="dropdownItem in item.items" :key="dropdownItem.name" :label="dropdownItem.name" @click="(event) => createDropEvent(event, dropdownItem)
          " />
      </template>
    </div>
    <!-- <Button
        v-for="buttonInfo in props.buttons"
        :key="buttonInfo.name"
        :label="buttonInfo.name"
        :disabled="buttonInfo.disabled"
        @click="(event)=>createEvent(event, buttonInfo as ButtonType)" /> -->
    <!-- <component :is="Panel"></component> -->
    <!-- <slot></slot> -->
  </div>
</template>
