<script setup lang="tsx">
  import { onBeforeMount, ref, h, computed, PropType } from 'vue';
  import { onMounted } from 'vue';
  import { MenuBarItem } from '@/types/controls.type';

  const props = defineProps({
    items: {
      type: Object as PropType<MenuBarItem[]>,
      default: [],
      required: true,
    },
  });

  const isMenuVisible = ref(false);
  const isMobileView = ref(false);

  const toggleMenu = () => {
    isMenuVisible.value = !isMenuVisible.value;
  };
  function checkMobileView() {
    isMobileView.value = window.innerWidth <= 768;
  }

  const logoItem = computed(() => {
    return props.items.find((item) => item.isLogo);
  });
  
  const renderMenuItem = (item: MenuBarItem) => {
    if (item.isLogo) return null; // Logo items are handled separately
    if (!item.custom) {
      return (
        <li class="menuitem">
          <a class="menuitem-link" href="javascript:void(0)" onClick={item.command}>
            {item.icon ? <span class={`menuitem-icon ${item.icon}`}></span> : null}
            <span class="menuitem-text">{item.label}</span>
          </a>
        </li>
      );
    } else if (item.template) {
      return (
        <li class="menuitem">
          <div class="menuitem-text">{item.template()}</div>
        </li>
      );
    }
    return null;
  };

  onMounted(() => {
    checkMobileView();
    window.addEventListener('resize', checkMobileView);
  });

  onBeforeMount(() => {
    window.removeEventListener('resize', checkMobileView);
  });
</script>

<style scoped>
  .menubar {
    display: flex;
    align-items: center;
    width: 100%;
    background: var(--p-content-background);
    color: var(--p-content-color);
    border: 1px solid var(--bs-border-color);
    padding: var(--p-list-option-padding);
    border-radius: var(--p-border-radius-md);
  }

  .menubar-root-list {
    display: flex;
    list-style: none;
    justify-content: flex-end;
    background: var(--p-content-background);
    margin: 0;
    padding: var(--p-list-option-padding);
    border-radius: var(--p-border-radius-md);
    margin-left: auto;
  }

  .menuitem-link {
    display: flex;
    align-items: center;
    margin: 0px 15px 0px 0px;
    text-decoration: none;
    color: inherit;
    cursor: pointer;
  }

  .menuitem-link:hover {
    background-color: var(--surface-hover);
    color: var(--primary-color);
  }

  .menuitem-icon {
    margin-right: 0.5rem;
  }

  @media (max-width: 768px) {
    .menubar-root-list {
      flex-direction: column;
    }

    .menuitem {
      width: 100%;
    }
  }

  .hamburger {
    display: flex;
    margin-left: auto;
    padding: 0.5rem;
    cursor: pointer;
  }

  .hamburger i {
    font-size: 1.5rem;
  }
</style>

<template>
  <!-- Desktop View -->
  <div v-if="!isMobileView">
    <div class="menubar">
      <template v-if="logoItem">
        <component :is="logoItem.template" />
      </template>
      <ul class="menubar-root-list">
        <li v-for="(item, index) in props.items" :key="index">
          <component :is="renderMenuItem(item)" />
        </li>
      </ul>
    </div>
  </div>
  <!-- Mobile View -->
  <div v-else>
    <div class="menubar">
      <template v-if="logoItem">
        <component :is="logoItem.template" />
      </template>
      <div class="hamburger" @click="toggleMenu">
        <i class="pi pi-bars"></i>
      </div>
    </div>
    <transition name="menu-slide">
      <ul v-if="isMenuVisible" class="menubar-root-list">
        <li v-for="(item, index) in props.items" :key="index" class="menuitem">
          <component :is="renderMenuItem(item)" />
        </li>
      </ul>
    </transition>
  </div>
</template>
