import 'jest';
import {
  describe,
  expect,
  test,
  beforeEach,
} from '@jest/globals';
import { store } from '../../../../store/index';
import { default as devices } from '../../../../../../docs/devices.json';
import { HubState } from '@/types/settings';

describe('hubState store', () => {
  let hubState: HubState;

  beforeEach(() => {
    hubState = createHubState();
  });

  test('should initialize with default values', () => {
    expect(hubState.isConnected).toBe(false);
    expect(hubState.error).toBeNull();
    expect(hubState.loading).toBe(false);
  });

  test('setConnected should update connection state', () => {
    hubState.setConnected(true);
    expect(hubState.isConnected).toBe(true);

    hubState.setConnected(false);
    expect(hubState.isConnected).toBe(false);
  });

  test('setError should update error state', () => {
    const testError = new Error('Test error');
    hubState.setError(testError);
    expect(hubState.error).toBe(testError);

    hubState.setError(null);
    expect(hubState.error).toBeNull();
  });

  test('setLoading should update loading state', () => {
    hubState.setLoading(true);
    expect(hubState.loading).toBe(true);

    hubState.setLoading(false);
    expect(hubState.loading).toBe(false);
  });

  test('reset should restore initial state', () => {
    // Set some values first
    hubState.setConnected(true);
    hubState.setError(new Error('Test error'));
    hubState.setLoading(true);

    // Reset the state
    hubState.reset();

    // Verify reset values
    expect(hubState.isConnected).toBe(false);
    expect(hubState.error).toBeNull();
    expect(hubState.loading).toBe(false);
  });
});
