import "moment-timezone";
import { format } from "timeago.js";
import { isProxy, toRaw } from "vue";

export default class DeviceFormatter {
  constructor() {}

  stopLastSeenUpdater() {
    if (this._lastSeenTimerId != undefined) {
      clearInterval(this._lastSeenTimerId);
      //console.log("[DEBUG] clearInterval " + this._lastSeenTimerId);
    }
  }

  dispose() {
    this.stopLastSeenUpdater();
  }

  startLastSeenUpdater(payload, callback) {
    this.stopLastSeenUpdater();

    this._lastSeenTimerId = setInterval(function () {
      this.lastSeen = formatLastSeen(payload);

      //console.log("[DEBUG] timer tick " + this.lastSeen);
      callback(this.lastSeen);
    }, 1000);
  }

  formatPayload(stats, updaterCallback) {
    if (isProxy(stats)) {
      stats = toRaw(stats);
    }
    this.lastSeen = formatLastSeen(stats);

    if (isCallback(updaterCallback)) {
      this.startLastSeenUpdater(stats, updaterCallback);
    }
  }
}

function isCallback(callback) {
  return callback && typeof callback == "function";
}

function formatLastSeen(payload) {
  if (payload["last_seen"] == undefined) {
    return "";
  }

  var sensorLastSeen = payload["last_seen"];
  return format(sensorLastSeen, "en_UK");
}

export function getPowerSourceIcon(power_source, value) {
  if (power_source == "") {
    return "";
  }
  if (power_source === "mains") {
    return "fa fa-plug";
  }

  var batteryClass = "";
  var battery = value;
  if (battery >= 85) {
    batteryClass += " fa-battery-full";
  } else if (battery >= 75) {
    batteryClass += " fa-battery-three-quarters";
  } else if (battery >= 50) {
    batteryClass += " fa-battery-half";
  } else if (battery >= 25) {
    batteryClass += " fa-battery-quarter";
  } else if (battery >= 10) {
    batteryClass += ` fa-battery-empty animation-blinking`;
  } else {
    return `animation-blinking text-danger`;
  }

  if (!batteryClass) {
    batteryClass = "fa-question";
  }
  return batteryClass;
}
