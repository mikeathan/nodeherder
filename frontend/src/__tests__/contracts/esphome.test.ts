import { describe, expect, test } from '@jest/globals';
import { parseESPHomeEnrollments } from '@/contracts/esphome';

describe('ESPHome enrollment wire contract', () => {
  test('every real connection status survives ingress', () => {
    // The hub's constants are connecting/online/offline; dropping "connecting" at ingress made a
    // device that was actively connecting render as "Not reported".
    const [connecting, online, offline] = parseESPHomeEnrollments({
      data: [
        { id: 'a', address: '192.168.1.2', port: 6053, status: 'connecting' },
        { id: 'b', address: '192.168.1.3', port: 6053, status: 'online' },
        { id: 'c', address: '192.168.1.4', port: 6053, status: 'offline' },
      ],
      error: null,
    });
    expect(connecting.status).toBe('connecting');
    expect(online.status).toBe('online');
    expect(offline.status).toBe('offline');
  });

  test('an unrecognised status is dropped rather than rendered as backend text', () => {
    const [enrollment] = parseESPHomeEnrollments({
      data: [{ id: 'a', address: '192.168.1.2', port: 6053, status: 'something-new', reconnections: -1 }],
      error: null,
    });
    expect(enrollment.status).toBeUndefined();
    expect(enrollment.reconnections).toBeUndefined();
  });
});
