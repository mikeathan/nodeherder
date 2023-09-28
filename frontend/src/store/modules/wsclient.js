const socketUri = "ws://localhost:3000/ws"; // used for testing
//const socketUri = "ws://" + document.location.host + "/ws";

const state = {
  ws: null,
  connected: false,
  messageQueue: [],
};

const getters = {
  isconnected: (state) => state.connected,
};

const actions = {
  connect({ state, commit, rootState }) {
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
        case "connected":
          commit("init", obj.payload, { root: true });
          break;
        case "deviceUpdated":
          commit("deviceUpdated", obj.payload, { root: true });
          break;
        case "automations":
          commit("initAutomations", obj.payload, { root: true });
          break;
        default:
          console.log("ws unhandled type: ", event.data);
      }
    };

    ws.onopen = function (event) {
      console.log("ws open");
      while (state.messageQueue.length > 0) {
        ws.send(state.messageQueue.pop());
      }
    };
    ws.onclose = function (event) {
      console.log("ws close");
      state.connected = false;
    };
    ws.onerror = function (event) {
      console.log("ws error: " + event.data);
    };

    console.log("ws connect");
    state.ws = ws;
    state.connected = true;
  },

  emit({ state }, event, message = "") {
    var payload = JSON.stringify({ type: event, payload: message });
    console.log("ws emit");
    if (state.ws.readyState !== 1) {
      state.messageQueue.push(payload);
    } else {
      state.ws.send(payload);
    }
  },
};

const mutations = {};

export default {
  namespaced: true,
  actions,
  state,
  mutations,
  getters,
};
