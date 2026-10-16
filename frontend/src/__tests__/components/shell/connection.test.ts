import { connectionNotices, connectionText } from '@/components/shell/connection';

test.each(['idle', 'connecting', 'connected'] as const)('%s never creates a transport warning', (status) => {
  expect(connectionNotices(status, 'checking', 'loading')).toEqual([]);
  expect(connectionText(status)).toBe(status === 'connected' ? 'Connected' : 'Connecting…');
});
test.each(['reconnecting', 'disconnected'] as const)('%s reports actual loss and offers retry', (status) => {
  expect(connectionNotices(status, 'authenticated', 'ready')).toEqual([
    expect.objectContaining({ kind: 'transport', retry: 'Reconnect' }),
  ]);
  expect(connectionText(status)).toContain(status === 'reconnecting' ? 'Reconnecting' : 'Disconnected');
});
test('session, transport and data failures remain distinct and clear independently', () => {
  expect(connectionNotices('connected', 'unavailable', 'failed').map((notice) => notice.kind)).toEqual(['session', 'snapshot']);
  expect(connectionNotices('disconnected', 'unavailable', 'failed')).toHaveLength(3);
  expect(connectionNotices('connected', 'authenticated', 'ready')).toEqual([]);
});
