import moment from "moment";
import "moment-timezone";
//import { format } from "timeago.js";
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
    this.linkQuality = formatLinkQuality(stats);
    this.powerSourceIconClass = getPowerSourceIcon(stats);
    this.linkQualityIconClass = getLinkQualityIcon(stats);

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

  var lastSeen = moment(sensorLastSeen);
  var diff = moment().diff(lastSeen);
  var duration = moment.duration(diff);

  // TODO:
  // format(lastSeen, i18n.language) ????
  var formatted = "just now";

  // TODO: handle days() > 0
  if (duration.hours() > 0) {
    formatted = duration.hours() + " hours ago";
  }
  if (duration.minutes() > 0) {
    formatted = duration.minutes() + " minutes ago";
  }
  if (duration.seconds() > 5) {
    formatted = duration.seconds() + " seconds ago";
  }

  return formatted;
}

function getPowerSourceIcon(payload) {
  if (payload["battery"] == undefined) {
    return "fa fa-plug";
  }

  var batteryClass = "";
  var battery = payload["battery"];
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

  /* // caption
  // var title = `${battery ? `, power_level` + ` ${battery}%` : ""}`; */
  if (!batteryClass) {
    batteryClass = "fa-question";
  }
  return batteryClass;
}
function getLinkQualityIcon() {
  return "fa-signal fa-fw";
}

function formatLinkQuality(payload) {
  if (payload["linkquality"] == undefined) {
    return "";
  }

  var linkQuality = payload["linkquality"];

  return linkQuality + " LQI";
}
