import { Module } from "vuex";
import { RootState } from "../../state";
import { AutomationModuleState } from "./state";
import {
  AutomationMap,
  Automation,
  Automations,
} from "../../../types/automation";

export const AutomationModule: Module<AutomationModuleState, RootState> = {
  namespaced: true,

  state: () => ({ automationsMap: {} as AutomationMap, initialized: false }),

  getters: {
    list: (state: AutomationModuleState) => state.automationsMap,
    initialized: (state: AutomationModuleState) => state.initialized,
    find: (state: AutomationModuleState) => (id: string) => {
      return state.automationsMap[id];
    },
  },

  mutations: {
    add(state: AutomationModuleState, automation: Automation) {
      state.automationsMap[automation.id] = automation;
    },
    update(state: AutomationModuleState, automation: Automation) {
      state.automationsMap[automation.id] = automation;
    },

    delete(state: AutomationModuleState, id: string) {
      delete state.automationsMap[id];
    },

    clear(state: AutomationModuleState) {
      for (var id in state.automationsMap) {
        delete state.automationsMap[id];
      }
      state.initialized = false;
    },
  },

  actions: {
    init({ state, commit }, automations: Automations) {
      commit("clear", []);

      automations.forEach((item: Automation) => {
        state.automationsMap[item.id] = item;
      });
      state.initialized = true;
    },
    save({ commit, dispatch, rootState }, automation: Automations) {
      commit("add", automation);
      dispatch(
        "ws/emit",
        { event: "saveAutomation", message: automation },
        { root: true }
      );
    },

    delete({ commit, dispatch, rootState }, id: string) {
      commit("delete", id);
      dispatch(
        "ws/emit",
        { event: "deleteAutomation", message: { id: id } },
        { root: true }
      );
    },
  },
};
