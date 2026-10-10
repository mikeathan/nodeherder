<script setup lang="ts">
  /* Main navigation. A fixed column on wide screens, a drawer below 960px. */
  import { useRoute } from 'vue-router';
  import { RouteName } from '@/types/router';
  import { useSession } from '@/composables/useSession';
  import { usePermitJoin } from '@/composables/usePermitJoin';
  import { confirm } from '@/composables/useConfirm';
  import UiIcon from '@/components/ui/UiIcon.vue';
  import AppLogo from '@/components/ui/AppLogo.vue';
  import { isNavActive, NAV_ITEMS } from './navigation';

  defineEmits<{ (e: 'navigate'): void }>();

  const route = useRoute();
  const { signOut } = useSession();
  const permit = usePermitJoin();
  const version = __APP_VERSION__;

  async function togglePermitJoin() {
    if (permit.active.value) {
      permit.stop();
      return;
    }
    const ok = await confirm({
      title: 'Allow new Zigbee devices?',
      message: 'New devices can join the network for 2 minutes. Put the device in pairing mode after you start.',
      confirmLabel: 'Permit join',
    });
    if (ok) permit.start();
  }
</script>

<template>
  <nav class="nh-nav" aria-label="Main">
    <RouterLink :to="{ name: RouteName.Overview }" class="nh-brand" @click="$emit('navigate')">
      <AppLogo :size="42" :version="version" />
    </RouterLink>
    <ul class="nh-nav-list">
      <li v-for="item in NAV_ITEMS" :key="item.route">
        <RouterLink
          :to="{ name: item.route }"
          class="nh-nav-item"
          :class="{ 'is-active': isNavActive(item, route.name) }"
          :aria-current="isNavActive(item, route.name) ? 'page' : undefined"
          @click="$emit('navigate')">
          <UiIcon :name="item.icon" />
          <span>{{ item.label }}</span>
        </RouterLink>
      </li>
    </ul>
    <div class="nh-nav-foot">
      <RouterLink :to="{ name: RouteName.Panel }" class="nh-nav-item" @click="$emit('navigate')">
        <UiIcon name="panel" /><span>Panel mode</span>
      </RouterLink>
      <button type="button" class="nh-nav-item" :class="{ 'is-on': permit.active.value }" :aria-pressed="permit.active.value" @click="togglePermitJoin">
        <UiIcon name="join" />
        <span v-if="permit.active.value">Joining · {{ permit.label.value }}</span>
        <span v-else>Permit join</span>
      </button>
      <button type="button" class="nh-nav-item" @click="signOut">
        <UiIcon name="signOut" /><span>Sign out</span>
      </button>
    </div>
  </nav>
</template>

<style scoped>
  .nh-nav {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    height: 100%;
    background: var(--nh-nav-bg, var(--nh-surface));
    border-right: var(--nh-border-w) solid var(--nh-border);
    padding: 0.75rem 0.6rem;
    overflow-y: auto;
  }
  .nh-brand {
    display: flex;
    align-items: center;
    padding: 0.2rem 0.5rem 0.9rem;
    text-decoration: none;
    border-radius: var(--nh-radius-m);
  }
  .nh-nav-list {
    list-style: none;
    margin: 0;
    padding: 0;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }
  .nh-nav-foot {
    margin-top: auto;
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    padding-top: 0.75rem;
    border-top: var(--nh-border-w) solid var(--nh-border);
  }
  .nh-nav-item {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0 0.7rem;
    min-height: var(--nh-ctl-h);
    border-radius: var(--nh-radius-m);
    text-decoration: none;
    color: inherit;
    background: none;
    border: 0;
    cursor: pointer;
    text-align: left;
    width: 100%;
    font-weight: 500;
    transition: background var(--nh-motion);
  }
  .nh-nav-item:hover {
    background: var(--nh-surface-2);
  }
  .nh-nav-item.is-active {
    background: var(--nh-accent-soft);
    color: var(--nh-accent);
    font-weight: 650;
  }
  .nh-nav-item.is-on {
    color: var(--nh-accent);
  }
</style>
