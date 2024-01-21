import "jest";
import { describe, expect, test } from "@jest/globals";
import { store_temp } from "../../../../store/index";
import { default as devices } from "../../../../../../docs/devices.json";
import { Device, Devices, DeviceUpdate } from "../../../../types/device";

test("test devices/find can load all inserted devices", () => {
  // add them to store
  devices.payload.forEach((device) => {
    var json = JSON.stringify(device);
    const newDevice: Device = JSON.parse(json);
    store_temp.commit("devices/add", newDevice);
  });

  // assert values
  devices.payload.forEach((device) => {
    var json = JSON.stringify(device);
    const newDevice: Device = JSON.parse(json);
    const result = store_temp.getters["devices/find"](newDevice.id) as Device;
    expect(result).toEqual(newDevice);
  });
});

test("test devices/list returns a list of all devices", () => {
  // add them to store
  devices.payload.forEach((device) => {
    var json = JSON.stringify(device);
    const newDevice: Device = JSON.parse(json);
    store_temp.commit("devices/add", newDevice);
  });

  // assert values
  const result = store_temp.getters["devices/list"]() as Devices;
  devices.payload.forEach((device, idx) => {
    var json = JSON.stringify(device);
    const newDevice: Device = JSON.parse(json);
    expect(result[idx]).toEqual(newDevice);
  });
  expect(result.length).toBe(devices.payload.length);
});

test("test devices/clear removes all devices from store", () => {
  // add them to store
  devices.payload.forEach((device) => {
    var json = JSON.stringify(device);
    const newDevice: Device = JSON.parse(json);
    store_temp.commit("devices/add", newDevice);
  });

  // assert values
  store_temp.commit("devices/clear");
  const result = store_temp.getters["devices/list"]() as Devices;

  expect(result.length).toBe(0);
});

test("test devices/init inserts all devices in store", () => {
  // assert values
  store_temp.dispatch("devices/init", devices.payload);
  const result = store_temp.getters["devices/list"]() as Devices;

  devices.payload.forEach((device, idx) => {
    var json = JSON.stringify(device);
    const newDevice: Device = JSON.parse(json);
    expect(result[idx]).toEqual(newDevice);
  });
  expect(result.length).toBe(devices.payload.length);
});

test("test devices/update update device properties", () => {
  // add device
  var json = JSON.stringify(devices.payload[0]);
  const newDevice: Device = JSON.parse(json);
  store_temp.commit("devices/add", newDevice);

  // update device
  var timestamp = new Date(Date.parse("11/30/2011")).toString();
  const update: DeviceUpdate = {
    id: newDevice.id,
    last_seen: timestamp,
    data: {
      temperature: 23.12,
      presure: 68,
      presence: true,
      state: "online",
    },
    properties: {
      availability: "online",
    },
  } as const;
  store_temp.commit("devices/update", update);

  // evaluate results
  const result = store_temp.getters["devices/find"](newDevice.id) as Device;

  expect(result.exposes["temperature"].data).toBe(23.12);
  expect(result.exposes["presence"].data).toBe(true);
  expect(result.properties["last_seen"].data).toBe(timestamp);

  expect(result.exposes["presure"].data).toBeNull();

  TODO;
});

// type updatePackage struct {
// 	Id         string         `json:"id"`
// 	LastSeen   string         `json:"last_seen"`
// 	Data       map[string]any `json:"data"`
// 	Properties map[string]any `json:"properties"`
// }
