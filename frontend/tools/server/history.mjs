import { getMetricsForDevice } from './metrics-mock.mjs';

function samples(device, expose, from, to) {
  const original = getMetricsForDevice(device.id, expose.name, from, to);
  if (original)
    return original.data.map((p) => ({ timestamp: p.x ?? p.timestamp, value: sampleValue(p.y ?? p.value, expose) }));
  return Array.from({ length: 25 }, (_, i) => ({
    timestamp: from + Math.round(((to - from) * i) / 24),
    value: fallbackValue(expose, i),
  }));
}
function sampleValue(value, expose) {
  if (expose.type !== 'binary') return value;
  if (typeof value === 'boolean') return value;
  return ['true', 'ON', 'on', '1'].includes(String(value));
}
function fallbackValue(expose, index) {
  if (expose.type === 'numeric') return Number(expose.data ?? 0) + Math.round(Math.sin(index / 5) * 10) / 10;
  if (expose.type === 'binary') return index % 5 === 0;
  return index % 4 === 0 ? 'double_press' : 'single_press';
}
export function historyResponse(body, devices) {
  const from = Date.parse(body.time?.from);
  const to = Date.parse(body.time?.to);
  if (!Number.isFinite(from) || !Number.isFinite(to) || to <= from) throw new Error('Invalid range');
  const device = devices.find((d) => d.id === body.deviceIds?.[0]);
  if (!device) throw new Error('Missing device');
  return body.exposes.flatMap((name) => {
    const expose = device.exposes[name];
    if (!expose) return [];
    const data = samples(device, expose, from, to);
    return [
      {
        expose: name,
        query_start_timestamp: from,
        query_end_timestamp: to,
        values: [
          {
            device_id: device.id,
            unit: expose.unit,
            timestamp: 0,
            value: {
              name,
              type: expose.type,
              from,
              to,
              data: data.slice(0, body.limit),
              partial: data.length > body.limit,
            },
          },
        ],
      },
    ];
  });
}
export function registerHistoryRoute(app, getDevices) {
  app.post('/api/metrics/query', (req, res) => {
    if (req.body.history !== true) {
      res.status(400).json({ error: 'Mock supports opt-in history' });
      return;
    }
    try {
      res.json(historyResponse(req.body, getDevices()));
    } catch {
      res.status(400).json({ error: 'Invalid history query' });
    }
  });
}
