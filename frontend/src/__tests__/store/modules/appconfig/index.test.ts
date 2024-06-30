import "jest";
import { describe, expect, test } from "@jest/globals";
import { store } from "../../../../store/index";

import { Automation, Automations } from "../../../../types/automation";
import { AppConfig } from "@/types/settings";

const mockAppconfig: AppConfig = {
  devices: {
    x01234: {
      id: "x01234",
      disabled: true,
      metricsEnabled: false,
      rateLimit: 50000,
    },
    x111111: {
      id: "x111111",
      disabled: false,
      metricsEnabled: false,
      rateLimit: 500000000,
    },
    x2222222: {
      id: "x2222222",
      disabled: false,
      metricsEnabled: true,
      rateLimit: 1000000000,
    },
  },
};

test("test appconfig gets initialized", () => {
  var result = store.getters["appconfig/initialized"]() as boolean;
  expect(result).toEqual(false);

  store.dispatch("appconfig/init", mockAppconfig);
  var result = store.getters["appconfig/initialized"]() as boolean;
  expect(result).toEqual(true);
});

test("test save device settigs saves the device settigs changes", () => {});

test("test find device settigs loads the correct device settings", () => {});

//saveDeviceConfig
