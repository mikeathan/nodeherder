const state = {
  items: {},
};

const actions = {
  init({ state, commit }, devices) {
    commit("clear", []);
    devices.forEach((device) => {
      state.items[device.id] = device;
    });
  },
};
const getters = {
  items: (state) => state.items,
  find: (state) => (id) => {
    return state.items[id];
  },
};

const mutations = {
  init({ state, commit }, items) {
    commit("clear", []);

    console.log("automations/init");
    items.forEach((item) => {
      state.items[item.id] = item;
    });
    state.initialized = true;
  },

  update(state, payload) {
    var device = state.items[payload.id];
    if (device == undefined) {
      console.log("device ", payload.id, " not found");
      return;
    }
    for (var key in payload.data) {
      if (device.exposes.hasOwnProperty(key)) {
        device.exposes[key].data = payload.data[key];
      }
    }
    for (var key in payload.properties) {
      if (device.properties.hasOwnProperty(key)) {
        device.properties[key].data = payload.properties[key];
      }
    }
    device.properties.last_seen = payload.last_seen;
  },

  updateproperties(state, payload) {
    var device = state.items[payload.id];
    if (device == undefined) {
      console.log("device ", payload.id, " not found");
      return;
    }

    for (var key in payload.properties) {
      if (device.properties.hasOwnProperty(key)) {
        device.properties[key] = payload.data[key];
      }
    }

    device.properties.last_seen = payload.last_seen;
  },
  add(state, device) {
    state.items[device.id] = device;
  },
  clear(state) {
    console.log("clear devices");
    for (var prop in state.items) {
      if (state.items.hasOwnProperty(prop)) {
        delete state.items[prop];
      }
    }
  },
};

export default {
  namespaced: true,
  state,
  getters,
  actions,
  mutations,
};
