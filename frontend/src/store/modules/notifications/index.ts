import type { Module } from 'vuex';
import type { RootState } from '../../state';
import type { AppNotification } from '@/types/notification.type';
import type { NotificationsModuleState } from './state';
import { createUUID } from '@/utils/unique';

const MAX_ITEMS = 200;
const MAX_AGE_MS = 7 * 24 * 60 * 60 * 1000;

/** Draft accepted by `notify`; identity, time and read state are owned by the store. */
export type NotificationInput = Omit<AppNotification, 'id' | 'at' | 'read'> & { at?: number };

function prune(items: AppNotification[], now: number): AppNotification[] {
  return items.filter((item) => now - item.at <= MAX_AGE_MS).slice(0, MAX_ITEMS);
}

/** Durable browser-local notification inbox. Never drives device commands. */
export const NotificationsModule: Module<NotificationsModuleState, RootState> = {
  namespaced: true,

  state: () => ({
    items: [],
  }),

  getters: {
    /** Newest-first copy for presentation; callers cannot mutate store state. */
    items: (state: NotificationsModuleState) => (): AppNotification[] => [...state.items],
    unreadCount: (state: NotificationsModuleState) => (): number => state.items.filter((item) => !item.read).length,
    /** First `limit` unread items, newest first. */
    unread: (state: NotificationsModuleState) => (limit: number): AppNotification[] =>
      state.items.filter((item) => !item.read).slice(0, Math.max(0, limit)),
  },

  mutations: {
    /**
     * Adds or refreshes one item. A matching unread `dedupeKey` is updated in place so a
     * flapping connection cannot produce hundreds of identical rows; resolved items are
     * allowed to re-open as a new event.
     */
    add(state: NotificationsModuleState, item: AppNotification) {
      const index = item.dedupeKey ? state.items.findIndex((existing) => existing.dedupeKey === item.dedupeKey && !existing.read) : -1;
      if (index >= 0) {
        state.items.splice(index, 1, { ...state.items[index], ...item, id: state.items[index].id });
        return;
      }
      state.items.unshift(item);
      state.items = prune(state.items, Date.now());
    },

    markRead(state: NotificationsModuleState, id: string) {
      const item = state.items.find((candidate) => candidate.id === id);
      if (item) item.read = true;
    },

    markAllRead(state: NotificationsModuleState) {
      for (const item of state.items) item.read = true;
    },

    /** Updates an existing notification even when it was already read. */
    resolveByKey(state: NotificationsModuleState, payload: { dedupeKey: string; patch: Partial<AppNotification> }) {
      const item = state.items.find((candidate) => candidate.dedupeKey === payload.dedupeKey);
      if (item) Object.assign(item, payload.patch);
    },

    remove(state: NotificationsModuleState, id: string) {
      state.items = state.items.filter((item) => item.id !== id);
    },

    clear(state: NotificationsModuleState) {
      state.items = [];
    },
  },

  actions: {
    /** Records one event. Returns the stored item so callers can link to it if needed. */
    notify({ state, commit }, input: NotificationInput): AppNotification {
      const item: AppNotification = {
        ...input,
        id: createUUID(),
        at: input.at ?? Date.now(),
        read: false,
      };
      commit('add', item);
      const stored = state.items.find((candidate) => (candidate.dedupeKey && candidate.dedupeKey === item.dedupeKey) || candidate.id === item.id);
      return stored ?? item;
    },

    /** Resolves the one outage item in place so a restart cannot add a second “restored” row. */
    resolveConnection({ commit }, durationMs: number) {
      const minutes = Math.max(1, Math.round(durationMs / 60_000));
      commit('resolveByKey', {
        dedupeKey: 'hub-connection',
        patch: {
          severity: 'success',
          title: 'Hub connection restored',
          body: `Live updates resumed after ${minutes} min. Values are current again.`,
          action: undefined,
        },
      });
    },

    markRead({ commit }, id: string) {
      commit('markRead', id);
    },

    markAllRead({ commit }) {
      commit('markAllRead');
    },

    clear({ commit }) {
      commit('clear');
    },
  },
};
