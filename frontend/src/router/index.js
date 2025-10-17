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
import { DashboardModes } from '@/types/controls.type';
import DeviceView from '@/components/device/DeviceView.vue';
import LoginPage from '@/components/auth/LoginPage.vue';
import { store } from '@/store/index.js';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'login',
      component: LoginPage,
      meta: {
        title: 'Node-herder - Login',
        public: true,
      },
    },
    {
      path: '/devicedashboard',
      name: 'devices',
      component: Dashboard,
      meta: {
        title: 'Node-herder - Device Dashboard',
        requiresAuth: true,
      },
    },
    {
      path: '/groupdashboard',
      name: 'groupdashboard',
      component: GroupDashboard,
      meta: {
        title: 'Node-herder - Groups Dashboard',
        requiresAuth: true,
      },
    },
    {
      path: '/devicelist',
      name: 'devicelist',
      component: DeviceList,
      meta: {
        title: 'Node-herder - Device List',
        requiresAuth: true,
      },
    },
    {
      path: '/viewer',
      name: 'viewer',
      component: AutomationsViewer,
      meta: {
        title: 'Node-herder - Automation viewer',
        requiresAuth: true,
      },
    },
    {
      path: '/settings',
      name: 'settings',
      component: Settings,
      meta: {
        title: 'Node-herder - Settings',
        requiresAuth: true,
      },
    },
    {
      path: '/consoleviewer',
      name: 'consoleviewer',
      component: ConsoleViewer,
      meta: {
        title: 'Node-herder - Console viewer',
        requiresAuth: true,
      },
    },
    {
      path: '/creator',
      name: 'creator',
      component: AutomationsCreator,
      meta: {
        title: 'Node-herder - Creator',
        requiresAuth: true,
      },
    },
    {
      path: '/editor/:id',
      name: 'editor',
      component: AutomationsEditor,
      props: true,
      meta: {
        title: 'Node-herder - Editor',
        requiresAuth: true,
      },
    },
    {
      path: '/devicepage/:id',
      name: 'devicepage',
      component: DevicePage,
      props: true,
      meta: {
        title: 'Node-herder - Device page',
        requiresAuth: true,
      },
    },
    {
      path: '/deviceview/:id',
      name: 'deviceview',
      component: DeviceView,
      props: true,
      meta: {
        title: 'Node-herder - Device view',
        requiresAuth: true,
      },
    },
  ],
});

router.beforeEach((to, from, next) => {
  document.title = to.meta.title || 'Node-herder';
  const isAuthenticated = store.getters['auth/isAuthenticated']();
  if (to.meta.requiresAuth && !isAuthenticated) {
    return next({ name: 'login' });
  }

  if (to.meta.public && isAuthenticated && to.name === 'login') {
    return next({ name: 'groupdashboard' });
  }
  next();
});

export default router;
