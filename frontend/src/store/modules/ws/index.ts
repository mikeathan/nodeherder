import { Module } from 'vuex';
import { RootState } from '../../state';
import { WSClientState } from './state';
import { ConnectionStatus } from '@/types/connection.type';

function getSocketUri() {
  const devSocketUri = 'ws://localhost:3000/ws';
  const productionSocketUri =
    'ws://' + document.location.host + '/ws';

  if (process.env.NODE_ENV == 'development') {
    console.info('Enviroment:', process.env.NODE_ENV);
    return devSocketUri;
  }

  return productionSocketUri;
}
export const WSClientModule: Module<
  WSClientState,
  RootState
> = {
  namespaced: true,

  state: () => ({
    socket: null,
    connectionStatus: 'disconnected',
  }),

  getters: {
    getConnectionStatus: (state) => state.connectionStatus,
  },

  mutations: {
    setSocket(state, socket: WebSocket) {
      state.socket = socket;
    },
    
    setConnected(state, connected: boolean) {
      state.connected = connected;
    },
    setConnectionStatus(state, status: ConnectionStatus) {
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
        console.error(
          'Cannot send message: WebSocket is not open.',
        );
      }
    },
  },

  actions: {
    connect({ state, commit, rootState, dispatch }) {
      const socket = new WebSocket(getSocketUri());
      
      socket.onopen = function (event) {
        console.log('ws connected');
        dispatch('emit', { event: 'loadDevices' });

        commit('setConnected', true);
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
      };
      socket.onclose = function (event) {
        console.log('ws close ', event);
        commit('setConnected', false);
        commit('setConnectionStatus', 'connecting'); // WIP

        // TODO use backoff reconnect
        setTimeout(() => {
          console.log('ws reconnecting');
          dispatch('connect', {});
        }, 2000);
      };

      socket.onerror = function (event) {
        console.error('ws error: ' + event);
      };

      commit('setSocket', socket);
    },

    // connect({ state, commit, rootState, dispatch }) {
    //   const initializeWebSocket = () => {
    //     const builder = WsClientBuilder.create();

    //     builder.withOnMessage((event) => {
    //       if (event == undefined) {
    //         console.error('ws undefined event: ' + event);
    //         return;
    //       }
    //       if (event.data == undefined) {
    //         console.error(
    //           'ws undefined data: ' + event.data,
    //         );
    //         return;
    //       }

    //       const obj = JSON.parse(event.data);
    //       switch (obj.type) {
    //         case 'deviceUpdated':
    //           commit('devices/update', obj.payload, {
    //             root: true,
    //           });
    //           break;
    //         case 'deviceAdded':
    //           commit('devices/add', obj.payload, {
    //             root: true,
    //           });
    //           break;
    //         case 'automations':
    //           dispatch('automations/init', obj.payload, {
    //             root: true,
    //           });
    //           break;
    //         case 'automationUpdated':
    //           commit('automations/update', obj.payload, {
    //             root: true,
    //           });
    //           break;
    //         case 'devices':
    //           dispatch('devices/init', obj.payload, {
    //             root: true,
    //           });
    //           break;
    //         case 'deviceList':
    //           dispatch('devices/updateItems', obj.payload, {
    //             root: true,
    //           });
    //           break;
    //         case 'appConfig':
    //           dispatch('appconfig/init', obj.payload, {
    //             root: true,
    //           });
    //           break;
    //         case 'metrics':
    //           dispatch('metrics/store', obj.payload, {
    //             root: true,
    //           });
    //           break;
    //         case 'logger':
    //           dispatch('console/addMessage', obj.payload, {
    //             root: true,
    //           });
    //           break;
    //         case 'operationSuccess':
    //           dispatch(
    //             'alerts/showSuccess',
    //             'Operation was successful.',
    //             {
    //               root: true,
    //             },
    //           );
    //           break;
    //         case 'operationFailed':
    //           dispatch('alerts/showError', obj.payload, {
    //             root: true,
    //           });
    //           break;
    //         default:
    //           console.error(
    //             'ws unhandled type: ',
    //             event.data,
    //           );
    //       }
    //     });

    //     builder.withOnReconnected(function () {
    //       console.info('ws reconnected');
    //       initializeWebSocket();
    //     });

    //     builder.withOnOpen(function (event) {
    //       commit('setConnected', true);
    //       dispatch('emit', { event: 'loadDevices' });
    //       commit('setConnectionStatus', 'connected');
    //     });

    //     builder.withOnClose(function (event) {
    //       console.info('ws close ', event);
    //       commit('setConnected', false);
    //       commit('setConnectionStatus', 'connecting'); // WIP
    // commit('setConnected', false);
    //       commit('setConnectionStatus', 'connecting'); // WIP
    //     });

    //     builder.withOnDisconnected(function () {
    //       console.info('ws disconnected');

    //       dispatch('cleanup', [], { root: true });
    //       commit('setConnectionStatus', 'disconnected'); // WIP
    //     });

    //     builder.withOnError(function (event) {
    //       console.error('ws error: ' + event);
    //     });

    //     const ws =
    //       WsClientService.createFromBuilder(builder);
    //     commit('setWs', ws);
    //   };

    //   initializeWebSocket();
    // },

    emit({ commit }, { event, message }) {
      commit('sendMessage', {
        event: event,
        message: message,
      });
    },
  },
};
