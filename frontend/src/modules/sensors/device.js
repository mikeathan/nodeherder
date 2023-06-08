import moment from "moment";
import "moment-timezone";
import { isProxy, toRaw } from "vue";
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

  var result = "just now";
  // TODO: handle days() > 0
  if (duration.hours() > 0) {
    result = duration.hours() + " hours ago";
  }
  if (duration.minutes() > 0) {
    result = duration.minutes() + " minutes ago";
  }
  if (duration.seconds() > 5) {
    result = duration.seconds() + " seconds ago";
  }

  console.log(
    payload["name"] + " last_seen: " + sensorLastSeen + " formated:" + result
  );
  return result;
}

export function formatDeviceInfo(payload) {}

export function formatBattery(payload) {
  if (payload["battery"] == undefined) {
    return "";
  }

  var batteryClass;
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

function formatLinkQuality(payload) {
  if (payload["linkquality"] != undefined) {
    return "";
  }

  //<i className="fa fa-signal fa-fw" /> {linkquality} LQI
  var linkQuality = payload["linkquality"];

  return linkQuality + " LQI";
}
