import { createRouter, createWebHistory } from "vue-router";
import DevicePage from "../components/device/DevicePage.vue";
import Dashboard from "../components/device-dashboard/DeviceDashboard.vue";
import AutomationsViewer from "../components/automations/viewer.vue";
import AutomationsEditor from "../components/automations/editor.vue";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      name: "home",
      component: Dashboard,
    },
    {
      path: "/viewer",
      name: "viewer",
      component: AutomationsViewer,
    },
    {
      path: "/editor/:id",
      name: "editor",
      component: AutomationsEditor,
      props: true,
    },
    {
      path: "/devicepage/:id",
      name: "devicepage",
      component: DevicePage,
      props: true,
    },
    {
      path: "/dashboard",
      name: "dashboard",
      component: Dashboard,
    },
  ],
});
export default router;
