import 'jest';
import { describe, expect, test, beforeEach } from '@jest/globals';
import { store } from '../../../../store/index';
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
    store.commit('appconfig/clear');
  });

  test('test appconfig gets initialized', () => {
    var result = store.getters['appconfig/initialized']() as boolean;
    expect(result).toEqual(false);

    store.dispatch('appconfig/init', mockAppconfig);
    var result = store.getters['appconfig/initialized']() as boolean;
    expect(result).toEqual(true);

    Object.values(mockAppconfig.hub.devices.overrides).forEach((value) => {
      const deviceSetting = store.getters['appconfig/findDeviceSetting']((value as DeviceConfig).id) as DeviceConfig;
      expect(value).toEqual(deviceSetting);
    });
  });

  test('test save device settigs saves the device settigs changes', () => {
    store.dispatch('appconfig/init', mockAppconfig);

    const dev = mockAppconfig.hub.devices.overrides['x2222222'];

    dev.disabled = true;
    dev.metricsEnabled = false;

    store.commit('appconfig/setDeviceSetting', dev);

    var deviceSetting = store.getters['appconfig/findDeviceSetting']('x2222222') as DeviceConfig;

    expect(deviceSetting.id).toEqual('x2222222');
    expect(deviceSetting.disabled).toEqual(true);
    expect(deviceSetting.metricsEnabled).toEqual(false);
  });

  test('test clear device settings, clears the device settings', () => {
    var result = store.getters['appconfig/initialized']() as boolean;
    expect(result).toEqual(false);
    store.dispatch('appconfig/init', mockAppconfig);

    var result = store.getters['appconfig/initialized']() as boolean;
    expect(result).toEqual(true);

    store.commit('appconfig/clear');

    var result = store.getters['appconfig/initialized']() as boolean;
    expect(result).toEqual(false);

    var deviceSetting = store.getters['appconfig/findDeviceSetting']('x2222222') as DeviceConfig;

    expect(deviceSetting).toBeUndefined();
  });

  test('test save logger settigs saves the logger settings changes', () => {
    store.dispatch('appconfig/init', mockAppconfig);

    mockAppconfig.hub.logger.enableRemoteLogger = true;
    mockAppconfig.hub.logger.level = 'debug';

    store.commit('appconfig/setLoggerSettings', mockAppconfig.hub.logger);

    var loggerSetting = store.getters['appconfig/logger']() as LoggerSettingsType;

    expect(loggerSetting.enableRemoteLogger).toEqual(true);
    expect(loggerSetting.level).toEqual('debug');
  });
});
