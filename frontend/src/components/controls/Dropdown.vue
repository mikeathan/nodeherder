<script setup lang="ts">
  import { computed, PropType, ref, watch } from 'vue';
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
    selected: {
      type: [Object, String, Number, Boolean, Array] as PropType<any>,
      default: null,
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
    text: {
      type: Boolean,
      default: false,
    },
    severity: {
      type: String as PropType<
        'primary' | 'secondary' | 'success' | 'info' | 'warning' | 'help' | 'danger' | undefined
      >,
      default: 'primary',
    },
  });
  const selected = ref<any>(props.selected);

  const convert = computed(() => {
    return props.items.map((item) => {
      return {
        label: item.label,
        icon: selected?.value === item.value ? 'pi pi-check' : '',
        command: () => {
          selected.value = item.value;
          item.command(item.value);
        },
      };
    });
  });
  watch(
    () => props.selected,
    (val) => {
      selected.value = val;
    }
  );
</script>
<style scoped>
  /* Safety: ensure long labels dxon’t break layout */
  :deep(.p-button .p-button-label) {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }
</style>
<template>
  <SplitButton
    :pt="{
      root: { class: 'w-full' },
      button: { class: 'w-full justify-center px-2 py-1 text-xs' },
      menuButton: { class: 'px-2 py-1 text-xs' },
    }"
    :label="props.label"
    :model="convert"
    :disabled="props.disabled"
    :icon="props.icon"
    :size="props.size"
    :severity="props.severity"
    :text="props.text"
    :class="['w-full', props.className]"
    :buttonProps="{ class: 'w-full justify-center px-2 py-1 text-xs' }"
    :menuButtonProps="{ class: 'px-2 py-1 text-xs' }" />
</template>
