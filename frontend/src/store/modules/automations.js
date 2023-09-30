const state = {
  initialized: false,
  items: {},
};

const getters = {
  items: (state) => state.items,
  isInitialized: (state) => state.initialized,
};

const actions = {};

const mutations = {
  init(state, items) {
    items.forEach((item) => {
      state.items[item.name] = item;
    });
    state.initialized = true;
  },

  clear(state) {
    state.initialized = false;
    console.log("clear automations");
    for (var prop in state.items) {
      if (state.items.hasOwnProperty(prop)) {
        console.log("deleting:" + prop);
        delete state.item[prop];
      }
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
