<script setup lang="ts">
  /*
   * Signed-in application frame: navigation, header, status banners and the shared hosts
   * (confirm dialog, legacy dialogs, toasts, appearance). Routes with meta.layout = 'panel'
   * get the full screen (panel mode, FR-10) with the same hosts.
   */
  import { computed, onMounted, ref, watch } from 'vue';
  import { useRoute } from 'vue-router';
  import { startHubConnection } from '@/composables/useHubConnection';
  import { useActivityFeed } from '@/composables/useActivityFeed';
  import Notifications from '@/components/hub/alerts/Notifications.vue';
  import DialogHost from '@/components/dialogs/DialogHost.vue';
  import UiConfirmHost from '@/components/ui/UiConfirmHost.vue';
  import NavSidebar from './NavSidebar.vue';
  import AppHeader from './AppHeader.vue';
  import StatusBanners from './StatusBanners.vue';
  import AppearanceDrawer from './AppearanceDrawer.vue';

  const route = useRoute();
  const isPanel = computed(() => route.meta.layout === 'panel');
  const navOpen = ref(false);
  const appearanceOpen = ref(false);
  const main = ref<HTMLElement | null>(null);

  useActivityFeed();
  onMounted(startHubConnection);

  watch(
    () => route.fullPath,
    () => {
      navOpen.value = false;
      main.value?.scrollTo?.({ top: 0 });
    }
  );

  function skipToContent() {
    main.value?.focus();
  }
</script>

<template>
  <template v-if="isPanel">
    <RouterView v-slot="{ Component }">
      <component :is="Component" @open-appearance="appearanceOpen = true" />
    </RouterView>
  </template>
  <div v-else class="nh-app" :data-nav-open="navOpen" @keydown.esc="navOpen = false">
    <button type="button" class="nh-skip nh-btn" @click="skipToContent">Skip to content</button>
    <div id="nh-nav" class="nh-app-nav">
      <NavSidebar @navigate="navOpen = false" />
    </div>
    <div class="nh-scrim" aria-hidden="true" @click="navOpen = false" />
    <div class="nh-app-main">
      <AppHeader :nav-open="navOpen" @toggle-nav="navOpen = !navOpen" @open-appearance="appearanceOpen = true" />
      <StatusBanners />
      <main id="nh-content" ref="main" class="nh-content" tabindex="-1">
        <RouterView />
      </main>
    </div>
  </div>

  <AppearanceDrawer :open="appearanceOpen" @close="appearanceOpen = false" />
  <UiConfirmHost />
  <DialogHost />
  <Notifications />
</template>

<style scoped>
  .nh-app {
    display: grid;
    grid-template-columns: var(--nh-nav-w) minmax(0, 1fr);
    height: 100dvh;
  }
  .nh-app-nav {
    min-height: 0;
  }
  .nh-app-main {
    display: flex;
    flex-direction: column;
    min-width: 0;
    min-height: 0;
    overflow: auto;
  }
  .nh-content {
    flex: 1;
    padding: calc(var(--nh-space) * 1.25) calc(var(--nh-space) * 1.5) 4rem;
    outline: none;
    min-width: 0;
  }
  .nh-skip {
    position: fixed;
    left: 0.5rem;
    top: -4rem;
    z-index: 2000;
  }
  .nh-skip:focus {
    top: 0.5rem;
  }
  .nh-scrim {
    display: none;
  }
  @media (max-width: 960px) {
    .nh-app {
      grid-template-columns: minmax(0, 1fr);
    }
    .nh-app-nav {
      position: fixed;
      inset: 0 auto 0 0;
      width: min(18rem, 85vw);
      z-index: 60;
      transform: translateX(-105%);
      /* hide only after the slide-out finishes, so closed links cannot take focus */
      transition:
        transform var(--nh-motion),
        visibility 0s linear var(--nh-motion);
      box-shadow: var(--nh-shadow-2);
      visibility: hidden;
    }
    [data-nav-open='true'] .nh-app-nav {
      transform: none;
      visibility: visible;
      transition: transform var(--nh-motion);
    }
    [data-nav-open='true'] .nh-scrim {
      display: block;
      position: fixed;
      inset: 0;
      z-index: 55;
      background: rgb(0 0 0 / 0.4);
    }
    .nh-content {
      padding: var(--nh-space) var(--nh-space) 4rem;
    }
  }
  @media (max-width: 420px) {
    .nh-content {
      padding: 0.75rem 0.75rem 3rem;
    }
  }
</style>
