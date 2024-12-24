const maxNumberOfAttempts = 10;
const intervalTimeMs = 200;

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

function createSocket(url?: string): WebSocket {
  return new WebSocket(url ?? getSocketUri());
}

class WsClient {
  private ws: WebSocket;

  constructor(ws: WebSocket) {
    this.ws = ws;
  }

  public emit(event: string, message: string) {
    var payload = JSON.stringify({
      type: event,
      payload: message,
    });

    this.ws.send(payload);

    // if (this.ws.readyState !== this.ws.OPEN) {
    //   let currentAttempt = 0;
    //   const interval = setInterval(() => {
    //     if (currentAttempt > maxNumberOfAttempts - 1) {
    //       clearInterval(interval);
    //       console.log(
    //         'emit:',
    //         event,
    //         ' failed. Maximum number of attempts exceeded.',
    //       );
    //       return;
    //     } else if (this.ws.readyState === this.ws.OPEN) {
    //       clearInterval(interval);
    //       this.ws.send(payload);
    //     }
    //     currentAttempt++;
    //   }, intervalTimeMs);
    // } else {
    //   this.ws.send(payload);
    // }
  }
}

export class WsClientService {
  private wsClient!: WsClient;
  private builder: WsClientBuilder;

  constructor() {
    this.builder = WsClientBuilder.create();
  }

  private connectWebSocket() {
    this.wsClient = this.builder.connect();
    console.log('WebSocket connected');
  }

  private client() {
    return this.wsClient ?? this.connectWebSocket();
  }
  public emit(event: string, message: string) {
    this.client().emit(event, message);
  }

  static createFromBuilder(
    builder: WsClientBuilder,
  ): WsClientService {
    var service = new WsClientService();
    service.wsClient = builder.connect();
    return service;
  }
}

export class WsClientBuilder {
  private url: string;
  private onOpen: ((ev: Event) => any) | null;
  private onClose: ((ev: CloseEvent) => any) | null;
  private onError: ((tev: Event) => any) | null;
  private onDisconnected: (() => void) | null;
  private onMessage: ((ev: MessageEvent) => any) | null;
  private reconnectAttempts: number = 0;
  private maxReconnectAttempts = 10;

  private constructor(url: string) {
    this.url = url;
    this.onOpen = null;
    this.onClose = null;
    this.onError = null;
    this.onMessage = null;
    this.onDisconnected = null;
  }

  withOnOpen(
    event: ((ev: Event) => any) | null,
  ): WsClientBuilder {
    this.onOpen = (ev: Event) => {
      this.reconnectAttempts = 0;
      event?.(ev);
    };
    return this;
  }

  withOnClose(
    event: ((ev: CloseEvent) => any) | null,
  ): WsClientBuilder {
    this.onClose = (ev: CloseEvent) => {
      this.reconnectWithBackoff();
      event?.(ev);
    };

    return this;
  }
  withOnDisconnected(event: (() => void) | null) {
    this.onDisconnected = event;
    return this;
  }

  withOnError(
    event: ((ev: Event) => any) | null,
  ): WsClientBuilder {
    this.onError = event;
    return this;
  }

  withOnMessage(
    event: ((ev: MessageEvent) => any) | null,
  ): WsClientBuilder {
    this.onMessage = event;
    return this;
  }

  connect(): WsClient {
    const ws: WebSocket = new WebSocket(this.url);
    ws.onopen = this.onOpen;
    ws.onclose = this.onClose;
    ws.onerror = this.onError;
    ws.onmessage = this.onMessage;

    return new WsClient(ws);
  }

  private reconnectWithBackoff() {
    if (
      this.reconnectAttempts < this.maxReconnectAttempts
    ) {
      const backoffDelay = Math.min(
        1000 * Math.pow(2, this.reconnectAttempts), // Exponential backoff: 1s, 2s, 4s, etc.
        30000, // Cap the delay at 30 seconds
      );
      console.log(
        `(${this.reconnectAttempts}/${
          this.maxReconnectAttempts
        }) Reconnecting in ${
          backoffDelay / 1000
        } seconds... `,
      );
      setTimeout(() => {
        this.reconnectAttempts += 1;
        this.connect();
      }, backoffDelay);
    } else {
      console.error('Max reconnect attempts reached');
      this.onDisconnected?.();
    }
  }

  static create(): WsClientBuilder {
    const url = getSocketUri();
    return new WsClientBuilder(url);
  }
}
