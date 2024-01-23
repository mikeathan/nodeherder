import { createApp } from "vue";
import App from "./App.vue";
import router from "./router";
import Notifications from "@kyvg/vue3-notification";
import "@fortawesome/fontawesome-free/css/all.css";
import "bootstrap/dist/css/bootstrap.min.css";
import "bootstrap";
import { store, key } from "./store/index";

//import "./assets/css/styles.global.css";
import "./assets/css/dark.css";

const app = createApp(App);
app.use(store, key);
app.use(router);
app.use(Notifications);
app.mount("#app");
