import 'jest';
import { describe, expect, test } from '@jest/globals';
import { default as devicesObj } from '../../../../docs/devices.json';
import { Device } from '../../types/device';

test('roundtrip serializing device', () => {
  devicesObj.payload.forEach((device) => {
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
      device['exposes']
    )) {
      console.log('comparing expose:', key);
      var newExpose = newDevice.exposes[key];
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

      for (const [key, prop] of Object.entries(
        device['properties']
      )) {
        console.log(
          'comparing device property:',
          key,
          prop
        );

        var newDeviceProperty = newDevice.properties[key];
        isEqualToValueAndNotNull(
          device['properties'],
          key,
          newDeviceProperty
        );
      }

      // Expose attributes
      if (expose.hasOwnProperty('attributes')) {
        if (newExpose.attributes == null) {
          throw new TypeError('expose.attributes null');
        }

        for (const key of Object.keys(
          expose['attributes']
        )) {
          console.log('comparing attribute:', key);
          var newAttribute = newExpose.attributes[key];
          isEqualToValueAndNotNull(
            expose['attributes'],
            key,
            newAttribute
          );
        }
      }

      // Expose presets
      if (expose.hasOwnProperty('presets')) {
        if (newExpose.presets == null) {
          throw new TypeError('expose.presets are null');
        }

        for (const key of Object.keys(expose['presets'])) {
          console.log('comparing preset:', key);

          var newPreset = newExpose.presets[key];
          isEqualToValueAndNotNull(
            expose['presets'],
            key,
            newPreset
          );
        }
      }

      // Expose properties
      if (expose.hasOwnProperty('properties')) {
        if (newExpose.properties == null) {
          throw new TypeError('expose.properties are null');
        }

        for (const key of Object.keys(
          expose['properties']
        )) {
          console.log('comparing property:', key);

          var newExposeProperty = newExpose.properties[key];
          isEqualToValueAndNotNull(
            expose['properties'],
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
