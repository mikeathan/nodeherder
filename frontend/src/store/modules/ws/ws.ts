const devSocketUri = 'ws://localhost:3000/ws';
const productionSocketUri = 'ws://' + document.location.host + '/ws';

const maxNumberOfAttempts = 10;
const intervalTimeMs = 200;

function getSocketUri() {
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
    var payload = JSON.stringify({ type: event, payload: message });

    if (this.ws.readyState !== this.ws.OPEN) {
      let currentAttempt = 0;
      const interval = setInterval(() => {
        if (currentAttempt > maxNumberOfAttempts - 1) {
          clearInterval(interval);
          console.log(
            'emit:',
            event,
            ' failed. Maximum number of attempts exceeded.'
          );
          return;
        } else if (this.ws.readyState === this.ws.OPEN) {
          clearInterval(interval);
          this.ws.send(payload);
        }
        currentAttempt++;
      }, intervalTimeMs);
    } else {
      this.ws.send(payload);
    }
  }
}

export class WsClientService {
  private static ws: WsClient;
  constructor() {}

  public emit(event: string, message: string) {
    WsClientService.client().emit(event, message);
  }

  static create(builder: WsClientBuilder): WsClient {
    this.ws = builder.build();
    return this.ws;
  }

  private static client(): WsClient {
    return this.ws ?? createSocket(getSocketUri());
  }
}

export class WsClientBuilder {
  private url: string;
  private onOpen: ((ev: Event) => any) | null;
  private onClose: ((ev: CloseEvent) => any) | null;
  private onError: ((tev: Event) => any) | null;
  private onMessage: ((ev: MessageEvent) => any) | null;

  private constructor(url: string) {
    this.url = url;
    this.onOpen = null;
    this.onClose = null;
    this.onError = null;
    this.onMessage = null;
  }

  withOnOpen(event: ((ev: Event) => any) | null): WsClientBuilder {
    this.onOpen = event;
    return this;
  }

  withOnClose(event: ((ev: CloseEvent) => any) | null): WsClientBuilder {
    this.onClose = event;
    return this;
  }

  withOnError(event: ((ev: Event) => any) | null): WsClientBuilder {
    this.onError = event;
    return this;
  }

  withOnMessage(event: ((ev: MessageEvent) => any) | null): WsClientBuilder {
    this.onMessage = event;
    return this;
  }

  build2(): WsClientService {
    const ws: WebSocket = new WebSocket(this.url);
    ws.onopen = this.onOpen;
    ws.onclose = this.onClose;
    ws.onerror = this.onError;
    ws.onmessage = this.onMessage;
    return new WsClientService();
  }

  build(): WsClient {
    const ws: WebSocket = new WebSocket(this.url);
    ws.onopen = this.onOpen;
    ws.onclose = this.onClose;
    ws.onerror = this.onError;
    ws.onmessage = this.onMessage;

    return new WsClient(ws);
  }

  static create(url?: string): WsClientBuilder {
    return new WsClientBuilder(url ?? getSocketUri());
  }
}

// function retryWithExponentialBackoff(fn, maxAttempts = 5, baseDelayMs = 1000) {
//   let attempt = 1

//   const execute = async () => {
//     try {
//       return await fn()
//     } catch (error) {
//       if (attempt >= maxAttempts) {
//         throw error
//       }

//       const delayMs = baseDelayMs * 2 ** attempt
//       console.log(`Retry attempt ${attempt} after ${delayMs}ms`)
//       await new Promise((resolve) => setTimeout(resolve, delayMs))

//       attempt++
//       return execute()
//     }
//   }

//   return execute()
// }
