import { Module } from 'vuex';
import { RootState } from '../../state';
import { WSClientState } from './state';
import { ConnectionStatus, ConnectionStatusType } from '@/types/connection.type';

function getSocketUri() {
  const devSocketUri = 'ws://localhost:3000/ws';
  const productionSocketUri = 'ws://' + document.location.host + '/ws';

  if (process.env.NODE_ENV == 'development') {
    console.info('Enviroment:', process.env.NODE_ENV);
    return devSocketUri;
  }

  return productionSocketUri;
}
let reconnectAttempts: number = 0;
const maxReconnectAttempts = 10;
const maxReconnectTimeout = 30000; //// Cap the delay at 30 seconds

export const WSClientModule: Module<WSClientState, RootState> = {
  namespaced: true,

  state: () => ({
    socket: null,
    connectionStatus: ConnectionStatus.disconnected,
  }),

  getters: {
    getConnectionStatus: (state) => state.connectionStatus,
  },

  mutations: {
    setSocket(state, socket: WebSocket) {
      state.socket = socket;
    },
    setConnectionStatus(state, status: ConnectionStatusType) {
      state.connectionStatus = status;
    },
    sendMessage(state: WSClientState, { event, message }) {
      var payload = JSON.stringify({
        type: event,
        payload: message,
      });

      if (state.socket?.readyState === WebSocket.OPEN) {
        state.socket.send(payload);
      } else {
        console.error('Cannot send message: WebSocket is not open.');
      }
    },
  },

  actions: {
    connect({ state, commit, rootState, dispatch }) {
      const socket = new WebSocket(getSocketUri());

      socket.onopen = function (event) {
        console.log('ws connected');
        reconnectAttempts = 0;
        dispatch('emit', { event: 'loadDevices' });

        ///////////////////////////////////////////////////////////
        dispatch('emit', { event: 'loadAppConfig' }); // TEMPORARY

        ///////////////////////////////////////////////////////////
        commit('setConnectionStatus', 'connected');
      };

      socket.onmessage = function (event) {
        if (event == undefined) {
          console.error('ws undefined event: ' + event);
          return;
        }
        if (event.data == undefined) {
          console.error('ws undefined data: ' + event.data);
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
            let message = 'Operation was successful.';
            if (obj.payload) {
              message = obj.payload;
            }
            dispatch('alerts/showSuccess', message, {
              root: true,
            });
            break;
          case 'operationFailed':
            dispatch('alerts/showError', obj.payload, {
              root: true,
            });
            break;
          default:
            console.error('ws unhandled type: ', event.data);
        }
      };
      socket.onclose = function (event) {
        console.log('ws close ', event, 'reconnectAttempts:', reconnectAttempts);
        commit('setConnectionStatus', 'connecting'); // WIP

        if (reconnectAttempts < maxReconnectAttempts) {
          const backoffDelay = Math.min(1000 * Math.pow(2, reconnectAttempts), maxReconnectTimeout);
          console.log(
            `(${reconnectAttempts}/${maxReconnectAttempts}) Reconnecting in ${backoffDelay / 1000} seconds... `
          );
          setTimeout(() => {
            reconnectAttempts += 1;
            dispatch('connect', {});
          }, backoffDelay);
        } else {
          console.error('Max reconnect attempts reached');
          commit('setConnectionStatus', 'disconnected');
          reconnectAttempts = 0;
        }
      };

      socket.onerror = function (event) {
        // no need to log errors when we are already connecting
        if (state.connectionStatus === ConnectionStatus.connecting) {
          return;
        }

        console.error('ws error: ' + event);
      };

      commit('setSocket', socket);
    },
    emit({ commit }, { event, message }) {
      commit('sendMessage', {
        event: event,
        message: message,
      });
    },
  },
};
