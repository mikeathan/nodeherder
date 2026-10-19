export type HistoryPoint = { timestamp: number; value: number | string | boolean | null };
export type HistorySeries = {
  name: string;
  type: 'numeric' | 'binary' | 'enum';
  unit: string;
  from: number;
  to: number;
  data: HistoryPoint[];
  partial: boolean;
};
export type HistoryRequest = { deviceId: string; exposes: string[]; from: number; to: number };
export type HistoryEntry = { status: 'loading' | 'ready' | 'error'; series: HistorySeries[]; error: string };
export const HISTORY_LIMIT = 10000;
export const RENDER_LIMIT = 1000;
export const historyKey = (r: HistoryRequest) =>
  JSON.stringify([r.deviceId, [...new Set(r.exposes)].sort(), r.from, r.to]);
