import "jest";
import { describe, expect, test } from "@jest/globals";
import { default as devicesObj } from "../../../../docs/devices.json";

test("roundtrip serializing device", () => {
  devicesObj.payload.forEach((device) => {
    var json = JSON.stringify(device);

    const verifyResult: Device = JSON.parse(json);
    var new_json = JSON.stringify(verifyResult);
    expect(new_json).toBe(json);
  });
});
