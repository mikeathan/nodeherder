import { createRouter, createWebHistory } from "vue-router";
import DevicePage from "../components/device/DevicePage.vue";
import Dashboard from "../components/device-dashboard/DeviceDashboard.vue";
import TestPage from "../components/Test.vue";
const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      component: Dashboard,
    },
    {
      path: "/devicepage",
      component: DevicePage,
    },
    {
      path: "/dashboard",
      component: Dashboard,
    },
  ],
});
export default router;
