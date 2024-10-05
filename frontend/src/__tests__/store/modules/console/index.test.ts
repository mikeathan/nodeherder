import 'jest';
import {
  describe,
  expect,
  test,
  beforeEach,
} from '@jest/globals';
import { store } from '../../../../store/index';
import { LogMessageType } from '@/types/event-logs.type';

const mockMessage: LogMessageType = {
  level: 'info',
  message: 'test message',
  timestamp: Date.now(),
};

describe('test console module', () => {
  beforeEach(() => {
    store.commit('console/clear');
  });

  test('test console message can be added', () => {
    store.dispatch('console/addMessage', mockMessage);

    store.getters['console/messages']().forEach(
      (message: LogMessageType) => {
        expect(message).toEqual(mockMessage);
      },
    );
    // Object.values(mockAppconfig.devices).forEach(
    //   (value) => {
    //     const deviceSetting = store.getters[
    //       'appconfig/findDeviceSetting'
    //     ]((value as DeviceSettings).id) as DeviceSettings;
    //     expect(value).toEqual(deviceSetting);
    //   },
    //);
  });
});
