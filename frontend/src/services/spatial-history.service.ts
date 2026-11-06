import { get } from '@/contracts/api';

export type SpatialSession = {
  id: string;
  from: number;
  to: number;
  points: number;
  slots: number[];
  capped: boolean;
};
export type SpatialPoint = { at: number; slot: number; x: number; y: number };
export type SpatialHeatCell = { x: number; y: number; count: number };

async function query<T>(params: Record<string, string | number>): Promise<T> {
  const search = new URLSearchParams(Object.entries(params).map(([key, value]) => [key, String(value)]));
  const response = await get(`/spatial/history?${search}`);
  if (!response.ok) throw new Error(response.status === 503 ? 'Spatial history is unavailable.' : 'Spatial history could not be loaded.');
  return (await response.json()) as T;
}

export const spatialHistoryApi = {
  sessions: (device: string, from: number, to: number) =>
    query<SpatialSession[]>({ device, from, to, mode: 'sessions' }),
  points: (device: string, session: string, from: number, to: number) =>
    query<SpatialPoint[]>({ device, session, from, to, mode: 'points' }),
  heatmap: (device: string, from: number, to: number, cell = 250) =>
    query<SpatialHeatCell[]>({ device, from, to, mode: 'heatmap', cell }),
};
