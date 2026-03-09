const fz = require("zigbee-herdsman-converters/converters/fromZigbee");
const tz = require("zigbee-herdsman-converters/converters/toZigbee");
const exposes = require("zigbee-herdsman-converters/lib/exposes");
const reporting = require("zigbee-herdsman-converters/lib/reporting");
const extend = require("zigbee-herdsman-converters/lib/extend");
const e = exposes.presets;
const ea = exposes.access;

const definition = {
  zigbeeModel: ["TWBulb03UK"],
  model: "TWBulb03UK",
  vendor: "Hive",
  description: "Light cool to warm white smart light bulb",
  extend: extend.light_onoff_brightness_colortemp(),
};

module.exports = definition;
