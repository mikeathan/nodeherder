import "jest";
import { describe, expect, test } from "@jest/globals";
import { default as devicesObj } from "../../../../docs/devices.json";
import { Device, Expose } from "./device"

test("roundtrip serializing device", () => {
  devicesObj.payload.forEach((device) => {
    var json = JSON.stringify(device);

    const newDevice: Device = JSON.parse(json);
    console.log("comparing device:", device["id"], "-", device["friendly_name"]);
    // Object.hasOwn(device, "id")
    isEqualToValueAndNotNull(device["id"], newDevice.id)
    isEqualToValueAndNotNull(device["friendly_name"], newDevice.friendly_name);
    isEqualToValueOrNull(device["description"], newDevice.description);
    isEqualToValueOrNull(device["power_source"], newDevice.power_source);
    isEqualToValueAndNotNull(device["connection_type"], newDevice.connection_type);

    for (const [key, expose] of Object.entries(device["exposes"])) {
      console.log("comparing expose:", key);
      var newExpose = newDevice.exposes[key]
      isEqualToValueAndNotNull(expose["name"], newExpose.name);

      isEqualToValueOrNull(expose["description"], newExpose.description);
      isEqualToValueOrNull(expose["unit"], newExpose.unit);

      isEqualToValueOrNull(expose["data"], newExpose.data);
      isEqualToValueOrNull(expose["type"], newExpose.type); // http expose might not have type - will needto fix it in backend

      for (const [key, prop] of Object.entries(device["properties"])) {
        console.log("comparing device property:", key, prop);

        var newDeviceProperty = newDevice.properties[key]
        isEqualToValueAndNotNull(prop, newDeviceProperty);
      }

      // Expose attributes
      if (expose.hasOwnProperty("attributes")) {
        if (newExpose.attributes == null) {
          throw new TypeError("expose.attributes null");
        }

        for (const key of Object.keys(expose["attributes"])) {
          console.log("comparing attribute:", key);

          var attribute = expose["attributes"][key]

          var newAttribute = newExpose.attributes[key]
          isEqualToValueAndNotNull(attribute, newAttribute);
        }
      }

      // Expose presets
      if (expose.hasOwnProperty("presets")) {
        if (newExpose.presets == null) {
          throw new TypeError("expose.presets are null");
        }

        for (const key of Object.keys(expose["presets"])) {
          console.log("comparing preset:", key);

          var preset = expose["presets"][key]

          var newPreset = newExpose.presets[key]
          isEqualToValueAndNotNull(preset, newPreset);
        }
      }

      // Expose properties
      if (expose.hasOwnProperty("properties")) {
        if (newExpose.properties == null) {
          throw new TypeError("expose.properties are null");
        }

        for (const key of Object.keys(expose["properties"])) {
          console.log("comparing property:", key);

          var exposeProperty = expose["properties"][key]

          var newExposeProperty = newExpose.properties[key]
          isEqualToValueAndNotNull(exposeProperty, newExposeProperty);
        }
      }
    }

  });
});

function isEqualToValueAndNotNull(sourceValue: any, destValue: any) {
  expect(destValue).not.toBeUndefined();
  expect(destValue).not.toBeNull();
  expect(destValue).not.toBe(null);
  expect(sourceValue).toBe(destValue);
}

function isEqualToValueOrNull(sourceValue: any, destValue: any) {
  expect(sourceValue).toBe(destValue);
}