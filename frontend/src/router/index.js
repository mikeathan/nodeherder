import { createRouter, createWebHistory } from 'vue-router';
import DevicePage from '../components/device/DevicePage.vue';
import Dashboard from '../components/device-dashboard/DeviceDashboard.vue';
import AutomationsViewer from '../components/automations/Viewer.vue';
import AutomationsEditor from '../components/automations/Editor.vue';
import AutomationsCreator from '../components/automations/Creator.vue';
import Settings from '../components/hub/settings/Settings.vue';
import ConsoleViewer from '../components/hub/console/ConsoleViewer.vue';
import DeviceList from '../components/device-list/DeviceList.vue';
import GroupDashboard from '../components/group-dashboard/GroupDashboard.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/deviceDashboard',
      name: 'home',
      component: Dashboard,
      meta: {
        title: 'Node-herder - Device Dashboard',
      },
    },
    {
      path: '/',
      name: 'groups',
      component: GroupDashboard,
      meta: {
        title: 'Node-herder - Groups Dashboard',
      },
    },
    {
      path: '/devicelist',
      name: 'deicelist',
      component: DeviceList,
      meta: {
        title: 'Node-herder - Device List',
      },
    },
    {
      path: '/viewer',
      name: 'viewer',
      component: AutomationsViewer,
      meta: {
        title: 'Node-herder - Automation viewer',
      },
    },
    {
      path: '/settings',
      name: 'settings',
      component: Settings,
      meta: {
        title: 'Node-herder - Settings',
      },
    },
    {
      path: '/consoleviewer',
      name: 'consoleviewer',
      component: ConsoleViewer,
      meta: {
        title: 'Node-herder - Console viewer',
      },
    },
    {
      path: '/creator',
      name: 'creator',
      component: AutomationsCreator,
      meta: {
        title: 'Node-herder - Creator',
      },
    },
    {
      path: '/editor/:id',
      name: 'editor',
      component: AutomationsEditor,
      props: true,
      meta: {
        title: 'Node-herder - Editor',
      },
    },
    {
      path: '/devicepage/:id',
      name: 'devicepage',
      component: DevicePage,
      props: true,
      meta: {
        title: 'Node-herder - Device page',
      },
    },
    {
      path: '/dashboard',
      name: 'dashboard',
      component: Dashboard,
      meta: {
        title: 'Node-herder - Dashboard',
      },
    },
  ],
});
export default router;

// Global navigation guard to
// set the title based on the route
router.beforeEach((to, from, next) => {
  document.title = to.meta.title || 'Node-herder';
  next();
});
