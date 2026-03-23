import 'jest';
import { describe, expect, test, beforeEach } from '@jest/globals';
import { store } from '@/store/index';
import { AppConfig, DeviceConfig, LoggerSettingsType } from '@/types/settings.type';

const mockAppconfig: AppConfig = {
  hub: {
    history: {
      sleepTimeout: { value: 1, unit: 'hours' },
      expireAt: { value: 10, unit: 'days' },
    },
    logger: {
      enableRemoteLogger: false,
      level: 'info',
    },
    mcp: {
      enabled: false,
    },
    assistant: {
      url: 'http://localhost:4001',
    },
    dashboardGroups: {},
    devices: {
      defaults: {
        id: 'defaults',
        disabled: false,
        metricsEnabled: false,
        defaultDebounceByCategory: {},
        debounceOverrides: {},
      },
      overrides: {
        x01234: {
          id: 'x01234',
          disabled: true,
          metricsEnabled: false,
          defaultDebounceByCategory: {},
          debounceOverrides: {},
        },
        x111111: {
          id: 'x111111',
          disabled: false,
          metricsEnabled: false,
          defaultDebounceByCategory: {},
          debounceOverrides: {},
        },
        x2222222: {
          id: 'x2222222',
          disabled: false,
          metricsEnabled: true,
          defaultDebounceByCategory: {},
          debounceOverrides: {},
        },
      },
    },
  },
  bridge: {
    permitJoin: false,
    maxTimeAllowed: { value: 254, unit: 'seconds' },
  },
};

describe('test appconfig module', () => {
  beforeEach(() => {
    store.commit('hub/clear');
  });

  test('test appconfig gets initialized', () => {
    var result1 = store.getters['hub/isInitialized']() as boolean;
    expect(result1).toEqual(false);

    store.dispatch('hub/init', { devices: [], config: mockAppconfig });
    var result2 = store.getters['hub/isInitialized']() as boolean;
    expect(result2).toEqual(true);

    Object.values(mockAppconfig.hub.devices.overrides).forEach((value) => {
      const deviceSetting = store.getters['hub/findDeviceSetting']((value as DeviceConfig).id) as DeviceConfig;
      expect(value).toEqual(deviceSetting);
    });
  });

  test('test save device settings saves the device settings changes', () => {
    store.dispatch('hub/init', { devices: [], config: mockAppconfig });

    const dev = mockAppconfig.hub.devices.overrides['x2222222'];

    dev.disabled = true;
    dev.metricsEnabled = false;

    store.commit('hub/setDeviceConfigOverride', dev);

    var deviceSetting = store.getters['hub/findDeviceSetting']('x2222222') as DeviceConfig;

    expect(deviceSetting.id).toEqual('x2222222');
    expect(deviceSetting.disabled).toEqual(true);
    expect(deviceSetting.metricsEnabled).toEqual(false);
  });

  test('test clear device settings, clears the device settings', () => {
    var result1 = store.getters['hub/isInitialized']() as boolean;
    expect(result1).toEqual(false);
    store.dispatch('hub/init', { devices: [], config: mockAppconfig });

    var result2 = store.getters['hub/isInitialized']() as boolean;
    expect(result2).toEqual(true);

    store.commit('hub/clear');

    var result3 = store.getters['hub/isInitialized']() as boolean;
    expect(result3).toEqual(false);

    var deviceSetting = store.getters['hub/findDeviceSetting']('x2222222') as DeviceConfig;

    expect(deviceSetting.id).toEqual('');
  });

  test('test save logger settings saves the logger settings changes', () => {
    store.dispatch('hub/init', { devices: [], config: mockAppconfig });

    mockAppconfig.hub.logger.enableRemoteLogger = true;
    mockAppconfig.hub.logger.level = 'debug';

    store.commit('hub/setLoggerSettings', mockAppconfig.hub.logger);

    var loggerSetting = store.getters['hub/logger']() as LoggerSettingsType;

    expect(loggerSetting.enableRemoteLogger).toEqual(true);
    expect(loggerSetting.level).toEqual('debug');
  });
});
