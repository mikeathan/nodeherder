import "jest";
import { describe, expect, test, beforeEach } from "@jest/globals";
import { store } from "../../../../store/index";
import { default as devices } from "../../../../../../docs/devices.json";
import { Device, Devices, DeviceUpdate } from "../../../../types/device";

describe("test automation module", () => {
  beforeEach(() => {
    store.commit("devices/clear");
  });
  test("test devices/find can load all inserted devices", () => {
    // add them to store
    devices.payload.forEach((device) => {
      var json = JSON.stringify(device);
      const newDevice: Device = JSON.parse(json);
      store.commit("devices/add", newDevice);
    });

    // assert values
    devices.payload.forEach((device) => {
      var json = JSON.stringify(device);
      const newDevice: Device = JSON.parse(json);
      const result = store.getters["devices/find"](newDevice.id) as Device;
      expect(result).toEqual(newDevice);
    });
  });

  test("test devices/listAll returns a list of all devices", () => {
    // add them to store
    devices.payload.forEach((device) => {
      var json = JSON.stringify(device);
      const newDevice: Device = JSON.parse(json);
      store.commit("devices/add", newDevice);
    });

    // assert values
    const result = store.getters["devices/listAll"]() as Devices;
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
      store.commit("devices/add", newDevice);
    });

    // assert values
    store.commit("devices/clear");
    const result = store.getters["devices/listAll"]() as Devices;

    expect(result.length).toBe(0);
  });

  test("test devices/init inserts all devices in store", () => {
    // assert values
    store.dispatch("devices/init", devices.payload);
    const result = store.getters["devices/listAll"]() as Devices;

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
    store.commit("devices/add", newDevice);

    // update device
    var timestamp = new Date(Date.parse("11/30/2011")).toString();
    const update: DeviceUpdate = {
      id: newDevice.id,
      last_seen: timestamp,
      data: {
        temperature: 23.12,
        presure: 68,
        occupancy: true,
        state: "online",
      },
      properties: {
        availability: "online",
      },
    } as const;
    store.commit("devices/update", update);

    // evaluate results
    const result = store.getters["devices/find"](newDevice.id) as Device;

    expect(result.exposes["temperature"].data).toBe(23.12);
    expect(result.exposes["occupancy"].data).toBe(true);
    expect(result.exposes["presure"]).toBeUndefined();
    expect(result.properties["availability"]).toBe(
      update.properties["availability"]
    );
    expect(result.properties["last_seen"]).toBe(update.last_seen);
  });

  test("test devices/updateList - update store from a list of existing devices", () => {
    // add devices
    devices.payload.forEach((device) => {
      var json = JSON.stringify(device);
      const newDevice: Device = JSON.parse(json);
      store.commit("devices/add", newDevice);
    });

    // update 2 devices only
    var json = JSON.stringify(devices.payload[1]);
    const th01: Device = JSON.parse(json);
    th01.exposes["temperature"].data = 23.3;
    th01.exposes["humidity"].data = 65.1;
    th01.properties["availability"] = "offline";

    var json = JSON.stringify(devices.payload[2]);
    const atticLight: Device = JSON.parse(json);
    atticLight.exposes["brightness"].data = 100;
    atticLight.exposes["color_temp"].data = 467;

    // evaluate
    store.commit("devices/updateList", [th01, atticLight] as Devices);

    var result = store.getters["devices/find"](th01.id) as Device;
    expect(result.exposes["temperature"].data).toBe(23.3);
    expect(result.exposes["humidity"].data).toBe(65.1);
    expect(result.properties["availability"]).toBe("offline");

    var result = store.getters["devices/find"](atticLight.id) as Device;
    expect(result.exposes["brightness"].data).toBe(100);
    expect(result.exposes["color_temp"].data).toBe(467);
  });
});
