import { jest } from '@jest/globals';
import { store } from '@/store/index';

export type SentFrame = { type: string; payload: unknown };

/**
 * Installs a fake open WebSocket on the ws store module and records every frame the
 * store sends. Lets tests assert the exact wire contract of store actions (FE-02).
 */
export function captureWsFrames(): { frames: SentFrame[]; restore: () => void } {
  const frames: SentFrame[] = [];
  const g = globalThis as unknown as { WebSocket?: { OPEN: number } };
  const previous = g.WebSocket;
  if (!g.WebSocket) g.WebSocket = { OPEN: 1 };
  const socket = {
    readyState: g.WebSocket.OPEN,
    send: jest.fn((raw: string) => {
      const parsed = JSON.parse(raw) as { type: string; payload: unknown };
      frames.push({ type: parsed.type, payload: parsed.payload });
    }),
  };
  store.commit('ws/setSocket', socket);
  return {
    frames,
    restore: () => {
      store.commit('ws/setSocket', null);
      g.WebSocket = previous;
    },
  };
}
