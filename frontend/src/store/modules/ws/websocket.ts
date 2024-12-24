import { reactive, onMounted, onUnmounted } from 'vue';

interface WebSocketState {
  socket: WebSocket | null;
  isConnected: boolean;
  message: string;
}

const BACKOFF_DELAYS = [1000, 3000, 5000];

export function useWebSocket(url: string) {
  const state = reactive<WebSocketState>({
    socket: null,
    isConnected: false,
    message: '',
  });

  let retryCount = 0;
  let reconnectTimeout: ReturnType<typeof setTimeout> | null = null;

  const connect = () => {
    if (state.socket && state.socket.readyState === WebSocket.OPEN) {
      return;
    }

    state.socket = new WebSocket(url);

    state.socket.onopen = () => {
      console.log('WebSocket connected.');
      state.isConnected = true;
      retryCount = 0; // Reset retry count on successful connection
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
        reconnectTimeout = null;
      }
    };

    state.socket.onmessage = (event) => {
      state.message = event.data as string;
    };

    state.socket.onerror = (error) => {
      console.error('WebSocket error:', error);
      state.isConnected = false;
      reconnect();
    };

    state.socket.onclose = () => {
      console.log('WebSocket connection closed.');
      state.isConnected = false;
      reconnect();
    };
  };

  const reconnect = () => {
      if (reconnectTimeout) return;
      const delay = BACKOFF_DELAYS[Math.min(retryCount, BACKOFF_DELAYS.length - 1)];
      console.log(`Reconnecting in ${delay / 1000} seconds...`);
      retryCount++;
      reconnectTimeout = setTimeout(() => {
        connect();
        reconnectTimeout = null;
      }, delay);
  }

  const sendMessage = (message: string) => {
    if (state.socket && state.isConnected) {
      state.socket.send(message);
    } else {
      console.warn('Cannot send message. Not connected.');
    }
  };

  const disconnect = () => {
    if (state.socket) {
      state.socket.close();
      state.socket = null;
      state.isConnected = false;
      if (reconnectTimeout) {
        clearTimeout(reconnectTimeout);
        reconnectTimeout = null;
      }
    }
  };

  onMounted(() => {
    connect();
  });

  onUnmounted(() => {
    disconnect();
  });

  return { state, sendMessage, disconnect };
}