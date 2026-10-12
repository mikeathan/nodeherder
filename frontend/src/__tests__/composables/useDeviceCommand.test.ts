import { afterEach, beforeEach, describe, expect, jest, test } from '@jest/globals';
import { store } from '@/store';
import { useDeviceCommand } from '@/composables/useDeviceCommand';
import { device, expose, hubConfig } from '../helpers/devices';
import { captureWsFrames } from '../helpers/ws-capture';

// Investigation characterizations of #54, not a claim that timeout-on-suppressed-echo is correct.
// Uses the real command composable, hub mutation and WebSocket-emitting action with a fake socket.
describe('brightness confirmation investigation (AC-42)', () => {
  const light = device('investigation-light', {}, [expose('brightness', { data: 50 })]);
  let capture: ReturnType<typeof captureWsFrames>;

  beforeEach(() => {
    jest.useFakeTimers();
    store.dispatch('hub/init', { devices: [light], config: hubConfig() });
    store.commit('ws/setConnectionStatus', 'connected');
    store.commit('alerts/clear');
    capture = captureWsFrames();
  });

  afterEach(() => {
    jest.runOnlyPendingTimers();
    capture.restore();
    store.commit('alerts/clear');
    jest.useRealTimers();
  });

  function echo(value: number) {
    store.commit('hub/updateDevice', { id: light.id, data: { brightness: value } });
  }

  test('second command is sent, but a suppressed echo makes it time out and blocks another command while pending', () => {
    const command = useDeviceCommand();
    expect(command.send(light, 'brightness', 100)).toBe(true);
    echo(100);
    expect(command.isPending(light.id, 'brightness')).toBe(false);

    expect(command.send(light, 'brightness', 150)).toBe(true);
    expect(capture.frames.map((f) => f.payload)).toEqual([
      { id: light.id, name: 'brightness', value: 100 },
      { id: light.id, name: 'brightness', value: 150 },
    ]);
    expect(command.send(light, 'brightness', 200)).toBe(false);
    expect(capture.frames).toHaveLength(2);
    jest.advanceTimersByTime(4999);
    expect(command.isPending(light.id, 'brightness')).toBe(true);
    jest.advanceTimersByTime(1);
    expect(command.isPending(light.id, 'brightness')).toBe(false);
    expect(store.getters['alerts/messages']()).toEqual(expect.arrayContaining([
      expect.objectContaining({ message: expect.stringContaining('did not confirm the change') }),
    ]));
    expect(store.getters['hub/findDevice'](light.id).exposes.brightness.data).toBe(100);

    expect(command.send(light, 'brightness', 220)).toBe(true);
    echo(220);
    expect(command.isPending(light.id, 'brightness')).toBe(false);
    expect(capture.frames).toHaveLength(3);
  });

  test('three promptly echoed changes all settle without waiting between commands', () => {
    const command = useDeviceCommand();
    for (const value of [100, 150, 200]) {
      expect(command.send(light, 'brightness', value)).toBe(true);
      echo(value);
      expect(command.isPending(light.id, 'brightness')).toBe(false);
    }
    expect(capture.frames).toHaveLength(3);
    jest.advanceTimersByTime(5000);
    expect(store.getters['alerts/messages']()).toEqual([]);
  });
});
