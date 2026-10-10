import { createRouter, createWebHistory } from 'vue-router';
import type { RouteRecordRaw } from 'vue-router';
import LoginPage from '@/components/auth/LoginPage.vue';
import { store } from '@/store/index.js';
import { RouteName } from '@/types/router';

declare module 'vue-router' {
  interface RouteMeta {
    /** Header and document title. */
    title?: string;
    public?: boolean;
    requiresAuth?: boolean;
    /** 'panel' renders the route full screen without the app shell (panel mode). */
    layout?: 'app' | 'panel';
  }
}

/* Paths are unchanged from the previous UI so bookmarks keep working; screens load on demand. */
const routes: RouteRecordRaw[] = [
  { path: '/', name: RouteName.Login, component: LoginPage, meta: { title: 'Sign in', public: true } },
  {
    path: '/overview',
    name: RouteName.Overview,
    component: () => import('@/components/overview/OverviewPage.vue'),
    meta: { title: 'Overview', requiresAuth: true },
  },
  {
    path: '/groupdashboard',
    name: RouteName.GroupDashboard,
    component: () => import('@/components/dashboards/GroupDashboard.vue'),
    meta: { title: 'Home', requiresAuth: true },
  },
  {
    path: '/devicedashboard',
    name: RouteName.Devices,
    component: () => import('@/components/dashboards/DeviceDashboard.vue'),
    meta: { title: 'Devices', requiresAuth: true },
  },
  {
    path: '/devicelist',
    name: RouteName.DeviceList,
    component: () => import('@/components/device-list/DeviceList.vue'),
    meta: { title: 'Device list', requiresAuth: true },
  },
  {
    path: '/devicepage/:id',
    name: RouteName.DevicePage,
    component: () => import('@/components/device/DevicePage.vue'),
    props: true,
    meta: { title: 'Device', requiresAuth: true },
  },
  // Kept for old links: the single-device view is now the device page.
  { path: '/deviceview/:id', name: RouteName.DeviceView, redirect: (to) => ({ name: RouteName.DevicePage, params: { id: to.params.id } }) },
  {
    path: '/viewer',
    name: RouteName.Viewer,
    component: () => import('@/components/automations/AutomationList.vue'),
    meta: { title: 'Automations', requiresAuth: true },
  },
  {
    path: '/creator',
    name: RouteName.Creator,
    component: () => import('@/components/automations/AutomationCreator.vue'),
    meta: { title: 'New automation', requiresAuth: true },
  },
  {
    path: '/editor/:id',
    name: RouteName.Editor,
    component: () => import('@/components/automations/editor/AutomationEditor.vue'),
    props: true,
    meta: { title: 'Automation', requiresAuth: true },
  },
  {
    path: '/assistant',
    name: RouteName.Assistant,
    component: () => import('@/components/assistant/AssistantView.vue'),
    meta: { title: 'Assistant', requiresAuth: true },
  },
  {
    path: '/consoleviewer',
    name: RouteName.ConsoleViewer,
    component: () => import('@/components/hub/console/ConsoleViewer.vue'),
    meta: { title: 'Console', requiresAuth: true },
  },
  {
    path: '/settings',
    name: RouteName.Settings,
    component: () => import('@/components/hub/settings/Settings.vue'),
    meta: { title: 'Settings', requiresAuth: true },
  },
  {
    path: '/panel',
    name: RouteName.Panel,
    component: () => import('@/components/panel/PanelMode.vue'),
    meta: { title: 'Panel', requiresAuth: true, layout: 'panel' },
  },
  { path: '/:pathMatch(.*)*', redirect: { name: RouteName.GroupDashboard } },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to) => {
  const isAuthenticated = store.getters['auth/isAuthenticated']();
  if (to.meta.requiresAuth && !isAuthenticated) {
    return { name: RouteName.Login, query: to.fullPath !== '/' ? { redirect: to.fullPath } : undefined };
  }
  if (to.meta.public && isAuthenticated && to.name === RouteName.Login) {
    return { name: RouteName.GroupDashboard };
  }
  return true;
});

router.afterEach((to) => {
  document.title = to.meta.title ? `${to.meta.title} · NodeHerder` : 'NodeHerder';
});

export default router;
