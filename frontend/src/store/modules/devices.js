const state = {
  items: {},
};

const getters = {
  items: (state) => state.items,
  findDevice: (state) => (id) => {
    return state.items[id];
  },
};

const mutations = {
  // todo: add try/catch
  update(state, payload) {
    var device = state.items[payload.id];
    for (var key in payload.data) {
      device.exposes[key].data = payload.data[key];
    }
    device.properties.last_seen = payload.last_seen;
  },
  updateproperties(state, payload) {
    var device = state.items[payload.id];
    for (var key in payload.data) {
      device.properties[key] = payload.data[key];
    }
    device.properties.last_seen = payload.last_seen;
  },
  add(state, device) {
    state.items[device.id] = device;
  },
  init(state, devices) {
    devices.forEach((device) => {
      state.items[device.id] = device;
    });
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
  mutations,
};
