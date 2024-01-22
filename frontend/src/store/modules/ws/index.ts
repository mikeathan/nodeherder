import { Module } from "vuex";
import { RootState } from "../../state";
import { WSClientState } from "./state";
import { useNotification } from "@kyvg/vue3-notification";
import { createSocket, sendMessage, getSocketUri, WsClient } from "./ws";
import wsclient from "../wsclient";
const { notify } = useNotification();

export const WSClientModule: Module<WSClientState, RootState> = {
  namespaced: true,

  state: () => ({ ws: new WsClient(), connected: false }),

  getters: { isconnected: (state) => state.connected },

  mutations: {
    sendMessage(state: WSClientState, { event, message }) {
      sendMessage(state.ws, event, message);
    },
  },

  actions: {
    connect({ state, commit, rootState, dispatch }) {
      var ws = createSocket();
      ws.onmessage = (event) => {
        if (event == undefined) {
          console.error("ws undefined event: " + event);
          return;
        }
        if (event.data == undefined) {
          console.error("ws undefined data: " + event.data);
          return;
        }

        const obj = JSON.parse(event.data);

        switch (obj.type) {
          case "deviceUpdated":
            commit("devices/update", obj.payload, { root: true });
            break;
          case "deviceAdded":
            commit("devices/add", obj.payload, { root: true });
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
          case "deviceList":
            dispatch("devices/updateItems", obj.payload, { root: true });
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
          default:
            console.error("ws unhandled type: ", event.data);
        }
      };

      ws.onopen = function (event) {
        console.info("ws open");
        dispatch("emit", { event: "loadDevices" });
      };

      ws.onclose = function (event) {
        console.info("ws close ", event);
        state.connected = false;
        dispatch("cleanup", [], { root: true });
      };

      ws.onerror = function (event) {
        console.error("ws error: " + event);
      };

      state.ws = ws;
      state.connected = true;
    },

    emit({ commit }, { event, message }) {
      commit("sendMessage", { event: event, message: message });
    },
  },
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
