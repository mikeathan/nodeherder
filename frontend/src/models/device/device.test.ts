import "jest";
import { describe, expect, test } from "@jest/globals";
import { default as devicesObj } from "../../../../docs/devices.json";

test("roundtrip serializing device", () => {
  let device = devicesObj.payload[0];
  var json = JSON.stringify(device);
  const verifyResult = JSON.parse(json) as Device;

  var json2 = JSON.stringify(verifyResult);
  expect(json2).toBe(json);
});
