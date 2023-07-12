import { createRouter, createWebHistory } from "vue-router";
import DevicePage from "../components/device/DevicePage.vue";
import Dashboard from "../components/device-dashboard/DeviceDashboard.vue";
const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: "/",
      name: "home",
      component: Dashboard,
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
