import { createRouter, createWebHistory } from 'vue-router';
import DevicePage from '../components/device/DevicePage.vue';
import Dashboard from '../components/dashboards/DeviceDashboard.vue';
import AutomationsViewer from '../components/automations/Viewer.vue';
import AutomationsEditor from '../components/automations/Editor.vue';
import AutomationsCreator from '../components/automations/Creator.vue';
import Settings from '../components/hub/settings/Settings.vue';
import ConsoleViewer from '../components/hub/console/ConsoleViewer.vue';
import DeviceList from '../components/device-list/DeviceList.vue';
import GroupDashboard from '../components/dashboards/GroupDashboard.vue';
import DeviceView from '@/components/device/DeviceView.vue';
import LoginPage from '@/components/auth/LoginPage.vue';
import AssistantView from '@/components/assistant/AssistantView.vue';
import { store } from '@/store/index.js';
import { RouteName } from '@/types/router';
import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    path: '/',
    name: RouteName.Login,
    component: LoginPage,
    meta: {
      title: 'Node-herder - Login',
      public: true,
    },
  },
  {
    path: '/devicedashboard',
    name: RouteName.Devices,
    component: Dashboard,
    meta: {
      title: 'Node-herder - Device Dashboard',
      requiresAuth: true,
    },
  },
  {
    path: '/groupdashboard',
    name: RouteName.GroupDashboard,
    component: GroupDashboard,
    meta: {
      title: 'Node-herder - Groups Dashboard',
      requiresAuth: true,
    },
  },
  {
    path: '/devicelist',
    name: RouteName.DeviceList,
    component: DeviceList,
    meta: {
      title: 'Node-herder - Device List',
      requiresAuth: true,
    },
  },
  {
    path: '/viewer',
    name: RouteName.Viewer,
    component: AutomationsViewer,
    meta: {
      title: 'Node-herder - Automation viewer',
      requiresAuth: true,
    },
  },
  {
    path: '/settings',
    name: RouteName.Settings,
    component: Settings,
    meta: {
      title: 'Node-herder - Settings',
      requiresAuth: true,
    },
  },
  {
    path: '/consoleviewer',
    name: RouteName.ConsoleViewer,
    component: ConsoleViewer,
    meta: {
      title: 'Node-herder - Console viewer',
      requiresAuth: true,
    },
  },
  {
    path: '/creator',
    name: RouteName.Creator,
    component: AutomationsCreator,
    meta: {
      title: 'Node-herder - Creator',
      requiresAuth: true,
    },
  },
  {
    path: '/editor/:id',
    name: RouteName.Editor,
    component: AutomationsEditor,
    props: true,
    meta: {
      title: 'Node-herder - Editor',
      requiresAuth: true,
    },
  },
  {
    path: '/devicepage/:id',
    name: RouteName.DevicePage,
    component: DevicePage,
    props: true,
    meta: {
      title: 'Node-herder - Device page',
      requiresAuth: true,
    },
  },
  {
    path: '/deviceview/:id',
    name: RouteName.DeviceView,
    component: DeviceView,
    props: true,
    meta: {
      title: 'Node-herder - Device view',
      requiresAuth: true,
    },
  },
  {
    path: '/assistant',
    name: RouteName.Assistant,
    component: AssistantView,
    meta: {
      title: 'Node-herder - Assistant',
      requiresAuth: true,
    },
  },
];

export const router = createRouter({
  history: createWebHistory(),
  routes,
});

router.beforeEach((to, from, next) => {
  // document.title = to.meta.title || 'Node-herder';
  const isAuthenticated = store.getters['auth/isAuthenticated']();
  if (to.meta.requiresAuth && !isAuthenticated) {
    console.log('Route requires auth and user is not authenticated, redirecting to login.');
    return next({ name: 'login' });
  }

  if (to.meta.public && isAuthenticated && to.name === 'login') {
    console.log('User is already authenticated, redirecting to group dashboard.');
    return next({ name: 'groupdashboard' });
  }
  next();
});

export default router;
