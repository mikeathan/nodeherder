const devSocketUri = "ws://localhost:3000/ws";
const productionSocketUri = "ws://" + document.location.host + "/ws";

import { useNotification } from "@kyvg/vue3-notification";

const { notify } = useNotification();

const maxNumberOfAttempts = 10;
const intervalTimeMs = 200;
var socketUri = getSocketUri()
function getSocketUri() {
  if (process.env.NODE_ENV == "development") {
    console.log("Enviroment:", process.env.NODE_ENV)
    return devSocketUri
  }

  return productionSocketUri;
}

const state = {
  ws: null,
  connected: false,
};

const getters = {
  isconnected: (state) => state.connected,
};

const actions = {
  connect({ state, commit, rootState, dispatch }) {

    var ws = new WebSocket(socketUri);
    ws.onmessage = (event) => {
      if (event == undefined) {
        console.log("ws undefined event: " + event);
        return;
      }
      if (event.data == undefined) {
        console.log("ws undefined data: " + event.data);
        return;
      }

      const obj = JSON.parse(event.data);

      console.log("ws message received:", obj.type);
      switch (obj.type) {
        case "deviceUpdated":
          commit("devices/update", obj.payload, { root: true });
          break;
        case "deviceAdded":
          commit("devices/add", obj.payload, { root: true });
          break;
        case "devicePropertiesUpdated":
          commit("devices/updateproperties", obj.payload, { root: true });
          break;
        case "automations":
          dispatch("automations/init", obj.payload, { root: true });
          break;
        case "automationUpdated":
          commit("automations/update", obj.payload, { root: true });
          break;
        case "devices":
          dispatch("devices/init", obj.payload, { root: true });
          break;
          break;
        case "operationSuccess":
          //https://classic.yarnpkg.com/en/package/@kyvg/vue3-notification
          notify({
            type: "success",
            title: "Operation was successful.",
            duration: 2000,
          });
          break;
        case "operationFailed":
          // https://classic.yarnpkg.com/en/package/@kyvg/vue3-notification
          notify({
            type: "error",
            text: obj.payload,
            duration: 3000,
          });
          break;
        case "ping":
          dispatch("emit", { event: "pong" });
          break;
        default:
          console.log("ws unhandled type: ", event.data);
      }
    };

    ws.onopen = function (event) {
      console.log("ws open");
      dispatch("emit", { event: "loadDevices" });
    };

    ws.onclose = function (event) {
      console.log("ws close");
      state.connected = false;
      dispatch("cleanup", [], { root: true });
    };

    ws.onerror = function (event) {
      console.log("ws error: " + event.data);
    };

    console.log("ws connect");
    state.ws = ws;
    state.connected = true;
  },

  emit({ commit }, { event, message }) {
    commit("sendMessage", { event: event, message: message });
  },
};

const mutations = {
  sendMessage(state, { event, message }) {
    var payload = JSON.stringify({ type: event, payload: message });

    if (state.ws.readyState !== state.ws.OPEN) {
      console.log("ws emit: ", event, " [PENDING]");

      let currentAttempt = 0;
      const interval = setInterval(() => {
        if (currentAttempt > maxNumberOfAttempts - 1) {
          clearInterval(interval);
          console.log(
            "emit:",
            event,
            " failed. Maximum number of attempts exceeded."
          );
          return;
        } else if (state.ws.readyState === state.ws.OPEN) {
          clearInterval(interval);
          console.log("ws emit: ", event);
          state.ws.send(payload);
        }
        currentAttempt++;
      }, intervalTimeMs);
    } else {
      console.log("ws emit: ", event);
      state.ws.send(payload);
    }
  },
};

export default {
  namespaced: true,
  actions,
  state,
  mutations,
  getters,
};

// function retryWithExponentialBackoff(fn, maxAttempts = 5, baseDelayMs = 1000) {
//   let attempt = 1

//   const execute = async () => {
//     try {
//       return await fn()
//     } catch (error) {
//       if (attempt >= maxAttempts) {
//         throw error
//       }

//       const delayMs = baseDelayMs * 2 ** attempt
//       console.log(`Retry attempt ${attempt} after ${delayMs}ms`)
//       await new Promise((resolve) => setTimeout(resolve, delayMs))

//       attempt++
//       return execute()
//     }
//   }

//   return execute()
// }
