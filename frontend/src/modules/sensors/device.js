import moment from "moment";
import "moment-timezone";
import { isProxy, toRaw } from "vue";

export default class Device {
  constructor(payload) {
    this.updatePayload(payload);
  }

  startLastSeenTimer() {
    this.stoplastSeenTimer();
    this.lastSeenTimerId = setInterval(function () {
      this.lastSeen = formatLastSeen(this.payload);
    }, 1000);

    console.log(
      "setInterval " + newpayload.name + " id: " + this.lastSeenTimerId
    );
  }

  stoplastSeenTimer() {
    if (this.lastSeenTimerId != undefined) {
      clearInterval(this.lastSeenTimerId);
      console.log("clearInterval " + this.lastSeenTimerId);
    }
  }

  updatePayload(payload) {
    if (isProxy(payload)) {
      payload = toRaw(payload);
    }

    this.lastSeen = formatLastSeen(payload);
    this.linkQuality = formatLinkQuality(payload);
    this.batteryIconClass = getBatteryIcon(payload);
    this.linkQualityIconClass = getLinkQualityIcon(payload);
  }
}

export function formatLastSeen(payload) {
  if (isProxy(payload)) {
    payload = toRaw(payload);
  }

  if (payload["last_seen"] == undefined) {
    return "";
  }

  var sensorLastSeen = payload["last_seen"];

  var lastSeen = moment(sensorLastSeen);
  var diff = moment().diff(lastSeen);
  var duration = moment.duration(diff);

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

export function getBatteryIcon(payload) {
  if (isProxy(payload)) {
    payload = toRaw(payload);
  }

  if (payload["battery"] == undefined) {
    return "";
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
export function getLinkQualityIcon() {
  return "fa-signal fa-fw";
}

export function formatLinkQuality(payload) {
  if (isProxy(payload)) {
    payload = toRaw(payload);
  }
  if (payload["linkquality"] == undefined) {
    return "";
  }

  var linkQuality = payload["linkquality"];

  return linkQuality + " LQI";
}
