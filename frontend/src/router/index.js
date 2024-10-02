import { createRouter, createWebHistory } from 'vue-router';
import DevicePage from '../components/device/DevicePage.vue';
import Dashboard from '../components/device-dashboard/DeviceDashboard.vue';
import AutomationsViewer from '../components/automations/Viewer.vue';
import AutomationsEditor from '../components/automations/Editor.vue';
import AutomationsCreator from '../components/automations/Creator.vue';
import Settings from '../components/hub/settings/Settings.vue';

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      name: 'home',
      component: Dashboard,
      meta: {
        title: 'Node-herder - Home',
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
