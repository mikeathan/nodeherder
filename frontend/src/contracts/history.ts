import type { HistoryPoint, HistoryRequest, HistorySeries } from '@/domain/history/types';
import { HISTORY_LIMIT } from '@/domain/history/types';

type RecordValue = Record<string, unknown>;
function record(value: unknown): RecordValue {
  if (!value || typeof value !== 'object' || Array.isArray(value)) throw new Error('Invalid history response');
  return value as RecordValue;
}
function array(value: unknown): unknown[] {
  if (!Array.isArray(value)) throw new Error('Invalid history samples');
  return value;
}
function point(value: unknown): HistoryPoint {
  const p = record(value);
  if (typeof p.timestamp !== 'number' || !Number.isFinite(p.timestamp)) throw new Error('Invalid history timestamp');
  const v = p.value;
  if (v === null || typeof v === 'boolean' || typeof v === 'string') return { timestamp: p.timestamp, value: v };
  if (typeof v === 'number' && Number.isFinite(v)) return { timestamp: p.timestamp, value: v };
  throw new Error('Invalid history value');
}
function seriesType(value: unknown): HistorySeries['type'] {
  if (value === 'numeric' || value === 'binary' || value === 'enum') return value;
  throw new Error('Unsupported history type');
}
function typedPoint(value: unknown, kind: HistorySeries['type']): HistoryPoint {
  const result = point(value);
  if (result.value === null) return result;
  const expected = { numeric: 'number', binary: 'boolean', enum: 'string' }[kind];
  if (typeof result.value !== expected) throw new Error('History value does not match its type');
  return result;
}
function series(raw: unknown, unit: unknown, request: HistoryRequest): HistorySeries {
  const metadata = record(raw);
  if (!request.exposes.includes(String(metadata.name))) throw new Error('Unexpected history expose');
  if (metadata.from !== request.from || metadata.to !== request.to) throw new Error('Unexpected history range');
  const kind = seriesType(metadata.type);
  const data = array(metadata.data);
  if (data.length > HISTORY_LIMIT + 1) throw new Error('History response exceeds sample limit');
  const points = data
    .map((rawPoint) => typedPoint(rawPoint, kind))
    .filter((p) => p.timestamp >= request.from && p.timestamp <= request.to);
  points.sort((a, b) => a.timestamp - b.timestamp);
  return {
    name: String(metadata.name),
    type: kind,
    unit: typeof unit === 'string' ? unit : '',
    from: request.from,
    to: request.to,
    data: points.slice(0, HISTORY_LIMIT),
    partial: metadata.partial === true || data.length > HISTORY_LIMIT,
  };
}
function group(raw: unknown, request: HistoryRequest): HistorySeries[] {
  const g = record(raw);
  if (!request.exposes.includes(String(g.expose))) throw new Error('Unexpected history group');
  const values = array(g.values);
  if (values.length > 1) throw new Error('Unexpected history device count');
  return values.map((rawValue) => {
    const v = record(rawValue);
    if (v.device_id !== request.deviceId) throw new Error('Unexpected history device');
    const result = series(v.value, v.unit, request);
    if (result.name !== g.expose) throw new Error('Mismatched history group');
    return result;
  });
}
export function decodeHistory(raw: unknown, request: HistoryRequest): HistorySeries[] {
  const groups = array(raw);
  if (groups.length > 32) throw new Error('Too many history series');
  return groups.flatMap((g) => group(g, request));
}
export function historyBody(request: HistoryRequest) {
  return {
    history: true,
    deviceIds: [request.deviceId],
    exposes: [...new Set(request.exposes)].sort(),
    time: { from: new Date(request.from).toISOString(), to: new Date(request.to).toISOString() },
    aggregation: '',
    filters: [],
    limit: HISTORY_LIMIT + 1,
    sortDesc: false,
  };
}
