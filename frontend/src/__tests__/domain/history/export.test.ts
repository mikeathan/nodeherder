import { historyCSV, historyExportUnavailable } from '@/domain/history/export';
import type { HistorySeries } from '@/domain/history/types';
const series: HistorySeries = { name: 'x', type: 'numeric', unit: '°C', from: 0, to: 1000, partial: false,
  data: [{ timestamp: 0, value: 0 }, { timestamp: 1000, value: null }] };
test('history export preserves zero, gaps, ISO offsets and quoted external values', () => {
  expect(historyCSV(series, 'history')).toContain('1970-01-01T00:00:00.000Z,0,°C\r\n');
  expect(historyCSV(series, 'history')).toContain('1970-01-01T00:00:01.000Z,,°C');
  expect(historyCSV({ ...series, type: 'enum', data: [{ timestamp: 0, value: 'a,"b' }] }, 'history')).toContain('"a,""b"');
  expect(historyCSV({ ...series, type: 'enum', data: [{ timestamp: 0, value: '=SUM(A1)' }] }, 'history')).toContain("'=SUM(A1)");
});
test('statistics schema retains every aggregate and refuses truncated or empty export', () => {
  expect(historyCSV(series, 'statistics').split('\r\n')[0]).toBe('bucket_start,bucket_end,min,average,max,last,samples,unit');
  expect(historyExportUnavailable({ ...series, data: [] })).toMatch(/No recorded/);
  expect(() => historyCSV({ ...series, partial: true }, 'history')).toThrow(/partial/);
});
