import 'jest';
import {
  describe,
  expect,
  test,
  beforeEach,
} from '@jest/globals';
import { store } from '../../../../store/index';

describe('test console module', () => {
  // beforeEach(() => {
  //   store.commit('appconfig/clear');
  // });

  test('test console messages are added', () => {
    var result = store.getters[
      'console/initialized'
    ]() as boolean;

    store.dispatch('console/add', mockAppconfig);

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
