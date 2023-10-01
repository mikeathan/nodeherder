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
  update(state, device) {
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
