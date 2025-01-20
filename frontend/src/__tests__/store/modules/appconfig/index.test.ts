import 'jest';
import {
  describe,
  expect,
  test,
  beforeEach,
} from '@jest/globals';
import { store } from '../../../../store/index';
import {
  AppConfig,
  DeviceSettings,
  LoggerSettingsType,
} from '@/types/settings.type';

const mockAppconfig: AppConfig = {
  history: {
    sleepTimeout: { value: 1, unit: 'hours' },
    expireAt: { value: 10, unit: 'days' },
  },
  logger: {
    enableRemoteLogger: false,
  },
  devices: {
    x01234: {
      id: 'x01234',
      disabled: true,
      metricsEnabled: false,
      rateLimit: 50000,
    },
    x111111: {
      id: 'x111111',
      disabled: false,
      metricsEnabled: false,
      rateLimit: 500000000,
    },
    x2222222: {
      id: 'x2222222',
      disabled: false,
      metricsEnabled: true,
      rateLimit: 1000000000,
    },
  },
};

describe('test appconfig module', () => {
  beforeEach(() => {
    store.commit('appconfig/clear');
  });

  test('test appconfig gets initialized', () => {
    var result = store.getters[
      'appconfig/initialized'
    ]() as boolean;
    expect(result).toEqual(false);

    store.dispatch('appconfig/init', mockAppconfig);
    var result = store.getters[
      'appconfig/initialized'
    ]() as boolean;
    expect(result).toEqual(true);

    Object.values(mockAppconfig.devices).forEach(
      (value) => {
        const deviceSetting = store.getters[
          'appconfig/findDeviceSetting'
        ]((value as DeviceSettings).id) as DeviceSettings;
        expect(value).toEqual(deviceSetting);
      }
    );
  });

  test('test save device settigs saves the device settigs changes', () => {
    store.dispatch('appconfig/init', mockAppconfig);

    const dev = mockAppconfig.devices['x2222222'];

    dev.disabled = true;
    dev.metricsEnabled = false;
    dev.rateLimit = 66666666;

    store.commit('appconfig/setDeviceSetting', dev);

    var deviceSetting = store.getters[
      'appconfig/findDeviceSetting'
    ]('x2222222') as DeviceSettings;

    expect(deviceSetting.id).toEqual('x2222222');
    expect(deviceSetting.disabled).toEqual(true);
    expect(deviceSetting.metricsEnabled).toEqual(false);
    expect(deviceSetting.rateLimit).toEqual(66666666);
  });

  test('test clear device settings, clears the device settings', () => {
    var result = store.getters[
      'appconfig/initialized'
    ]() as boolean;
    expect(result).toEqual(false);
    store.dispatch('appconfig/init', mockAppconfig);

    var result = store.getters[
      'appconfig/initialized'
    ]() as boolean;
    expect(result).toEqual(true);

    store.commit('appconfig/clear');

    var result = store.getters[
      'appconfig/initialized'
    ]() as boolean;
    expect(result).toEqual(false);

    var deviceSetting = store.getters[
      'appconfig/findDeviceSetting'
    ]('x2222222') as DeviceSettings;

    expect(deviceSetting).toBeUndefined();
  });

  test('test save logger settigs saves the logger settings changes', () => {
    store.dispatch('appconfig/init', mockAppconfig);

    mockAppconfig.logger.enableRemoteLogger = true;

    store.commit(
      'appconfig/setLoggerSettings',
      mockAppconfig.logger
    );

    var loggerSetting = store.getters[
      'appconfig/logger'
    ]() as LoggerSettingsType;

    expect(loggerSetting.enableRemoteLogger).toEqual(true);
  });
});
