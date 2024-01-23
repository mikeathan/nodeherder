import { createApp } from "vue";
import App from "./App.vue";
import store from "./store/store.js";
import router from "./router";
import Notifications from "@kyvg/vue3-notification";
import "@fortawesome/fontawesome-free/css/all.css";
import "bootstrap/dist/css/bootstrap.min.css";
import "bootstrap";
import { store_temp, key } from "./store/index";

//import "./assets/css/styles.global.css";
import "./assets/css/dark.css";

const app = createApp(App);
app.use(store_temp, key);

app.use(store);
app.use(router);
app.use(Notifications);
app.mount("#app");
