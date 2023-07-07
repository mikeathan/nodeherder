import { createApp } from "vue";
import App from "./App.vue";
import store from "./store/store.js";
import "./services/ws.js";
import "@fortawesome/fontawesome-free/css/all.css";
import "bootstrap/dist/css/bootstrap.min.css";
import "bootstrap";

import "./assets/css/styles.global.css";
import "./assets/css/dark.css";

const app = createApp(App);
app.use(store);
app.mount("#app");
