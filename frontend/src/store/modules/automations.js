const state = {
  initialized: false,
  items: {},
};

const actions = {
  save({ state, dispatch, rootState }, id) {
    console.log("automations/save");
    dispatch(
      "ws/emit",
      { event: "saveAutomation", message: state.items[id] },
      { root: true }
    );
  },

  deleteTrigger({ state, dispatch, automationId, triggerId }) {
    console.log("automations/deleteTrigger", automationId, "=", triggerId);
    var automation = state.items[automationId];
    if (automation == undefined) {
      console.log("automation id ", automationId, " not found");
      return
    }

    // TODO:
    // delete trigger
  },
};

const getters = {
  items: (state) => state.items,
  isInitialized: (state) => state.initialized,
  find: (state) => (id) => {
    return state.items[id];
  },
};

const mutations = {
  init(state, items) {
    console.log("automations/init");
    items.forEach((item) => {
      state.items[item.id] = item;
    });
    state.initialized = true;
  },

  clear(state) {
    console.log("clear automations");
    for (var prop in state.items) {
      if (state.items.hasOwnProperty(prop)) {
        delete state.items[prop];
      }
    }
    state.initialized = false;
  },
};

export default {
  namespaced: true,
  state,
  mutations,
  actions,
  getters,
};
