import { describe, expect, test } from '@jest/globals';
import { connectionTransition } from '@/domain/notifications';

describe('connection notifications', () => {
  test('a live link dropping becomes one loss event', () => {
    expect(connectionTransition('connected', 'reconnecting')).toBe('lost');
    expect(connectionTransition('connected', 'disconnected')).toBe('lost');
    expect(connectionTransition('reconnecting', 'disconnected')).toBeNull();
  });

  test('only a reconnect after a drop becomes a restore event', () => {
    expect(connectionTransition('reconnecting', 'connected')).toBe('restored');
    expect(connectionTransition('disconnected', 'connected')).toBe('restored');
    expect(connectionTransition('connecting', 'connected')).toBeNull();
    expect(connectionTransition('connected', 'connected')).toBeNull();
  });
});
