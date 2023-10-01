const socketUri = "ws://localhost:3000/ws"; // used for testing
//const socketUri = "ws://" + document.location.host + "/ws";
const maxNumberOfAttempts = 10;
const intervalTimeMs = 200;

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
        case "connected": // TODO: remove !!!!!
          break;
        case "deviceUpdated":
          commit("devices/update", obj.payload, { root: true });
          break;
        case "automations":
          commit("automations/init", obj.payload, { root: true });
          break;
        case "devices":
          commit("devices/init", obj.payload, { root: true });
          break;
        case "ping":
          dispatch("emit", "pong");
          break;
        default:
          console.log("ws unhandled type: ", event.data);
      }
    };

    ws.onopen = function (event) {
      console.log("ws open");
      dispatch("emit", "loadDevices");
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

  emit({ commit, state, dispatch }, event, message = "") {
    commit("sendMessage", event, message);
  },
};

const mutations = {
  sendMessage(state, event, message = "") {
    var payload = JSON.stringify({ type: event, payload: message });

    if (state.ws.readyState !== state.ws.OPEN) {
      console.log("ws emit: ", event, " [PENDING]");

      let currentAttempt = 0;
      const interval = setInterval(() => {
        if (currentAttempt > maxNumberOfAttempts - 1) {
          clearInterval(interval);
          throw new Error("Maximum number of attempts exceeded.");
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
