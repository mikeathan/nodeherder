import 'jest';
import { describe, expect, test } from '@jest/globals';
import { default as hubStateObj } from '../../../../docs/hub_state.json';
import { Device } from '../../types/device.d';

test('roundtrip serializing device', () => {
  (hubStateObj.payload as any).devices.forEach((device: any) => {
    var json = JSON.stringify(device);

    const newDevice: Device = JSON.parse(json);
    console.log(
      'comparing device:',
      device['id'],
      '-',
      device['friendly_name']
    );
    isEqualToValueAndNotNull(device, 'id', newDevice.id);
    isEqualToValueAndNotNull(
      device,
      'friendly_name',
      newDevice.friendly_name
    );
    isEqualToValueOrNull(
      device,
      'description',
      newDevice.description
    );
    isEqualToValueOrNull(
      device,
      'power_source',
      newDevice.power_source
    );
    isEqualToValueAndNotNull(
      device,
      'connection_type',
      newDevice.connection_type
    );

    for (const [key, expose] of Object.entries(
      (device['exposes'] as any) || {}
    )) {
      console.log('comparing expose:', key);
      var newExpose = (newDevice.exposes as any)[key];
      isEqualToValueAndNotNull(
        expose,
        'name',
        newExpose.name
      );

      isEqualToValueOrNull(
        expose,
        'description',
        newExpose.description
      );
      isEqualToValueOrNull(expose, 'unit', newExpose.unit);

      isEqualToValueOrNull(expose, 'data', newExpose.data);
      isEqualToValueOrNull(expose, 'type', newExpose.type); // http expose might not have type - will needto fix it in backend

      for (const [propKey, prop] of Object.entries(
        (device['properties'] as any) || {}
      )) {
        console.log(
          'comparing device property:',
          propKey,
          prop
        );

        var newDeviceProperty = (newDevice.properties as any)[propKey];
        isEqualToValueAndNotNull(
          device['properties'],
          propKey,
          newDeviceProperty
        );
      }

      // Expose attributes
      if ((expose as any).hasOwnProperty('attributes')) {
        if ((newExpose as any).attributes == null) {
          throw new TypeError('expose.attributes null');
        }

        for (const key of Object.keys(
          (expose as any)['attributes']
        )) {
          console.log('comparing attribute:', key);
          var newAttribute = (newExpose as any).attributes[key];
          isEqualToValueAndNotNull(
            (expose as any)['attributes'],
            key,
            newAttribute
          );
        }
      }

      // Expose presets
      if ((expose as any).hasOwnProperty('presets')) {
        if ((newExpose as any).presets == null) {
          throw new TypeError('expose.presets are null');
        }

        for (const key of Object.keys((expose as any)['presets'])) {
          console.log('comparing preset:', key);

          var newPreset = (newExpose as any).presets[key];
          isEqualToValueAndNotNull(
            (expose as any)['presets'],
            key,
            newPreset
          );
        }
      }

      // Expose properties
      if ((expose as any).hasOwnProperty('properties')) {
        if ((newExpose as any).properties == null) {
          throw new TypeError('expose.properties are null');
        }

        for (const key of Object.keys(
          (expose as any)['properties']
        )) {
          console.log('comparing property:', key);

          var newExposeProperty = (newExpose as any).properties[key];
          isEqualToValueAndNotNull(
            (expose as any)['properties'],
            key,
            newExposeProperty
          );
        }
      }
    }
  });
});

function isEqualToValueAndNotNull(
  obj: any,
  propName: string,
  value: any
) {
  expect(obj.hasOwnProperty(propName)).toBe(true);
  const srcValue = obj[propName];
  expect(value).not.toBeUndefined();
  expect(value).not.toBeNull();
  expect(value).not.toBe(null);
  expect(srcValue).toBe(value);
}

function isEqualToValueOrNull(
  obj: any,
  propName: string,
  value: any
) {
  const srcValue = obj[propName];
  expect(srcValue).toBe(value);
}
