<script setup lang="ts">
  /* Top bar: drawer button and logo on small screens, page title, hub status and quick settings. */
  import { computed } from 'vue';
  import { useRoute } from 'vue-router';
  import { RouteName } from '@/types/router';
  import { useSession } from '@/composables/useSession';
  import { useThemeSettings } from '@/composables/useThemeSettings';
  import UiButton from '@/components/ui/UiButton.vue';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import AppLogo from '@/components/ui/AppLogo.vue';
  import ConnectionStatus from './ConnectionStatus.vue';

  defineProps<{ navOpen: boolean }>();
  defineEmits<{ (e: 'toggle-nav'): void; (e: 'open-appearance'): void }>();

  const route = useRoute();
  const { user } = useSession();
  const { mode, toggleMode } = useThemeSettings();
  const title = computed(() => (route.meta.title as string | undefined) ?? '');
</script>

<template>
  <header class="nh-header">
    <UiButton class="nh-menu-btn" variant="ghost" icon="menu" label="Open navigation" :aria-expanded="navOpen" aria-controls="nh-nav" @click="$emit('toggle-nav')" />
    <RouterLink :to="{ name: RouteName.GroupDashboard }" class="nh-header-logo" aria-label="NodeHerder home">
      <AppLogo :size="32" :wordmark="false" />
    </RouterLink>
    <div class="nh-header-title">{{ title }}</div>
    <div class="nh-header-actions">
      <ConnectionStatus />
      <UiButton variant="ghost" :icon="mode === 'dark' ? 'light' : 'dark'" :label="mode === 'dark' ? 'Switch to light mode' : 'Switch to dark mode'" @click="toggleMode" />
      <UiButton variant="ghost" icon="palette" label="Appearance" @click="$emit('open-appearance')" />
      <span v-if="user?.username" class="nh-user" :title="`Signed in as ${user.username}`">
        <UiIcon name="user" /><span class="hide-sm">{{ user.username }}</span>
      </span>
    </div>
  </header>
</template>

<style scoped>
  .nh-header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    min-height: var(--nh-header-h);
    padding: 0 var(--nh-space);
    background: var(--nh-header-bg, var(--nh-bg));
    position: sticky;
    top: 0;
    z-index: 20;
  }
  .nh-header-title {
    font-weight: 600;
    font-size: 1.05rem;
    min-width: 0;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .nh-header-actions {
    margin-left: auto;
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }
  .nh-user {
    display: inline-flex;
    align-items: center;
    gap: 0.35rem;
    color: var(--nh-text-muted);
    font-size: 0.85rem;
    padding: 0 0.3rem;
    max-width: 12rem;
    overflow: hidden;
    text-overflow: ellipsis;
  }
  .nh-menu-btn,
  .nh-header-logo {
    display: none;
  }
  .nh-header-logo {
    text-decoration: none;
  }
  @media (max-width: 960px) {
    .nh-menu-btn,
    .nh-header-logo {
      display: inline-flex;
    }
  }
  @media (max-width: 420px) {
    .nh-header {
      padding: 0 0.5rem;
    }
    .nh-header-title {
      font-size: 0.95rem;
    }
  }
</style>
