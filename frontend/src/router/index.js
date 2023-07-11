import { createRouter, createWebHistory } from "vue-router";
import DevicePage from "../components/device/DevicePage.vue";
import Dashboard from "../components/device-dashboard/DeviceDashboard.vue";

const routes = [
  {
    path: "/devicepage",
    name: "DevicePage",
    component: DevicePage,
  },
  {
    path: "/dashboard",
    name: "Cats",
    component: Dashboard,
  },
];

const router = createRouter({ history: createWebHistory(), routes });
export default router;
