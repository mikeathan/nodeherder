const sensorReadingsWhitelist = {
  temperature: "Temperature",
  pressure: "Pressure",
  humidity: "Humidity",
};

const sensorUnits = {
  temperature: "°C",
  pressure: "$hPa",
  humidity: "%",
  voltage: "mV",
  linkquality: "LQI",
};

function getUnit(sensor) {
  if (sensorUnits[sensor] == undefined) {
    return "";
  }
  return sensorUnits[sensor];
}

export function formatSensorValue(sensor, value) {
  var sensorName = sensor.charAt(0).toUpperCase() + sensor.slice(1);
  return sensorName + " " + value + getUnit(sensor);
}

export function isSensorWhitelisted(reading) {
  return sensorReadingsWhitelist[reading] != undefined;
}
