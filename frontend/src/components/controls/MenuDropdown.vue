<script setup lang="ts">
  import { ref, PropType, watch, computed } from 'vue';
  import { MenuBarItem } from '@/types/controls.type';
  import Icon from '../controls/Icon.vue';
  import Menu from 'primevue/menu';
  import { IconProps } from '@/types/icon.type';

  const props = defineProps({
    icon: {
      type: Object as PropType<IconProps>,
      default: [],
      required: true,
    },
    text: {
      type: String,
      default: '',
      required: false,
    },
    backgroundColor: {
      type: String,
      default: 'var(--surface-card)',
    },
    selected: {
      type: [Object, String, Number, Boolean, Array] as PropType<any>,
      default: null,
    },
    size: {
      type: Number,
      default: 38,
    },
    disabled: {
      type: Boolean,
      default: false,
    },
    children: {
      type: Object as PropType<MenuBarItem[]>,
      default: [],
      required: true,
    },
  });

  const emit = defineEmits(['close']);

  function close() {
    emit('close', false);
  }

  const menu = ref<InstanceType<typeof Menu> | null>(null);
  const toggleMenu = (event: Event) => {
    menu.value?.toggle(event);
  };

  const selected = ref<any | null>(props.selected);

  watch(
    () => props.selected,
    (val) => {
      selected.value = val;
    }
  );

  const convertedChildren = computed(() => {
    return props.children.map((item) => ({
      ...item,
      command: () => {
        // we only want to update selectd value if its been provided
        // as the menu can select items or just on click to only trigger command
        if (selected.value != null) {
          selected.value = item.value;
        }
        item.command(item.value);
        close();
      },
    }));
  });
</script>

<template>
  <div class="relative inline-block">
    <div
      class="button-content"
      :style="{ backgroundColor, opacity: disabled ? 0.5 : 1, pointerEvents: disabled ? 'none' : 'auto' }"
      @click="toggleMenu">
      <Icon :icon="icon" :size="size" background="transparent" />
      <span class="text-md font-medium whitespace-nowrap"> {{ text }} </span>
    </div>
    <Menu ref="menu" :model="convertedChildren" popup appendTo="body">
      <template #item="{ item }">
        <div class="flex justify-between items-center w-full">
          <span>{{ item.label }}</span>
          <i v-if="selected !== null && selected === item.value" class="pi pi-check ml-2"></i>
        </div>
      </template>
    </Menu>
  </div>
</template>
<style scoped>
  .button-content {
    display: flex;
    gap: 3px;
    padding: 4px 12px;
    border-radius: 10px;
    cursor: pointer;
    transition: opacity 0.3s ease;
    align-items: center;

    background-color: var(--p-button-secondary-bg);
    color: var(--p-button-secondary-color);
    border: 1px solid var(--p-button-secondary-border-color, transparent);
  }

  .button-content:hover {
    background-color: var(--p-button-secondary-hover-bg);
    color: var(--p-button-secondary-hover-color);
    border-color: var(--p-button-secondary-hover-border-color, transparent);
  }
</style>
