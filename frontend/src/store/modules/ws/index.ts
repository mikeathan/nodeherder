import { Module } from 'vuex';
import { RootState } from '../../state';
import { WSClientState } from './state';
import { WsClientService, WsClientBuilder } from './ws';
import { ConnectionStatus } from '@/types/connection.type';

export const WSClientModule: Module<
  WSClientState,
  RootState
> = {
  namespaced: true,

  state: () => ({
    ws: new WsClientService(),
    connected: false,
    connectionStatus: 'disconnected',
  }),

  getters: {
    isconnected: (state) => state.connected,
    getConnectionStatus: (state) => state.connectionStatus,
  },

  mutations: {
    setWs(state, ws: WsClientService) {
      state.ws = ws;
    },
    setConnected(state, connected: boolean) {
      state.connected = connected;
    },
    setConnectionStatus(state, status: ConnectionStatus) {
      state.connectionStatus = status;
    },
    sendMessage(state: WSClientState, { event, message }) {
      state.ws.emit(event, message);
    },
  },

  actions: {
    connect({ state, commit, rootState, dispatch }) {
      const initializeWebSocket = () => {
        const builder = WsClientBuilder.create();

        builder.withOnMessage((event) => {
          if (event == undefined) {
            console.error('ws undefined event: ' + event);
            return;
          }
          if (event.data == undefined) {
            console.error(
              'ws undefined data: ' + event.data,
            );
            return;
          }

          const obj = JSON.parse(event.data);
          switch (obj.type) {
            case 'deviceUpdated':
              commit('devices/update', obj.payload, {
                root: true,
              });
              break;
            case 'deviceAdded':
              commit('devices/add', obj.payload, {
                root: true,
              });
              break;
            case 'automations':
              dispatch('automations/init', obj.payload, {
                root: true,
              });
              break;
            case 'automationUpdated':
              commit('automations/update', obj.payload, {
                root: true,
              });
              break;
            case 'devices':
              dispatch('devices/init', obj.payload, {
                root: true,
              });
              break;
            case 'deviceList':
              dispatch('devices/updateItems', obj.payload, {
                root: true,
              });
              break;
            case 'appConfig':
              dispatch('appconfig/init', obj.payload, {
                root: true,
              });
              break;
            case 'metrics':
              dispatch('metrics/store', obj.payload, {
                root: true,
              });
              break;
            case 'logger':
              dispatch('console/addMessage', obj.payload, {
                root: true,
              });
              break;
            case 'operationSuccess':
              dispatch(
                'alerts/showSuccess',
                'Operation was successful.',
                {
                  root: true,
                },
              );
              break;
            case 'operationFailed':
              dispatch('alerts/showError', obj.payload, {
                root: true,
              });
              break;
            default:
              console.error(
                'ws unhandled type: ',
                event.data,
              );
          }
        });

        builder.withOnReconnected(function () {
          console.info('ws reconnected');
          initializeWebSocket();
        });

        builder.withOnOpen(function (event) {
          commit('setConnected', true);
          dispatch('emit', { event: 'loadDevices' });
          commit('setConnectionStatus', 'connected');
        });

        builder.withOnClose(function (event) {
          console.info('ws close ', event);
          commit('setConnected', false);
          commit('setConnectionStatus', 'connecting'); // WIP
        });

        builder.withOnDisconnected(function () {
          console.info('ws disconnected');

          dispatch('cleanup', [], { root: true });
          commit('setConnectionStatus', 'disconnected'); // WIP
        });

        builder.withOnError(function (event) {
          console.error('ws error: ' + event);
        });

        const ws =
          WsClientService.createFromBuilder(builder);
        commit('setWs', ws);
      };

      initializeWebSocket();
    },

    emit({ commit }, { event, message }) {
      commit('sendMessage', {
        event: event,
        message: message,
      });
    },
  },
};
